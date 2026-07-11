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
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer) int {
	cfg, err := config.LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit config error: %v\n", err)
		return 1
	}
	format, outputPath, err := applyFlags(args, &cfg, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateMongoAdmin(cfg); err != nil {
		fmt.Fprintf(stderr, "mongo audit config error: %v\n", err)
		return 1
	}
	if format != "text" && format != "json" {
		fmt.Fprintf(stderr, "mongo audit config error: format must be text or json\n")
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

	client, err := mongo.NewClient(mongo.Options{
		URI:            cfg.MongoDBURI,
		Database:       cfg.MongoDBDatabase,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit client error: %v\n", err)
		return 1
	}
	auditCtx, cancel := context.WithTimeout(ctx, cfg.AuditTimeout)
	defer cancel()
	if err := client.Connect(auditCtx); err != nil {
		fmt.Fprintf(stderr, "mongo audit connect error: %v\n", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "mongo audit disconnect error: %v\n", err)
		}
	}()

	database, err := client.Database()
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit database error: %v\n", err)
		return 1
	}
	report, err := mongo.AuditDatabase(auditCtx, database, mongo.DefaultCollectionCatalog(), mongo.AuditOptions{
		SampleLimit:           cfg.AuditSampleLimit,
		LargeDocumentBytes:    cfg.AuditLargeDocumentBytes,
		DuplicateAuditEnabled: true,
	})
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit error: %v\n", err)
		return 1
	}
	writer, closeOutput, err := outputWriter(stdout, outputPath)
	if err != nil {
		fmt.Fprintf(stderr, "mongo audit output error: %v\n", err)
		return 1
	}
	defer closeOutput()
	if err := mongo.FormatAuditReport(writer, report, format); err != nil {
		fmt.Fprintf(stderr, "mongo audit output error: %v\n", err)
		return 1
	}
	return 0
}

func applyFlags(args []string, cfg *config.MongoAdminConfig, stderr io.Writer) (string, string, error) {
	flags := flag.NewFlagSet("mhcat-mongo-audit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sampleLimit := flags.Int("sample-limit", cfg.AuditSampleLimit, "number of documents to sample per collection")
	largeDocBytes := flags.Int("large-doc-bytes", cfg.AuditLargeDocumentBytes, "large document warning threshold in bytes")
	timeout := flags.Duration("timeout", cfg.AuditTimeout, "audit timeout")
	format := flags.String("format", "text", "output format: text or json")
	output := flags.String("output", "", "optional output path")
	if err := flags.Parse(args); err != nil {
		return "", "", err
	}
	if flags.NArg() != 0 {
		return "", "", fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	cfg.AuditSampleLimit = *sampleLimit
	cfg.AuditLargeDocumentBytes = *largeDocBytes
	cfg.AuditTimeout = *timeout
	return *format, *output, nil
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
