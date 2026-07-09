package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWorkPayoutDefaultsSafe(t *testing.T) {
	cfg, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load work payout config: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("work payout must be disabled by default")
	}
	if !cfg.DryRun {
		t.Fatal("work payout must default to dry-run")
	}
	if cfg.LeaseName != "work-payout" {
		t.Fatalf("lease name = %q", cfg.LeaseName)
	}
	if cfg.Timeout != time.Minute {
		t.Fatalf("timeout = %v", cfg.Timeout)
	}
	if cfg.SchedulerLeaseGate {
		t.Fatal("scheduler lease gate must be disabled by default")
	}
}

func TestWorkPayoutApplyRequiresFeatureGate(t *testing.T) {
	_, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_DRY_RUN": "false",
		"MHCAT_SCHEDULER_LEASE_ENABLED":  "true",
		"MHCAT_SCHEDULER_LEASE_OWNER":    "worker-a",
	}))
	if !errors.Is(err, ErrInvalidWorkPayoutConfig) || !strings.Contains(err.Error(), "MHCAT_JOBS_WORK_PAYOUT_ENABLED") {
		t.Fatalf("expected feature gate error, got %v", err)
	}
}

func TestWorkPayoutApplyRequiresSchedulerGate(t *testing.T) {
	_, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_ENABLED": "true",
		"MHCAT_JOBS_WORK_PAYOUT_DRY_RUN": "false",
		"MHCAT_SCHEDULER_LEASE_OWNER":    "worker-a",
	}))
	if !errors.Is(err, ErrInvalidWorkPayoutConfig) || !strings.Contains(err.Error(), "MHCAT_SCHEDULER_LEASE_ENABLED") {
		t.Fatalf("expected scheduler gate error, got %v", err)
	}
}

func TestWorkPayoutApplyRequiresSchedulerOwner(t *testing.T) {
	_, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_ENABLED": "true",
		"MHCAT_JOBS_WORK_PAYOUT_DRY_RUN": "false",
		"MHCAT_SCHEDULER_LEASE_ENABLED":  "true",
	}))
	if !errors.Is(err, ErrInvalidWorkPayoutConfig) || !strings.Contains(err.Error(), "MHCAT_SCHEDULER_LEASE_OWNER") {
		t.Fatalf("expected scheduler owner error, got %v", err)
	}
}

func TestWorkPayoutApplyParsesWhenFullyGated(t *testing.T) {
	cfg, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":                 "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":            "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_ENABLED":    "true",
		"MHCAT_JOBS_WORK_PAYOUT_DRY_RUN":    "false",
		"MHCAT_JOBS_WORK_PAYOUT_TIMEOUT":    "30s",
		"MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME": "work-payout-staging",
		"MHCAT_SCHEDULER_LEASE_ENABLED":     "true",
		"MHCAT_SCHEDULER_LEASE_OWNER":       "worker-a",
		"MHCAT_SCHEDULER_LEASE_TTL":         "3m",
	}))
	if err != nil {
		t.Fatalf("load work payout config: %v", err)
	}
	if !cfg.Enabled || cfg.DryRun || cfg.LeaseName != "work-payout-staging" || cfg.Timeout != 30*time.Second || cfg.SchedulerLeaseTTL != 3*time.Minute {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestWorkPayoutApplyRequiresLeaseTTLGreaterThanTimeout(t *testing.T) {
	_, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_ENABLED": "true",
		"MHCAT_JOBS_WORK_PAYOUT_DRY_RUN": "false",
		"MHCAT_JOBS_WORK_PAYOUT_TIMEOUT": "2m",
		"MHCAT_SCHEDULER_LEASE_ENABLED":  "true",
		"MHCAT_SCHEDULER_LEASE_OWNER":    "worker-a",
		"MHCAT_SCHEDULER_LEASE_TTL":      "2m",
	}))
	if !errors.Is(err, ErrInvalidWorkPayoutConfig) || !strings.Contains(err.Error(), "MHCAT_SCHEDULER_LEASE_TTL") {
		t.Fatalf("expected ttl greater than timeout error, got %v", err)
	}
}

func TestWorkPayoutInvalidDurationFails(t *testing.T) {
	_, err := LoadWorkPayoutWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_JOBS_WORK_PAYOUT_TIMEOUT": "nope",
	}))
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}
