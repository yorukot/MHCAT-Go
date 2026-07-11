package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer) int {
	cfg, err := config.LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "mongo index config error: %v\n", err)
		return 1
	}
	cli, err := applyFlags(args, &cfg, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "mongo index flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateMongoAdmin(cfg); err != nil {
		fmt.Fprintf(stderr, "mongo index config error: %v\n", err)
		return 1
	}
	if cli.format != "text" && cli.format != "json" {
		fmt.Fprintf(stderr, "mongo index config error: format must be text or json\n")
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

	catalog := mongo.DefaultCollectionCatalog()
	plan := mongo.DefaultIndexPlan(catalog)
	if cli.planPath != "" {
		loaded, err := loadPlan(cli.planPath)
		if err != nil {
			fmt.Fprintf(stderr, "mongo index plan error: %v\n", err)
			return 1
		}
		plan = loaded
	}

	client, err := mongo.NewClient(mongo.Options{
		URI:            cfg.MongoDBURI,
		Database:       cfg.MongoDBDatabase,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
	if err != nil {
		fmt.Fprintf(stderr, "mongo index client error: %v\n", err)
		return 1
	}
	indexCtx, cancel := context.WithTimeout(ctx, cfg.IndexTimeout)
	defer cancel()
	if err := client.Connect(indexCtx); err != nil {
		fmt.Fprintf(stderr, "mongo index connect error: %v\n", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "mongo index disconnect error: %v\n", err)
		}
	}()

	database, err := client.Database()
	if err != nil {
		fmt.Fprintf(stderr, "mongo index database error: %v\n", err)
		return 1
	}
	collectionNames, err := database.ListCollectionNames(indexCtx, bson.D{})
	if err != nil {
		fmt.Fprintf(stderr, "mongo index list collections error: %v\n", mongo.MapError(err))
		return 1
	}
	liveIndexes, err := mongo.ListIndexes(indexCtx, database, collectionNames)
	if err != nil {
		fmt.Fprintf(stderr, "mongo index list indexes error: %v\n", err)
		return 1
	}
	duplicateAuditClean := map[string]bool{}
	if cfg.IndexAllowUnique {
		report, err := mongo.AuditDatabase(indexCtx, database, catalog, mongo.AuditOptions{
			SampleLimit:           0,
			LargeDocumentBytes:    cfg.AuditLargeDocumentBytes,
			DuplicateAuditEnabled: true,
		})
		if err != nil {
			fmt.Fprintf(stderr, "mongo index duplicate audit error: %v\n", err)
			return 1
		}
		duplicateAuditClean = duplicateAuditCleanMap(catalog, report)
	}
	diff, err := mongo.DiffIndexes(plan, liveIndexes, mongo.IndexDiffOptions{
		AllowUnique:         cfg.IndexAllowUnique,
		AllowTTL:            cfg.IndexAllowTTL,
		DuplicateAuditClean: duplicateAuditClean,
	})
	if err != nil {
		fmt.Fprintf(stderr, "mongo index diff error: %v\n", err)
		return 1
	}
	writer, closeOutput, err := outputWriter(stdout, cli.outputPath)
	if err != nil {
		fmt.Fprintf(stderr, "mongo index output error: %v\n", err)
		return 1
	}
	defer closeOutput()
	if err := mongo.FormatIndexDiffPlan(writer, diff, cli.format); err != nil {
		fmt.Fprintf(stderr, "mongo index output error: %v\n", err)
		return 1
	}
	if cfg.IndexDryRun {
		fmt.Fprintln(stderr, "mongo index dry-run: no indexes created")
		return 0
	}
	operations := mongo.SafeIndexApplyOperations(diff)
	if err := mongo.EnsureIndexes(indexCtx, database, plan, operations); err != nil {
		fmt.Fprintf(stderr, "mongo index apply error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "mongo index apply complete: created=%d\n", len(operations))
	return 0
}

type cliOptions struct {
	format     string
	outputPath string
	planPath   string
}

func applyFlags(args []string, cfg *config.MongoAdminConfig, stderr io.Writer) (cliOptions, error) {
	flags := flag.NewFlagSet("mhcat-mongo-index", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dryRun := flags.Bool("dry-run", true, "print index diff without creating indexes")
	apply := flags.Bool("apply", false, "create safe missing indexes")
	allowUnique := flags.Bool("allow-unique", false, "allow unique index creation when duplicate audit is clean; requires --apply")
	allowTTL := flags.Bool("allow-ttl", false, "allow ttl index creation when retention ADR exists; requires --apply")
	timeout := flags.Duration("timeout", cfg.IndexTimeout, "index operation timeout")
	format := flags.String("format", "text", "output format: text or json")
	output := flags.String("output", "", "optional output path")
	planPath := flags.String("plan", "", "optional index plan JSON path")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if flags.NArg() != 0 {
		return cliOptions{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}

	dryRunSet := flagWasSet(args, "dry-run")
	if *apply && dryRunSet && *dryRun {
		return cliOptions{}, fmt.Errorf("--apply and --dry-run=true cannot be used together")
	}
	if !*apply && dryRunSet && !*dryRun {
		return cliOptions{}, fmt.Errorf("apply mode requires explicit --apply")
	}
	cfg.IndexDryRun = true
	cfg.IndexApply = false
	if *apply {
		cfg.IndexDryRun = false
		cfg.IndexApply = true
	}
	if dryRunSet && *dryRun {
		cfg.IndexDryRun = true
		cfg.IndexApply = false
	}
	cfg.IndexAllowUnique = *allowUnique
	cfg.IndexAllowTTL = *allowTTL
	cfg.IndexTimeout = *timeout
	return cliOptions{format: strings.TrimSpace(*format), outputPath: *output, planPath: *planPath}, nil
}

func flagWasSet(args []string, name string) bool {
	prefix := "--" + name
	for _, arg := range args {
		if arg == prefix || strings.HasPrefix(arg, prefix+"=") {
			return true
		}
	}
	return false
}

func loadPlan(path string) (mongo.IndexPlan, error) {
	file, err := os.Open(path)
	if err != nil {
		return mongo.IndexPlan{}, err
	}
	defer file.Close()
	return mongo.LoadIndexPlan(file)
}

func duplicateAuditCleanMap(catalog []mongo.CollectionSpec, report mongo.AuditReport) map[string]bool {
	risksByCollection := map[string]bool{}
	for _, collection := range report.Collections {
		risksByCollection[collection.Name] = len(collection.DuplicateKeyRisks) > 0
	}
	clean := map[string]bool{}
	for _, spec := range catalog {
		for _, index := range spec.PlannedIndexes {
			if !index.Unique {
				continue
			}
			clean[spec.Name+"/"+index.Name] = !risksByCollection[spec.Name]
		}
	}
	return clean
}

func outputWriter(stdout io.Writer, path string) (io.Writer, func(), error) {
	if path == "" {
		return stdout, func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, func() {}, err
	}
	return file, func() { _ = file.Close() }, nil
}

func aliasAttrs(fields map[string]string) []any {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return attrs
}
