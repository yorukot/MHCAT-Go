package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestEconomyResetMissingMongoEnvFails(t *testing.T) {
	exitCode, _, stderr, _ := runWithFake(t, nil, map[string]string{}, &fakemongo.DailyResetRepository{})
	if exitCode == 0 {
		t.Fatal("expected missing config to fail")
	}
	if !strings.Contains(stderr, "MHCAT_MONGODB_URI") {
		t.Fatalf("expected missing URI error, stderr=%q", stderr)
	}
}

func TestEconomyResetDefaultsToDryRunPreviewOnly(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{
		PreviewResult: domain.DailyResetResult{
			ExcludedGuilds:       2,
			CoinsMatched:         10,
			WorkGuilds:           3,
			WorkEnergyIncrements: 4,
			WorkEnergyClamps:     5,
		},
	}
	exitCode, stdout, stderr, _ := runWithFake(t, nil, baseEnv(), repository)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if repository.PreviewCalls != 1 || repository.RunCalls != 0 {
		t.Fatalf("unexpected calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if !strings.Contains(stderr, "dry-run") {
		t.Fatalf("expected dry-run notice, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "coins_matched=10") || !strings.Contains(stdout, "coins_modified=0") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
}

func TestEconomyResetApplyRequiresEnabledGate(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{}
	exitCode, _, stderr, _ := runWithFake(t, []string{"--apply"}, baseEnv(), repository)
	if exitCode == 0 {
		t.Fatal("expected apply without gate to fail")
	}
	if repository.PreviewCalls != 0 || repository.RunCalls != 0 {
		t.Fatalf("unsafe calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if !strings.Contains(stderr, "MHCAT_JOBS_DAILY_RESET_ENABLED") {
		t.Fatalf("expected enabled gate error, stderr=%q", stderr)
	}
}

func TestEconomyResetApplyRunsWhenExplicitlyEnabled(t *testing.T) {
	env := applyEnv()
	repository := &fakemongo.DailyResetRepository{
		RunResult: domain.DailyResetResult{CoinsMatched: 10, CoinsModified: 7},
	}
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--apply"}, env, repository)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if repository.PreviewCalls != 0 || repository.RunCalls != 1 {
		t.Fatalf("unexpected calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if !strings.Contains(stdout, "mode=apply") || !strings.Contains(stdout, "lease_name=daily-reset") || !strings.Contains(stdout, "lease_acquired=true") || !strings.Contains(stdout, "coins_modified=7") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
}

func TestEconomyResetApplySkipsWhenLeaseIsHeld(t *testing.T) {
	now := time.Now().UTC()
	leases := fakemongo.NewSchedulerLeaseStore()
	if _, err := leases.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "other-worker", TTL: 2 * time.Minute, Now: now}); err != nil {
		t.Fatalf("seed lease: %v", err)
	}
	repository := &fakemongo.DailyResetRepository{}
	exitCode, stdout, stderr, _ := runWithFakeLease(t, []string{"--apply"}, applyEnv(), repository, leases)
	if exitCode != 2 {
		t.Fatalf("expected held-lease exit 2, got %d stderr=%q", exitCode, stderr)
	}
	if repository.RunCalls != 0 || !strings.Contains(stdout, "lease_acquired=false") {
		t.Fatalf("run_calls=%d stdout=%q", repository.RunCalls, stdout)
	}
}

func TestEconomyResetApplyReleasesLeaseAfterRepositoryFailure(t *testing.T) {
	leases := fakemongo.NewSchedulerLeaseStore()
	repository := &fakemongo.DailyResetRepository{RunErr: errors.New("reset failed")}
	exitCode, _, stderr, _ := runWithFakeLease(t, []string{"--apply"}, applyEnv(), repository, leases)
	if exitCode == 0 || !strings.Contains(stderr, "daily reset error") {
		t.Fatalf("expected apply failure, code=%d stderr=%q", exitCode, stderr)
	}
	status, err := leases.Inspect(context.Background(), "daily-reset", time.Now().UTC())
	if err != nil || status.Held {
		t.Fatalf("lease status=%#v err=%v", status, err)
	}
}

func TestEconomyResetDryRunFalseIsRejected(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{}
	exitCode, _, stderr, _ := runWithFake(t, []string{"--dry-run=false"}, baseEnv(), repository)
	if exitCode == 0 {
		t.Fatal("expected dry-run=false to fail")
	}
	if repository.PreviewCalls != 0 || repository.RunCalls != 0 {
		t.Fatalf("unsafe calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if !strings.Contains(stderr, "--apply") {
		t.Fatalf("expected apply hint, stderr=%q", stderr)
	}
}

func TestEconomyResetApplyAndDryRunConflict(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{}
	env := applyEnv()
	exitCode, _, stderr, _ := runWithFake(t, []string{"--apply", "--dry-run"}, env, repository)
	if exitCode == 0 {
		t.Fatal("expected conflicting flags to fail")
	}
	if !strings.Contains(stderr, "cannot be used together") {
		t.Fatalf("expected conflict error, stderr=%q", stderr)
	}
}

func TestEconomyResetJSONOutputDeterministic(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{PreviewResult: domain.DailyResetResult{CoinsMatched: 1}}
	exitCode, stdout, stderr, _ := runWithFake(t, []string{"--format", "json"}, baseEnv(), repository)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "\"mode\": \"dry-run\"") || !strings.Contains(stdout, "\"CoinsMatched\": 1") {
		t.Fatalf("unexpected json output=%q", stdout)
	}
}

func TestEconomyResetDoesNotPrintRawMongoPassword(t *testing.T) {
	env := baseEnv()
	primaryPassword := strings.Join([]string{"super", "secret", "password"}, "-")
	aliasPassword := strings.Join([]string{"another", "secret", "password"}, "-")
	rawURI := "mongodb://user:" + primaryPassword + "@localhost:27017/mhcat"
	aliasURI := "mongodb://user:" + aliasPassword + "@localhost:27017/mhcat"
	env["MHCAT_MONGODB_URI"] = rawURI
	env["MONGOOSE_CONNECTION_STRING"] = aliasURI
	repository := &fakemongo.DailyResetRepository{}
	exitCode, stdout, stderr, _ := runWithFake(t, nil, env, repository)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	combined := stdout + stderr
	for _, secret := range []string{primaryPassword, aliasPassword} {
		if strings.Contains(combined, secret) {
			t.Fatalf("raw secret appeared in output: %q", combined)
		}
	}
}

func TestEconomyResetRepositoryErrorFails(t *testing.T) {
	repository := &fakemongo.DailyResetRepository{PreviewErr: errors.New("boom")}
	exitCode, _, stderr, _ := runWithFake(t, nil, baseEnv(), repository)
	if exitCode == 0 {
		t.Fatal("expected repository error to fail")
	}
	if !strings.Contains(stderr, "daily reset error") {
		t.Fatalf("expected reset error, stderr=%q", stderr)
	}
}

func runWithFake(t *testing.T, args []string, env map[string]string, repository *fakemongo.DailyResetRepository) (int, string, string, *fakemongo.DailyResetRepository) {
	return runWithFakeLease(t, args, env, repository, fakemongo.NewSchedulerLeaseStore())
}

func runWithFakeLease(t *testing.T, args []string, env map[string]string, repository *fakemongo.DailyResetRepository, leases ports.SchedulerLeaseStore) (int, string, string, *fakemongo.DailyResetRepository) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	factory := func(context.Context, config.DailyResetConfig) (ports.DailyResetRepository, ports.SchedulerLeaseStore, func(context.Context) error, error) {
		return repository, leases, func(context.Context) error { return nil }, nil
	}
	exitCode := runWithFactory(context.Background(), args, mapLookup(env), &stdout, &stderr, factory)
	return exitCode, stdout.String(), stderr.String(), repository
}

func applyEnv() map[string]string {
	env := baseEnv()
	env["MHCAT_JOBS_DAILY_RESET_ENABLED"] = "true"
	env["MHCAT_SCHEDULER_LEASE_ENABLED"] = "true"
	env["MHCAT_SCHEDULER_LEASE_OWNER"] = "worker-a"
	return env
}

func baseEnv() map[string]string {
	return map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}
}

func mapLookup(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
