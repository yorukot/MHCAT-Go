package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSchedulerLeaseMissingMongoEnvFails(t *testing.T) {
	exitCode, _, stderr, _ := runWithFake(t, nil, map[string]string{}, fakemongo.NewSchedulerLeaseStore())
	if exitCode == 0 {
		t.Fatal("expected missing config to fail")
	}
	if !strings.Contains(stderr, "MHCAT_MONGODB_URI") {
		t.Fatalf("expected missing URI error, stderr=%q", stderr)
	}
}

func TestSchedulerLeaseStatusIsReadOnlyDefault(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--name", "daily-reset"}, baseEnv(), store)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "action=status") || !strings.Contains(stdout, "held=false") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "read-only") {
		t.Fatalf("expected read-only notice, stderr=%q", stderr)
	}
}

func TestSchedulerLeaseAcquireRequiresApplyAndGate(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	env := baseEnv()
	env["MHCAT_SCHEDULER_LEASE_OWNER"] = "worker-a"
	exitCode, _, stderr, _ := runWithFake(t, []string{"--action", "acquire", "--name", "daily-reset"}, env, store)
	if exitCode == 0 {
		t.Fatal("expected acquire without apply/gate to fail")
	}
	if !strings.Contains(stderr, "requires --apply") {
		t.Fatalf("expected safety error, stderr=%q", stderr)
	}
}

func TestSchedulerLeaseAcquireRunsWhenExplicitlyEnabled(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	env := enabledEnv()
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--action", "acquire", "--name", "daily-reset", "--apply"}, env, store)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "action=acquire") || !strings.Contains(stdout, "acquired=true") || !strings.Contains(stdout, "fence=1") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
}

func TestSchedulerLeaseAcquireRequiresOwner(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	env := baseEnv()
	env["MHCAT_SCHEDULER_LEASE_ENABLED"] = "true"
	exitCode, _, stderr, _ := runWithFake(t, []string{"--action", "acquire", "--name", "daily-reset", "--apply"}, env, store)
	if exitCode == 0 {
		t.Fatal("expected acquire without owner to fail")
	}
	if !strings.Contains(stderr, "owner") {
		t.Fatalf("expected owner error, stderr=%q", stderr)
	}
}

func TestSchedulerLeaseRenewRequiresFence(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	exitCode, _, stderr, _ := runWithFake(t, []string{"--action", "renew", "--name", "daily-reset", "--apply"}, enabledEnv(), store)
	if exitCode == 0 {
		t.Fatal("expected missing fence to fail")
	}
	if !strings.Contains(stderr, "--fence") {
		t.Fatalf("expected fence error, stderr=%q", stderr)
	}
}

func TestSchedulerLeaseJSONStatus(t *testing.T) {
	store := fakemongo.NewSchedulerLeaseStore()
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--name", "daily-reset", "--format", "json"}, baseEnv(), store)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "\"action\": \"status\"") || !strings.Contains(stdout, "\"Held\": false") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
}

func TestSchedulerLeaseDoesNotPrintRawMongoPassword(t *testing.T) {
	env := baseEnv()
	primaryPassword := strings.Join([]string{"lease", "primary", "password"}, "-")
	aliasPassword := strings.Join([]string{"lease", "alias", "password"}, "-")
	env["MHCAT_MONGODB_URI"] = "mongodb://user:" + primaryPassword + "@localhost:27017/mhcat"
	env["MONGOOSE_CONNECTION_STRING"] = "mongodb://user:" + aliasPassword + "@localhost:27017/mhcat"
	store := fakemongo.NewSchedulerLeaseStore()
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--name", "daily-reset"}, env, store)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	combined := stdout + stderr
	for _, secret := range []string{primaryPassword, aliasPassword} {
		if strings.Contains(combined, secret) {
			t.Fatalf("raw password appeared in output: %q", combined)
		}
	}
}

func runWithFake(t *testing.T, args []string, env map[string]string, store *fakemongo.SchedulerLeaseStore) (int, string, string, *fakemongo.SchedulerLeaseStore) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	factory := func(context.Context, config.SchedulerLeaseConfig) (ports.SchedulerLeaseStore, func(context.Context) error, error) {
		return store, func(context.Context) error { return nil }, nil
	}
	exitCode := runWithFactory(context.Background(), args, mapLookup(env), &stdout, &stderr, factory)
	return exitCode, stdout.String(), stderr.String(), store
}

func baseEnv() map[string]string {
	return map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}
}

func enabledEnv() map[string]string {
	env := baseEnv()
	env["MHCAT_SCHEDULER_LEASE_ENABLED"] = "true"
	env["MHCAT_SCHEDULER_LEASE_OWNER"] = "worker-a"
	return env
}

func mapLookup(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
