package config

import (
	"errors"
	"testing"
	"time"
)

func TestMongoAdminAuditDefaultsSafe(t *testing.T) {
	cfg, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load mongo admin config: %v", err)
	}
	if cfg.AuditSampleLimit != DefaultMongoAuditSampleLimit {
		t.Fatalf("audit sample limit = %d", cfg.AuditSampleLimit)
	}
	if cfg.AuditLargeDocumentBytes != DefaultMongoAuditLargeDocBytes {
		t.Fatalf("large doc bytes = %d", cfg.AuditLargeDocumentBytes)
	}
	if cfg.AuditTimeout != DefaultMongoAuditTimeout {
		t.Fatalf("audit timeout = %s", cfg.AuditTimeout)
	}
}

func TestMongoAdminIndexDefaultsDryRun(t *testing.T) {
	cfg, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load mongo admin config: %v", err)
	}
	if !cfg.IndexDryRun {
		t.Fatal("index dry-run should default true")
	}
	if cfg.IndexApply {
		t.Fatal("index apply should default false")
	}
}

func TestMongoAdminIndexApplyRequiresExplicitNonDryRun(t *testing.T) {
	_, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":       "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE":  "mhcat",
		"MHCAT_MONGO_INDEX_APPLY": "true",
	}))
	if !errors.Is(err, ErrInvalidMongoAdminConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestMongoAdminUniqueAndTTLDisabledByDefault(t *testing.T) {
	cfg, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load mongo admin config: %v", err)
	}
	if cfg.IndexAllowUnique {
		t.Fatal("allow unique should default false")
	}
	if cfg.IndexAllowTTL {
		t.Fatal("allow ttl should default false")
	}
}

func TestMongoAdminInvalidDurationFails(t *testing.T) {
	_, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":         "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE":    "mhcat",
		"MHCAT_MONGO_AUDIT_TIMEOUT": "not-a-duration",
	}))
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}

func TestMongoAdminDurationOverrides(t *testing.T) {
	cfg, err := LoadMongoAdminWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":         "mongodb://localhost:27017",
		"MHCAT_MONGODB_DATABASE":    "mhcat",
		"MHCAT_MONGO_AUDIT_TIMEOUT": "2m",
		"MHCAT_MONGO_INDEX_TIMEOUT": "3m",
	}))
	if err != nil {
		t.Fatalf("load mongo admin config: %v", err)
	}
	if cfg.AuditTimeout != 2*time.Minute {
		t.Fatalf("audit timeout = %s", cfg.AuditTimeout)
	}
	if cfg.IndexTimeout != 3*time.Minute {
		t.Fatalf("index timeout = %s", cfg.IndexTimeout)
	}
}
