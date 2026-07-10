package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/repositories"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
)

type workPayoutFactory func(context.Context, config.WorkPayoutConfig) (ports.WorkPayoutRepository, ports.SchedulerLeaseStore, func(context.Context) error, error)

type workPayoutReport struct {
	Mode          string                  `json:"mode"`
	LeaseName     string                  `json:"lease_name,omitempty"`
	LeaseAcquired bool                    `json:"lease_acquired,omitempty"`
	LeaseFence    int64                   `json:"lease_fence,omitempty"`
	Result        domain.WorkPayoutResult `json:"result"`
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer) int {
	return runWithFactory(ctx, args, lookup, stdout, stderr, defaultWorkPayoutFactory)
}

func runWithFactory(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer, factory workPayoutFactory) int {
	cfg, err := config.LoadWorkPayoutRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "work payout config error: %v\n", err)
		return 1
	}
	format, err := applyFlags(args, &cfg, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "work payout flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateWorkPayout(cfg); err != nil {
		fmt.Fprintf(stderr, "work payout config error: %v\n", err)
		return 1
	}
	if format != "text" && format != "json" {
		fmt.Fprintf(stderr, "work payout config error: format must be text or json\n")
		return 1
	}

	logger := observability.NewLogger(observability.LoggerOptions{Level: cfg.LogLevel, Format: cfg.LogFormat, Writer: stderr})
	for _, warning := range cfg.AliasWarnings {
		logger.WarnContext(ctx, warning.Message(), aliasAttrs(warning.RedactedFields())...)
	}

	operationCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	repository, leaseStore, cleanup, err := factory(operationCtx, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "work payout setup error: %v\n", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cleanup(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "work payout cleanup error: %v\n", err)
		}
	}()

	now := time.Now().UTC()
	nowUnix := legacyRoundedUnix(now)
	report, err := runWorkPayout(operationCtx, repository, leaseStore, cfg, now, nowUnix)
	if err != nil {
		fmt.Fprintf(stderr, "work payout error: %v\n", err)
		return 1
	}
	if cfg.DryRun {
		fmt.Fprintln(stderr, "work payout dry-run: no Mongo writes were performed")
	}
	if err := formatWorkPayoutReport(stdout, report, format); err != nil {
		fmt.Fprintf(stderr, "work payout output error: %v\n", err)
		return 1
	}
	if !cfg.DryRun && !report.LeaseAcquired {
		return 2
	}
	return 0
}

func runWorkPayout(ctx context.Context, repository ports.WorkPayoutRepository, leaseStore ports.SchedulerLeaseStore, cfg config.WorkPayoutConfig, now time.Time, nowUnix int64) (workPayoutReport, error) {
	if cfg.DryRun {
		result, err := repository.PreviewWorkPayout(ctx, nowUnix)
		return workPayoutReport{Mode: "dry-run", Result: result}, err
	}
	lease, err := leaseStore.TryAcquire(ctx, domain.SchedulerLeaseRequest{
		Name:  cfg.LeaseName,
		Owner: cfg.SchedulerLeaseOwner,
		TTL:   cfg.SchedulerLeaseTTL,
		Now:   now,
	})
	if err != nil {
		return workPayoutReport{}, err
	}
	report := workPayoutReport{
		Mode:          "apply",
		LeaseName:     lease.Name,
		LeaseAcquired: lease.Acquired,
		LeaseFence:    lease.Fence,
	}
	if !lease.Acquired {
		return report, nil
	}
	result, err := repository.RunWorkPayout(ctx, nowUnix)
	report.Result = result
	releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	releaseErr := leaseStore.Release(releaseCtx, lease)
	if err != nil {
		return report, err
	}
	if releaseErr != nil {
		return report, fmt.Errorf("release work payout lease: %w", releaseErr)
	}
	return report, nil
}

func applyFlags(args []string, cfg *config.WorkPayoutConfig, stderr io.Writer) (string, error) {
	flags := flag.NewFlagSet("mhcat-work-payout", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dryRun := flags.Bool("dry-run", true, "preview work payouts without writing Mongo")
	apply := flags.Bool("apply", false, "apply due work payouts; requires explicit env gates and scheduler lease ownership")
	timeout := flags.Duration("timeout", cfg.Timeout, "work payout timeout")
	leaseName := flags.String("lease-name", cfg.LeaseName, "scheduler lease name")
	owner := flags.String("owner", cfg.SchedulerLeaseOwner, "scheduler lease owner")
	ttl := flags.Duration("lease-ttl", cfg.SchedulerLeaseTTL, "scheduler lease ttl")
	format := flags.String("format", "text", "output format: text or json")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	dryRunExplicit := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "dry-run" {
			dryRunExplicit = true
		}
	})
	if *apply && dryRunExplicit {
		return "", fmt.Errorf("--apply and --dry-run cannot be used together")
	}
	cfg.Timeout = *timeout
	cfg.LeaseName = strings.TrimSpace(*leaseName)
	cfg.SchedulerLeaseOwner = strings.TrimSpace(*owner)
	cfg.SchedulerLeaseTTL = *ttl
	if *apply {
		cfg.DryRun = false
	} else {
		if dryRunExplicit && !*dryRun {
			return "", fmt.Errorf("use --apply to run writes; --dry-run=false is not supported")
		}
		cfg.DryRun = *dryRun
	}
	return strings.TrimSpace(*format), nil
}

func defaultWorkPayoutFactory(ctx context.Context, cfg config.WorkPayoutConfig) (ports.WorkPayoutRepository, ports.SchedulerLeaseStore, func(context.Context) error, error) {
	client, err := mongo.NewClient(mongo.Options{
		URI:            cfg.MongoDBURI,
		Database:       cfg.MongoDBDatabase,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
	if err != nil {
		return nil, nil, func(context.Context) error { return nil }, err
	}
	if err := client.Connect(ctx); err != nil {
		return nil, nil, func(context.Context) error { return nil }, err
	}
	cleanup := client.Disconnect
	database, err := client.Database()
	if err != nil {
		_ = cleanup(context.Background())
		return nil, nil, func(context.Context) error { return nil }, err
	}
	repository, err := repositories.NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		_ = cleanup(context.Background())
		return nil, nil, func(context.Context) error { return nil }, err
	}
	leaseStore, err := mongo.NewSchedulerLeaseStoreFromDatabase(database)
	if err != nil {
		_ = cleanup(context.Background())
		return nil, nil, func(context.Context) error { return nil }, err
	}
	return repository, leaseStore, cleanup, nil
}

func formatWorkPayoutReport(writer io.Writer, report workPayoutReport, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	_, err := fmt.Fprintf(writer, "mode=%s\nlease_name=%s\nlease_acquired=%s\nlease_fence=%d\neligible_jobs=%d\nprocessed_jobs=%d\ncoin_matched=%d\ncoin_modified=%d\ncoin_upserted=%d\nstate_matched=%d\nstate_modified=%d\nskipped_invalid_jobs=%d\n",
		report.Mode,
		report.LeaseName,
		strconv.FormatBool(report.LeaseAcquired),
		report.LeaseFence,
		report.Result.EligibleJobs,
		report.Result.ProcessedJobs,
		report.Result.CoinMatched,
		report.Result.CoinModified,
		report.Result.CoinUpserted,
		report.Result.StateMatched,
		report.Result.StateModified,
		report.Result.SkippedInvalidJobs,
	)
	return err
}

func legacyRoundedUnix(now time.Time) int64 {
	return int64(math.Round(float64(now.UnixNano()) / float64(time.Second)))
}

func aliasAttrs(fields map[string]string) []any {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return attrs
}
