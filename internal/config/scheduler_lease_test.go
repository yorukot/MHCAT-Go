package config

import (
	"testing"
	"time"
)

func TestSchedulerLeaseDefaultsSafe(t *testing.T) {
	cfg, err := LoadSchedulerLeaseWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}))
	if err != nil {
		t.Fatalf("load scheduler lease config: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("scheduler lease writes must be disabled by default")
	}
	if cfg.Owner != "" {
		t.Fatalf("owner = %q", cfg.Owner)
	}
	if cfg.TTL != 2*time.Minute {
		t.Fatalf("ttl = %v", cfg.TTL)
	}
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("timeout = %v", cfg.Timeout)
	}
}

func TestSchedulerLeaseEnabledWithoutOwnerCanLoadForReadOnlyStatus(t *testing.T) {
	cfg, err := LoadSchedulerLeaseWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":             "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":        "mhcat",
		"MHCAT_SCHEDULER_LEASE_ENABLED": "true",
		"MHCAT_SCHEDULER_LEASE_TIMEOUT": "1s",
		"MHCAT_SCHEDULER_LEASE_TTL":     "1m",
	}))
	if err != nil {
		t.Fatalf("load scheduler lease config: %v", err)
	}
	if !cfg.Enabled || cfg.Owner != "" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestSchedulerLeaseEnabledParses(t *testing.T) {
	cfg, err := LoadSchedulerLeaseWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":             "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":        "mhcat",
		"MHCAT_SCHEDULER_LEASE_ENABLED": "true",
		"MHCAT_SCHEDULER_LEASE_OWNER":   "worker-a",
		"MHCAT_SCHEDULER_LEASE_TTL":     "3m",
		"MHCAT_SCHEDULER_LEASE_TIMEOUT": "5s",
	}))
	if err != nil {
		t.Fatalf("load scheduler lease config: %v", err)
	}
	if !cfg.Enabled || cfg.Owner != "worker-a" || cfg.TTL != 3*time.Minute || cfg.Timeout != 5*time.Second {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestSchedulerLeaseInvalidDurationFails(t *testing.T) {
	_, err := LoadSchedulerLeaseWithLookup(mapLookup(map[string]string{
		"MHCAT_MONGODB_URI":             "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":        "mhcat",
		"MHCAT_SCHEDULER_LEASE_TIMEOUT": "nope",
	}))
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}
