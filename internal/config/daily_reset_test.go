package config

import (
	"errors"
	"strings"
	"testing"
	"time"
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
	if cfg.SchedulerLeaseGate || cfg.SchedulerLeaseTTL != DefaultSchedulerLeaseTTL || cfg.SchedulerLeaseTimeout != DefaultSchedulerLeaseTimeout {
		t.Fatalf("lease config = %#v", cfg)
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
		"MHCAT_SCHEDULER_LEASE_ENABLED":     "true",
		"MHCAT_SCHEDULER_LEASE_OWNER":       "worker-a",
		"MHCAT_SCHEDULER_LEASE_TTL":         "3m",
		"MHCAT_SCHEDULER_LEASE_TIMEOUT":     "10s",
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
	if !cfg.SchedulerLeaseGate || cfg.SchedulerLeaseOwner != "worker-a" || cfg.SchedulerLeaseTTL != 3*time.Minute {
		t.Fatalf("lease config = %#v", cfg)
	}
}

func TestDailyResetApplyRequiresLeaseOwnership(t *testing.T) {
	base := map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_DAILY_RESET_ENABLED": "true",
		"MHCAT_JOBS_DAILY_RESET_DRY_RUN": "false",
	}
	tests := []struct {
		name    string
		values  map[string]string
		wantKey string
	}{
		{name: "lease gate", values: map[string]string{}, wantKey: "MHCAT_SCHEDULER_LEASE_ENABLED"},
		{name: "owner", values: map[string]string{"MHCAT_SCHEDULER_LEASE_ENABLED": "true"}, wantKey: "MHCAT_SCHEDULER_LEASE_OWNER"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := make(map[string]string, len(base)+len(test.values))
			for key, value := range base {
				values[key] = value
			}
			for key, value := range test.values {
				values[key] = value
			}
			_, err := LoadDailyResetWithLookup(mapLookup(values))
			if !errors.Is(err, ErrInvalidDailyResetConfig) || !strings.Contains(err.Error(), test.wantKey) {
				t.Fatalf("expected %s error, got %v", test.wantKey, err)
			}
		})
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
