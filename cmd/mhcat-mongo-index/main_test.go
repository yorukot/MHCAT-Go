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

func TestMongoIndexRunRejectsPositionalArgumentsBeforeConnect(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"--apply", "unexpected"}, mongoIndexLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://127.0.0.1:1",
		"MHCAT_MONGODB_DATABASE": "test",
	}), &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "unexpected positional arguments") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestMongoIndexApplyFlagsFailClosed(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "dry run false", args: []string{"--dry-run=false"}},
		{name: "apply and dry run", args: []string{"--apply", "--dry-run=true"}},
		{name: "allow unique without apply", args: []string{"--allow-unique"}},
		{name: "allow ttl without apply", args: []string{"--allow-ttl"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.MongoAdminConfig{IndexTimeout: time.Minute}
			_, err := applyFlags(test.args, &cfg, &bytes.Buffer{})
			if err == nil {
				err = config.ValidateMongoAdmin(config.MongoAdminConfig{
					LogLevel: "info", LogFormat: "text", MongoDBURI: "mongodb://localhost", MongoDBDatabase: "test",
					MongoConnectTimeout: time.Second, MongoPingTimeout: time.Second, AuditLargeDocumentBytes: 1,
					AuditTimeout: time.Second, IndexTimeout: cfg.IndexTimeout, IndexDryRun: cfg.IndexDryRun,
					IndexApply: cfg.IndexApply, IndexAllowUnique: cfg.IndexAllowUnique, IndexAllowTTL: cfg.IndexAllowTTL,
				})
			}
			if err == nil {
				t.Fatalf("args %#v unexpectedly accepted", test.args)
			}
		})
	}
}

func TestMongoIndexLoadPlanAndDuplicateAuditMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.json")
	payload := `{"indexes":[{"collection":"coins","name":"coins_lookup","keys":[{"field":"guild","order":1}]}]}`
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	plan, err := loadPlan(path)
	if err != nil || len(plan.Indexes) != 1 || plan.Indexes[0].Name != "coins_lookup" {
		t.Fatalf("plan=%#v err=%v", plan, err)
	}

	catalog := []mongo.CollectionSpec{{
		Name:           "coins",
		PlannedIndexes: []mongo.IndexSpec{{Collection: "coins", Name: "coins_unique", Unique: true}},
	}}
	clean := duplicateAuditCleanMap(catalog, mongo.AuditReport{Collections: []mongo.CollectionAudit{{Name: "coins"}}})
	if !clean["coins/coins_unique"] {
		t.Fatalf("clean map=%#v", clean)
	}
	dirty := duplicateAuditCleanMap(catalog, mongo.AuditReport{Collections: []mongo.CollectionAudit{{
		Name: "coins", DuplicateKeyRisks: []mongo.DuplicateKeyRisk{{KeyName: "guild_member"}},
	}}})
	if dirty["coins/coins_unique"] {
		t.Fatalf("dirty map=%#v", dirty)
	}
}

func TestMongoIndexRunDryRunIntegration(t *testing.T) {
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"--dry-run", "--format", "json"}, os.LookupEnv, &stdout, &stderr)
	if code != 0 || !strings.Contains(stderr.String(), "no indexes created") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	var plan mongo.IndexDiffPlan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("decode index output: %v", err)
	}
	if len(plan.Operations) == 0 {
		t.Fatal("expected default index plan operations")
	}
}

func TestMongoIndexOutputWriterWritesRequestedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "indexes.json")
	writer, closeOutput, err := outputWriter(&bytes.Buffer{}, path)
	if err != nil {
		t.Fatalf("output writer: %v", err)
	}
	if _, err := writer.Write([]byte("indexes")); err != nil {
		t.Fatalf("write indexes: %v", err)
	}
	closeOutput()
	payload, err := os.ReadFile(path)
	if err != nil || string(payload) != "indexes" {
		t.Fatalf("payload=%q err=%v", payload, err)
	}
}

func mongoIndexLookup(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
