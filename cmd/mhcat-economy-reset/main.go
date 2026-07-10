package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/repositories"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
)

type resetRepositoryFactory func(context.Context, config.DailyResetConfig) (ports.DailyResetRepository, ports.SchedulerLeaseStore, func(context.Context) error, error)

type outputReport struct {
	Mode          string                  `json:"mode"`
	LeaseName     string                  `json:"lease_name,omitempty"`
	LeaseAcquired bool                    `json:"lease_acquired,omitempty"`
	LeaseFence    int64                   `json:"lease_fence,omitempty"`
	Result        domain.DailyResetResult `json:"result"`
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer) int {
	return runWithFactory(ctx, args, lookup, stdout, stderr, defaultRepositoryFactory)
}

func runWithFactory(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer, factory resetRepositoryFactory) int {
	cfg, err := config.LoadDailyResetRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "daily reset config error: %v\n", err)
		return 1
	}
	format, err := applyFlags(args, &cfg, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "daily reset flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateDailyReset(cfg); err != nil {
		fmt.Fprintf(stderr, "daily reset config error: %v\n", err)
		return 1
	}
	if format != "text" && format != "json" {
		fmt.Fprintf(stderr, "daily reset config error: format must be text or json\n")
		return 1
	}

	logger := observability.NewLogger(observability.LoggerOptions{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Writer: stderr,
	})
	for _, warning := range cfg.AliasWarnings {
		logger.WarnContext(ctx, warning.Message(), aliasAttrs(warning.RedactedFields())...)
	}

	resetCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	repository, leaseStore, cleanup, err := factory(resetCtx, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "daily reset setup error: %v\n", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cleanup(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "daily reset cleanup error: %v\n", err)
		}
	}()

	report, err := runDailyReset(resetCtx, repository, leaseStore, cfg, time.Now().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "daily reset error: %v\n", err)
		return 1
	}
	if cfg.DryRun {
		fmt.Fprintln(stderr, "daily reset dry-run: no Mongo writes were performed")
	}
	if err := formatReport(stdout, report, format); err != nil {
		fmt.Fprintf(stderr, "daily reset output error: %v\n", err)
		return 1
	}
	if !cfg.DryRun && !report.LeaseAcquired {
		return 2
	}
	return 0
}

func runDailyReset(ctx context.Context, repository ports.DailyResetRepository, leaseStore ports.SchedulerLeaseStore, cfg config.DailyResetConfig, now time.Time) (report outputReport, runErr error) {
	if cfg.DryRun {
		result, err := repository.PreviewDailyReset(ctx)
		return outputReport{Mode: "dry-run", Result: result}, err
	}
	leaseCtx, cancel := context.WithTimeout(ctx, cfg.SchedulerLeaseTimeout)
	lease, err := leaseStore.TryAcquire(leaseCtx, domain.SchedulerLeaseRequest{
		Name:  coreeconomy.DailyResetSchedulerLeaseName,
		Owner: cfg.SchedulerLeaseOwner,
		TTL:   cfg.SchedulerLeaseTTL,
		Now:   now,
	})
	cancel()
	if err != nil {
		return outputReport{}, err
	}
	report = outputReport{
		Mode:          "apply",
		LeaseName:     lease.Name,
		LeaseAcquired: lease.Acquired,
		LeaseFence:    lease.Fence,
	}
	if !lease.Acquired {
		return report, nil
	}
	if err := lease.ValidateHeld(); err != nil {
		return report, err
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), cfg.SchedulerLeaseTimeout)
		defer cancel()
		if err := leaseStore.Release(releaseCtx, lease); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("release daily reset lease: %w", err))
		}
	}()
	report.Result, runErr = repository.RunDailyReset(ctx)
	return report, runErr
}

func applyFlags(args []string, cfg *config.DailyResetConfig, stderr io.Writer) (string, error) {
	flags := flag.NewFlagSet("mhcat-economy-reset", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dryRun := flags.Bool("dry-run", true, "preview reset impact without writing Mongo")
	apply := flags.Bool("apply", false, "apply the reset writes; requires MHCAT_JOBS_DAILY_RESET_ENABLED=true")
	timeout := flags.Duration("timeout", cfg.Timeout, "reset timeout")
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
	return *format, nil
}

func defaultRepositoryFactory(ctx context.Context, cfg config.DailyResetConfig) (ports.DailyResetRepository, ports.SchedulerLeaseStore, func(context.Context) error, error) {
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
	repository, err := repositories.NewDailyResetRepositoryFromDatabase(database)
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

func formatReport(writer io.Writer, report outputReport, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	_, err := fmt.Fprintf(writer, "mode=%s\nlease_name=%s\nlease_acquired=%s\nlease_fence=%d\nexcluded_guilds=%d\ncoins_matched=%d\ncoins_modified=%d\nwork_guilds=%d\nwork_energy_increments=%d\nwork_energy_clamps=%d\n",
		report.Mode,
		report.LeaseName,
		strconv.FormatBool(report.LeaseAcquired),
		report.LeaseFence,
		report.Result.ExcludedGuilds,
		report.Result.CoinsMatched,
		report.Result.CoinsModified,
		report.Result.WorkGuilds,
		report.Result.WorkEnergyIncrements,
		report.Result.WorkEnergyClamps,
	)
	return err
}

func aliasAttrs(fields map[string]string) []any {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return attrs
}
