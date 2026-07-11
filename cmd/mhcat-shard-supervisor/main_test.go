package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLoadSupervisorConfig(t *testing.T) {
	cfg, err := loadSupervisorConfig(mapLookup(map[string]string{
		"MHCAT_BOT_PATH":                    "/bot",
		"MHCAT_DISCORD_SHARD_COUNT":         "16",
		"MHCAT_DISCORD_SHARD_SPAWN_DELAY":   "5s",
		"MHCAT_DISCORD_SHARD_RESTART_DELAY": "3s",
		"MHCAT_DISCORD_SHARD_STOP_TIMEOUT":  "20s",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.BotPath != "/bot" || cfg.ShardCount != 16 || cfg.SpawnDelay != 5*time.Second || cfg.RestartDelay != 3*time.Second || cfg.StopTimeout != 20*time.Second {
		t.Fatalf("config = %#v", cfg)
	}
}

func TestLoadSupervisorConfigRejectsInvalidValues(t *testing.T) {
	for _, env := range []map[string]string{
		{"MHCAT_DISCORD_SHARD_COUNT": "0"},
		{"MHCAT_DISCORD_SHARD_COUNT": "nope"},
		{"MHCAT_DISCORD_SHARD_SPAWN_DELAY": "-1s"},
		{"MHCAT_DISCORD_SHARD_STOP_TIMEOUT": "0s"},
		{"MHCAT_BOT_PATH": " "},
	} {
		if _, err := loadSupervisorConfig(mapLookup(env)); err == nil {
			t.Fatalf("expected error for %#v", env)
		}
	}
}

func TestShardEnvironmentReplacesExistingShardValues(t *testing.T) {
	got := shardEnvironment([]string{"A=1", "MHCAT_DISCORD_SHARD_ID=9", "MHCAT_DISCORD_SHARD_COUNT=10", "MHCAT_SCHEDULER_LEASE_OWNER=production"}, 3, 16)
	want := []string{"A=1", "MHCAT_DISCORD_SHARD_ID=3", "MHCAT_DISCORD_SHARD_COUNT=16", "MHCAT_SCHEDULER_LEASE_OWNER=production-shard-3"}
	if len(got) != len(want) {
		t.Fatalf("environment = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("environment[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestShardEnvironmentLeavesEmptyLeaseOwnerUnset(t *testing.T) {
	got := shardEnvironment([]string{"MHCAT_SCHEDULER_LEASE_OWNER="}, 0, 1)
	for _, value := range got {
		if strings.HasPrefix(value, "MHCAT_SCHEDULER_LEASE_OWNER=") {
			t.Fatalf("unexpected lease owner: %#v", got)
		}
	}
}

func TestSleepContextStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if sleepContext(ctx, time.Minute) {
		t.Fatal("sleep should stop on cancellation")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
