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

func TestWorkPayoutMissingMongoEnvFails(t *testing.T) {
	exitCode, _, stderr, _, _ := runWithFake(t, nil, map[string]string{}, &fakemongo.WorkPayoutRepository{}, newTestLeaseStore(true))
	if exitCode == 0 {
		t.Fatal("expected missing config to fail")
	}
	if !strings.Contains(stderr, "MHCAT_MONGODB_URI") {
		t.Fatalf("expected missing URI error, stderr=%q", stderr)
	}
}

func TestWorkPayoutDefaultsToDryRunPreviewOnly(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{PreviewResult: domain.WorkPayoutResult{EligibleJobs: 7}}
	leaseStore := newTestLeaseStore(true)
	exitCode, stdout, stderr, _, store := runWithFake(t, nil, baseEnv(), repository, leaseStore)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if repository.PreviewCalls != 1 || repository.RunCalls != 0 {
		t.Fatalf("unexpected repository calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if store.AcquireCalls != 0 || store.ReleaseCalls != 0 {
		t.Fatalf("dry-run must not acquire lease: acquire=%d release=%d", store.AcquireCalls, store.ReleaseCalls)
	}
	if !strings.Contains(stderr, "dry-run") || !strings.Contains(stdout, "eligible_jobs=7") {
		t.Fatalf("unexpected output stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestFormatWorkPayoutReportIncludesIdempotentReplays(t *testing.T) {
	var output bytes.Buffer
	report := workPayoutReport{
		Mode: "apply",
		Result: domain.WorkPayoutResult{
			ProcessedJobs:     3,
			IdempotentReplays: 2,
		},
	}
	if err := formatWorkPayoutReport(&output, report, "text"); err != nil {
		t.Fatalf("format report: %v", err)
	}
	if !strings.Contains(output.String(), "idempotent_replays=2") {
		t.Fatalf("report missing replay count: %q", output.String())
	}
}

func TestWorkPayoutApplyRequiresFeatureGate(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{}
	leaseStore := newTestLeaseStore(true)
	exitCode, _, stderr, _, store := runWithFake(t, []string{"--apply"}, baseEnv(), repository, leaseStore)
	if exitCode == 0 {
		t.Fatal("expected apply without gates to fail")
	}
	if repository.PreviewCalls != 0 || repository.RunCalls != 0 || store.AcquireCalls != 0 {
		t.Fatalf("unsafe calls: preview=%d run=%d acquire=%d", repository.PreviewCalls, repository.RunCalls, store.AcquireCalls)
	}
	if !strings.Contains(stderr, "MHCAT_JOBS_WORK_PAYOUT_ENABLED") {
		t.Fatalf("expected work payout gate error, stderr=%q", stderr)
	}
}

func TestWorkPayoutApplyRequiresSchedulerGateAndOwner(t *testing.T) {
	env := baseEnv()
	env["MHCAT_JOBS_WORK_PAYOUT_ENABLED"] = "true"
	repository := &fakemongo.WorkPayoutRepository{}
	leaseStore := newTestLeaseStore(true)
	exitCode, _, stderr, _, _ := runWithFake(t, []string{"--apply"}, env, repository, leaseStore)
	if exitCode == 0 {
		t.Fatal("expected missing scheduler gate to fail")
	}
	if !strings.Contains(stderr, "MHCAT_SCHEDULER_LEASE_ENABLED") {
		t.Fatalf("expected scheduler gate error, stderr=%q", stderr)
	}

	env["MHCAT_SCHEDULER_LEASE_ENABLED"] = "true"
	exitCode, _, stderr, _, _ = runWithFake(t, []string{"--apply"}, env, repository, leaseStore)
	if exitCode == 0 {
		t.Fatal("expected missing owner to fail")
	}
	if !strings.Contains(stderr, "MHCAT_SCHEDULER_LEASE_OWNER") {
		t.Fatalf("expected owner error, stderr=%q", stderr)
	}
}

func TestWorkPayoutApplyRunsUnderLease(t *testing.T) {
	env := applyEnv()
	repository := &fakemongo.WorkPayoutRepository{
		RunResult: domain.WorkPayoutResult{
			EligibleJobs:  2,
			ProcessedJobs: 2,
			CoinModified:  1,
			CoinUpserted:  1,
			StateModified: 2,
		},
	}
	leaseStore := newTestLeaseStore(true)
	exitCode, stdout, stderr, _, store := runWithFake(t, []string{"--apply"}, env, repository, leaseStore)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if repository.PreviewCalls != 0 || repository.RunCalls != 1 {
		t.Fatalf("unexpected repository calls: preview=%d run=%d", repository.PreviewCalls, repository.RunCalls)
	}
	if store.AcquireCalls != 1 || store.ReleaseCalls != 1 {
		t.Fatalf("expected lease acquire/release, got acquire=%d release=%d", store.AcquireCalls, store.ReleaseCalls)
	}
	if !strings.Contains(stdout, "mode=apply") || !strings.Contains(stdout, "lease_acquired=true") || !strings.Contains(stdout, "processed_jobs=2") {
		t.Fatalf("unexpected stdout=%q", stdout)
	}
}

func TestWorkPayoutApplySkipsWhenLeaseNotAcquired(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{}
	leaseStore := newTestLeaseStore(false)
	exitCode, stdout, stderr, _, store := runWithFake(t, []string{"--apply"}, applyEnv(), repository, leaseStore)
	if exitCode != 2 {
		t.Fatalf("expected contention exit 2, got %d stderr=%q", exitCode, stderr)
	}
	if repository.RunCalls != 0 || store.AcquireCalls != 1 || store.ReleaseCalls != 0 {
		t.Fatalf("unexpected calls: run=%d acquire=%d release=%d", repository.RunCalls, store.AcquireCalls, store.ReleaseCalls)
	}
	if !strings.Contains(stdout, "lease_acquired=false") {
		t.Fatalf("expected lease skip report, stdout=%q", stdout)
	}
}

func TestWorkPayoutApplyFailsOnReleaseError(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{RunResult: domain.WorkPayoutResult{ProcessedJobs: 1}}
	leaseStore := newTestLeaseStore(true)
	leaseStore.ReleaseErr = errors.New("release failed")
	exitCode, _, stderr, _, store := runWithFake(t, []string{"--apply"}, applyEnv(), repository, leaseStore)
	if exitCode == 0 {
		t.Fatal("expected release error to fail")
	}
	if repository.RunCalls != 1 || store.AcquireCalls != 1 || store.ReleaseCalls != 1 {
		t.Fatalf("unexpected calls: run=%d acquire=%d release=%d", repository.RunCalls, store.AcquireCalls, store.ReleaseCalls)
	}
	if !strings.Contains(stderr, "release work payout lease") {
		t.Fatalf("expected release error, stderr=%q", stderr)
	}
}

func TestWorkPayoutApplyAndDryRunConflict(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{}
	exitCode, _, stderr, _, _ := runWithFake(t, []string{"--apply", "--dry-run"}, applyEnv(), repository, newTestLeaseStore(true))
	if exitCode == 0 {
		t.Fatal("expected conflicting flags to fail")
	}
	if !strings.Contains(stderr, "cannot be used together") {
		t.Fatalf("expected conflict error, stderr=%q", stderr)
	}
}

func TestWorkPayoutDryRunFalseIsRejected(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{}
	exitCode, _, stderr, _, _ := runWithFake(t, []string{"--dry-run=false"}, baseEnv(), repository, newTestLeaseStore(true))
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

func TestWorkPayoutJSONOutput(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{PreviewResult: domain.WorkPayoutResult{EligibleJobs: 3}}
	exitCode, stdout, stderr, _, _ := runWithFake(t, []string{"--format", "json"}, baseEnv(), repository, newTestLeaseStore(true))
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if !strings.Contains(stdout, "\"mode\": \"dry-run\"") || !strings.Contains(stdout, "\"EligibleJobs\": 3") {
		t.Fatalf("unexpected json output=%q", stdout)
	}
}

func TestWorkPayoutDoesNotPrintRawMongoPassword(t *testing.T) {
	env := baseEnv()
	primaryPassword := strings.Join([]string{"super", "secret", "password"}, "-")
	aliasPassword := strings.Join([]string{"another", "secret", "password"}, "-")
	env["MHCAT_MONGODB_URI"] = "mongodb://user:" + primaryPassword + "@localhost:27017/mhcat"
	env["MONGOOSE_CONNECTION_STRING"] = "mongodb://user:" + aliasPassword + "@localhost:27017/mhcat"
	repository := &fakemongo.WorkPayoutRepository{}
	exitCode, stdout, stderr, _, _ := runWithFake(t, nil, env, repository, newTestLeaseStore(true))
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

func TestWorkPayoutRepositoryErrorFails(t *testing.T) {
	repository := &fakemongo.WorkPayoutRepository{PreviewErr: errors.New("boom")}
	exitCode, _, stderr, _, _ := runWithFake(t, nil, baseEnv(), repository, newTestLeaseStore(true))
	if exitCode == 0 {
		t.Fatal("expected repository error to fail")
	}
	if !strings.Contains(stderr, "work payout error") {
		t.Fatalf("expected work payout error, stderr=%q", stderr)
	}
}

func TestLegacyRoundedUnixMatchesNodeMathRoundSeconds(t *testing.T) {
	base := time.Unix(100, 0)
	if got := domain.LegacyRoundedWorkPayoutUnix(base.Add(400 * time.Millisecond)); got != 100 {
		t.Fatalf("round .4s = %d", got)
	}
	if got := domain.LegacyRoundedWorkPayoutUnix(base.Add(600 * time.Millisecond)); got != 101 {
		t.Fatalf("round .6s = %d", got)
	}
}

func runWithFake(t *testing.T, args []string, env map[string]string, repository *fakemongo.WorkPayoutRepository, leaseStore *testLeaseStore) (int, string, string, *fakemongo.WorkPayoutRepository, *testLeaseStore) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	factory := func(context.Context, config.WorkPayoutConfig) (ports.WorkPayoutRepository, ports.SchedulerLeaseStore, func(context.Context) error, error) {
		return repository, leaseStore, func(context.Context) error { return nil }, nil
	}
	exitCode := runWithFactory(context.Background(), args, mapLookup(env), &stdout, &stderr, factory)
	return exitCode, stdout.String(), stderr.String(), repository, leaseStore
}

func baseEnv() map[string]string {
	return map[string]string{
		"MHCAT_MONGODB_URI":      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE": "mhcat",
	}
}

func applyEnv() map[string]string {
	env := baseEnv()
	env["MHCAT_JOBS_WORK_PAYOUT_ENABLED"] = "true"
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

type testLeaseStore struct {
	AcquireResult domain.SchedulerLease
	AcquireErr    error
	ReleaseErr    error
	AcquireCalls  int
	ReleaseCalls  int
	LastRequest   domain.SchedulerLeaseRequest
	LastRelease   domain.SchedulerLease
}

func newTestLeaseStore(acquired bool) *testLeaseStore {
	return &testLeaseStore{
		AcquireResult: domain.SchedulerLease{
			Name:      "work-payout",
			Owner:     "worker-a",
			Fence:     9,
			Acquired:  acquired,
			ExpiresAt: time.Now().Add(time.Minute),
		},
	}
}

func (s *testLeaseStore) Inspect(ctx context.Context, name string, now time.Time) (domain.SchedulerLeaseStatus, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLeaseStatus{}, err
	}
	return domain.SchedulerLeaseStatus{Name: name, Owner: s.AcquireResult.Owner, Fence: s.AcquireResult.Fence, ExpiresAt: s.AcquireResult.ExpiresAt, Held: s.AcquireResult.Acquired}, nil
}

func (s *testLeaseStore) TryAcquire(ctx context.Context, request domain.SchedulerLeaseRequest) (domain.SchedulerLease, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	s.AcquireCalls++
	s.LastRequest = request
	if s.AcquireErr != nil {
		return domain.SchedulerLease{}, s.AcquireErr
	}
	lease := s.AcquireResult
	lease.Name = request.Name
	lease.Owner = request.Owner
	return lease, nil
}

func (s *testLeaseStore) Renew(ctx context.Context, lease domain.SchedulerLease, ttl time.Duration, now time.Time) (domain.SchedulerLease, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	return lease, nil
}

func (s *testLeaseStore) Release(ctx context.Context, lease domain.SchedulerLease) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.ReleaseCalls++
	s.LastRelease = lease
	return s.ReleaseErr
}
