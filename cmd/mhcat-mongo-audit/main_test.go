package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func TestMongoAuditRunRejectsPositionalArgumentsBeforeConnect(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"unexpected"}, mongoAuditLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://127.0.0.1:1",
		"MHCAT_MONGODB_DATABASE": "test",
	}), &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "unexpected positional arguments") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestMongoAuditApplyFlagsOverridesConfig(t *testing.T) {
	cfg := config.MongoAdminConfig{
		AuditSampleLimit:        20,
		AuditLargeDocumentBytes: 1024,
		AuditTimeout:            time.Minute,
	}
	format, output, err := applyFlags([]string{
		"--sample-limit", "7",
		"--large-doc-bytes", "2048",
		"--timeout", "5s",
		"--format", "json",
		"--output", "report.json",
	}, &cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("apply flags: %v", err)
	}
	if format != "json" || output != "report.json" || cfg.AuditSampleLimit != 7 || cfg.AuditLargeDocumentBytes != 2048 || cfg.AuditTimeout != 5*time.Second {
		t.Fatalf("format=%q output=%q cfg=%#v", format, output, cfg)
	}
}

func TestMongoAuditOutputWriterWritesRequestedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	writer, closeOutput, err := outputWriter(&bytes.Buffer{}, path)
	if err != nil {
		t.Fatalf("output writer: %v", err)
	}
	if _, err := writer.Write([]byte("audit")); err != nil {
		t.Fatalf("write report: %v", err)
	}
	closeOutput()
	payload, err := os.ReadFile(path)
	if err != nil || string(payload) != "audit" {
		t.Fatalf("payload=%q err=%v", payload, err)
	}
}

func TestMongoAuditOutputWriterRejectsDirectory(t *testing.T) {
	if _, _, err := outputWriter(&bytes.Buffer{}, t.TempDir()); err == nil {
		t.Fatal("expected directory output to fail")
	}
}

func TestMongoAuditRunIntegration(t *testing.T) {
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"--sample-limit", "0", "--format", "json"}, os.LookupEnv, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var report mongo.AuditReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode audit output: %v", err)
	}
	if report.Database != os.Getenv("MHCAT_MONGODB_DATABASE") {
		t.Fatalf("database=%q output=%q", report.Database, stdout.String())
	}
}

func mongoAuditLookup(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
