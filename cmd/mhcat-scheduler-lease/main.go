package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
)

type leaseStoreFactory func(context.Context, config.SchedulerLeaseConfig) (ports.SchedulerLeaseStore, func(context.Context) error, error)

type cliOptions struct {
	Action string
	Name   string
	Owner  string
	Fence  int64
	Apply  bool
	Format string
}

type leaseReport struct {
	Action string                      `json:"action"`
	Status domain.SchedulerLeaseStatus `json:"status,omitempty"`
	Lease  domain.SchedulerLease       `json:"lease,omitempty"`
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer) int {
	return runWithFactory(ctx, args, lookup, stdout, stderr, defaultLeaseStoreFactory)
}

func runWithFactory(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer, factory leaseStoreFactory) int {
	cfg, err := config.LoadSchedulerLeaseRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "scheduler lease config error: %v\n", err)
		return 1
	}
	opts, err := applyFlags(args, &cfg, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "scheduler lease flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateSchedulerLease(cfg); err != nil {
		fmt.Fprintf(stderr, "scheduler lease config error: %v\n", err)
		return 1
	}
	if opts.Format != "text" && opts.Format != "json" {
		fmt.Fprintf(stderr, "scheduler lease config error: format must be text or json\n")
		return 1
	}
	if opts.Name == "" {
		fmt.Fprintf(stderr, "scheduler lease config error: --name is required\n")
		return 1
	}
	if opts.Action != "status" && (!opts.Apply || !cfg.Enabled) {
		fmt.Fprintf(stderr, "scheduler lease safety error: %s requires --apply and MHCAT_SCHEDULER_LEASE_ENABLED=true\n", opts.Action)
		return 1
	}
	if opts.Action != "status" && firstNonEmpty(opts.Owner, cfg.Owner) == "" {
		fmt.Fprintf(stderr, "scheduler lease config error: owner is required for %s\n", opts.Action)
		return 1
	}

	logger := observability.NewLogger(observability.LoggerOptions{Level: cfg.LogLevel, Format: cfg.LogFormat, Writer: stderr})
	for _, warning := range cfg.AliasWarnings {
		logger.WarnContext(ctx, warning.Message(), aliasAttrs(warning.RedactedFields())...)
	}

	leaseCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	store, cleanup, err := factory(leaseCtx, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "scheduler lease setup error: %v\n", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cleanup(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "scheduler lease cleanup error: %v\n", err)
		}
	}()

	now := time.Now().UTC()
	report, err := runAction(leaseCtx, store, cfg, opts, now)
	if err != nil {
		fmt.Fprintf(stderr, "scheduler lease error: %v\n", err)
		return 1
	}
	if err := formatLeaseReport(stdout, report, opts.Format); err != nil {
		fmt.Fprintf(stderr, "scheduler lease output error: %v\n", err)
		return 1
	}
	if opts.Action == "status" {
		fmt.Fprintln(stderr, "scheduler lease status: read-only; no Mongo writes were performed")
	}
	return 0
}

func runAction(ctx context.Context, store ports.SchedulerLeaseStore, cfg config.SchedulerLeaseConfig, opts cliOptions, now time.Time) (leaseReport, error) {
	switch opts.Action {
	case "status":
		status, err := store.Inspect(ctx, opts.Name, now)
		return leaseReport{Action: opts.Action, Status: status}, err
	case "acquire":
		owner := firstNonEmpty(opts.Owner, cfg.Owner)
		lease, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: opts.Name, Owner: owner, TTL: cfg.TTL, Now: now})
		return leaseReport{Action: opts.Action, Lease: lease}, err
	case "renew":
		owner := firstNonEmpty(opts.Owner, cfg.Owner)
		lease, err := store.Renew(ctx, domain.SchedulerLease{Name: opts.Name, Owner: owner, Fence: opts.Fence, Acquired: true, ExpiresAt: now.Add(time.Nanosecond)}, cfg.TTL, now)
		return leaseReport{Action: opts.Action, Lease: lease}, err
	case "release":
		owner := firstNonEmpty(opts.Owner, cfg.Owner)
		err := store.Release(ctx, domain.SchedulerLease{Name: opts.Name, Owner: owner, Fence: opts.Fence, Acquired: true, ExpiresAt: now.Add(time.Nanosecond)})
		return leaseReport{Action: opts.Action, Lease: domain.SchedulerLease{Name: opts.Name, Owner: owner, Fence: opts.Fence}}, err
	default:
		return leaseReport{}, fmt.Errorf("unsupported action %q", opts.Action)
	}
}

func applyFlags(args []string, cfg *config.SchedulerLeaseConfig, stderr io.Writer) (cliOptions, error) {
	flags := flag.NewFlagSet("mhcat-scheduler-lease", flag.ContinueOnError)
	flags.SetOutput(stderr)
	action := flags.String("action", "status", "action: status, acquire, renew, or release")
	name := flags.String("name", "", "lease lock name")
	owner := flags.String("owner", cfg.Owner, "lease owner")
	fence := flags.Int64("fence", 0, "lease fence token for renew/release")
	ttl := flags.Duration("ttl", cfg.TTL, "lease ttl for acquire/renew")
	timeout := flags.Duration("timeout", cfg.Timeout, "operation timeout")
	apply := flags.Bool("apply", false, "allow write action")
	format := flags.String("format", "text", "output format: text or json")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if flags.NArg() != 0 {
		return cliOptions{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	cfg.Owner = strings.TrimSpace(*owner)
	cfg.TTL = *ttl
	cfg.Timeout = *timeout
	opts := cliOptions{
		Action: strings.ToLower(strings.TrimSpace(*action)),
		Name:   strings.TrimSpace(*name),
		Owner:  cfg.Owner,
		Fence:  *fence,
		Apply:  *apply,
		Format: strings.TrimSpace(*format),
	}
	if opts.Action == "renew" || opts.Action == "release" {
		if opts.Fence <= 0 {
			return cliOptions{}, fmt.Errorf("--fence is required for %s", opts.Action)
		}
	}
	return opts, nil
}

func defaultLeaseStoreFactory(ctx context.Context, cfg config.SchedulerLeaseConfig) (ports.SchedulerLeaseStore, func(context.Context) error, error) {
	client, err := mongo.NewClient(mongo.Options{
		URI:            cfg.MongoDBURI,
		Database:       cfg.MongoDBDatabase,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
	if err != nil {
		return nil, func(context.Context) error { return nil }, err
	}
	if err := client.Connect(ctx); err != nil {
		return nil, func(context.Context) error { return nil }, err
	}
	cleanup := client.Disconnect
	database, err := client.Database()
	if err != nil {
		_ = cleanup(context.Background())
		return nil, func(context.Context) error { return nil }, err
	}
	store, err := mongo.NewSchedulerLeaseStoreFromDatabase(database)
	if err != nil {
		_ = cleanup(context.Background())
		return nil, func(context.Context) error { return nil }, err
	}
	return store, cleanup, nil
}

func formatLeaseReport(writer io.Writer, report leaseReport, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	switch report.Action {
	case "status":
		_, err := fmt.Fprintf(writer, "action=status\nname=%s\nheld=%s\nowner=%s\nfence=%d\nexpires_at=%s\n",
			report.Status.Name,
			strconv.FormatBool(report.Status.Held),
			report.Status.Owner,
			report.Status.Fence,
			formatTime(report.Status.ExpiresAt),
		)
		return err
	default:
		_, err := fmt.Fprintf(writer, "action=%s\nname=%s\nacquired=%s\nowner=%s\nfence=%d\nexpires_at=%s\n",
			report.Action,
			report.Lease.Name,
			strconv.FormatBool(report.Lease.Acquired),
			report.Lease.Owner,
			report.Lease.Fence,
			formatTime(report.Lease.ExpiresAt),
		)
		return err
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func aliasAttrs(fields map[string]string) []any {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return attrs
}
