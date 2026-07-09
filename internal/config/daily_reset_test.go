package config

import (
	"errors"
	"strings"
	"testing"
)

func TestDailyResetDefaultsDryRunDisabled(t *testing.T) {
	cfg, err := LoadDailyResetWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load daily reset config: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("daily reset must be disabled by default")
	}
	if !cfg.DryRun {
		t.Fatal("daily reset must default to dry-run")
	}
	if cfg.Timeout != DefaultDailyResetTimeout {
		t.Fatalf("timeout = %v", cfg.Timeout)
	}
}

func TestDailyResetApplyRequiresEnabledGate(t *testing.T) {
	_, err := LoadDailyResetWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_DAILY_RESET_DRY_RUN": "false",
	}))
	if !errors.Is(err, ErrInvalidDailyResetConfig) {
		t.Fatalf("expected ErrInvalidDailyResetConfig, got %v", err)
	}
	if !strings.Contains(err.Error(), "MHCAT_JOBS_DAILY_RESET_ENABLED") {
		t.Fatalf("expected enabled gate in error, got %v", err)
	}
}

func TestDailyResetApplyParsesWhenEnabled(t *testing.T) {
	cfg, err := LoadDailyResetWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":                 "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":            "mhcat",
		"MHCAT_JOBS_DAILY_RESET_ENABLED":    "true",
		"MHCAT_JOBS_DAILY_RESET_DRY_RUN":    "false",
		"MHCAT_JOBS_DAILY_RESET_TIMEOUT":    "90s",
		"MHCAT_MONGO_CONNECT_TIMEOUT":       "2s",
		"MHCAT_MONGO_PING_TIMEOUT":          "2s",
		"MHCAT_MONGO_AUDIT_LARGE_DOC_BYTES": "1",
	}))
	if err != nil {
		t.Fatalf("load daily reset config: %v", err)
	}
	if !cfg.Enabled || cfg.DryRun {
		t.Fatalf("unexpected gate state: enabled=%v dryRun=%v", cfg.Enabled, cfg.DryRun)
	}
	if cfg.Timeout.String() != "1m30s" {
		t.Fatalf("timeout = %v", cfg.Timeout)
	}
}

func TestDailyResetInvalidTimeoutFails(t *testing.T) {
	_, err := LoadDailyResetWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_DAILY_RESET_TIMEOUT": "nope",
	}))
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
