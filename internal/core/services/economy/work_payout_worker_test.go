package economy

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestNewWorkPayoutWorkerBuildsProductionScheduler(t *testing.T) {
	worker, err := NewWorkPayoutWorker(
		&workPayoutWorkerRepository{},
		fakemongo.NewSchedulerLeaseStore(),
		"work-payout",
		"worker-a",
		2*time.Minute,
		10*time.Second,
		time.Minute,
		nil,
	)
	if err != nil || worker == nil {
		t.Fatalf("new production worker: worker=%#v err=%v", worker, err)
	}
}

func TestWorkPayoutWorkerSchedulesRoundedRunUnderConfiguredLease(t *testing.T) {
	now := time.Unix(100, 600*time.Millisecond.Nanoseconds()).UTC()
	repository := &workPayoutWorkerRepository{result: domain.WorkPayoutResult{ProcessedJobs: 2, IdempotentReplays: 1}}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeWorkPayoutScheduler()
	worker, err := newWorkPayoutWorker(repository, leases, scheduler, "custom-work-payout", "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardWorkPayoutLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if scheduler.spec != WorkPayoutSchedulerCronSpec {
		t.Fatalf("cron spec = %q", scheduler.spec)
	}
	if !worker.Start(context.Background()) {
		t.Fatal("worker did not start")
	}
	scheduler.triggerAll()
	waitForWorkPayout(t, func() bool { return repository.runCount() == 1 })
	scheduler.wait()
	if got := repository.lastNowUnix(); got != 101 {
		t.Fatalf("rounded payout timestamp = %d", got)
	}
	status, err := leases.Inspect(context.Background(), "custom-work-payout", now)
	if err != nil {
		t.Fatalf("inspect lease: %v", err)
	}
	if status.Held {
		t.Fatalf("lease was not released: %#v", status)
	}
	stopWorkPayoutWorker(t, worker)
}

func TestWorkPayoutWorkerSkipsWhenAnotherOwnerHoldsLease(t *testing.T) {
	now := time.Now().UTC()
	leases := fakemongo.NewSchedulerLeaseStore()
	if _, err := leases.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "work-payout", Owner: "worker-a", TTL: 2 * time.Minute, Now: now}); err != nil {
		t.Fatalf("seed lease: %v", err)
	}
	repository := &workPayoutWorkerRepository{}
	scheduler := newFakeWorkPayoutScheduler()
	worker, err := newWorkPayoutWorker(repository, leases, scheduler, "work-payout", "worker-b", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardWorkPayoutLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	scheduler.wait()
	if repository.runCount() != 0 {
		t.Fatalf("non-owner ran payout %d times", repository.runCount())
	}
	stopWorkPayoutWorker(t, worker)
}

func TestWorkPayoutWorkerReleasesLeaseAfterRepositoryFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &workPayoutWorkerRepository{err: errors.New("payout failed")}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeWorkPayoutScheduler()
	worker, err := newWorkPayoutWorker(repository, leases, scheduler, "work-payout", "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardWorkPayoutLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	scheduler.wait()
	status, err := leases.Inspect(context.Background(), "work-payout", now)
	if err != nil || status.Held {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	stopWorkPayoutWorker(t, worker)
}

func TestWorkPayoutWorkerRejectsOverlappingLocalRun(t *testing.T) {
	now := time.Now().UTC()
	repository := &workPayoutWorkerRepository{block: true, started: make(chan struct{})}
	scheduler := newFakeWorkPayoutScheduler()
	worker, err := newWorkPayoutWorker(repository, fakemongo.NewSchedulerLeaseStore(), scheduler, "work-payout", "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardWorkPayoutLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	select {
	case <-repository.started:
	case <-time.After(time.Second):
		t.Fatal("first payout did not start")
	}
	if _, _, err := worker.run(); !errors.Is(err, errWorkPayoutRunInProgress) {
		t.Fatalf("expected local overlap rejection, got %v", err)
	}
	if repository.runCount() != 1 {
		t.Fatalf("overlap reached repository: runs=%d", repository.runCount())
	}
	stopWorkPayoutWorker(t, worker)
}

func TestWorkPayoutWorkerShutdownCancelsRunAndReleasesLease(t *testing.T) {
	now := time.Now().UTC()
	repository := &workPayoutWorkerRepository{block: true, started: make(chan struct{})}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeWorkPayoutScheduler()
	worker, err := newWorkPayoutWorker(repository, leases, scheduler, "work-payout", "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardWorkPayoutLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	select {
	case <-repository.started:
	case <-time.After(time.Second):
		t.Fatal("payout did not start")
	}
	stopWorkPayoutWorker(t, worker)
	status, err := leases.Inspect(context.Background(), "work-payout", now)
	if err != nil || status.Held {
		t.Fatalf("status=%#v err=%v", status, err)
	}
}

func TestWorkPayoutWorkerRejectsUnsafeConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		leaseName     string
		owner         string
		leaseTTL      time.Duration
		leaseTimeout  time.Duration
		payoutTimeout time.Duration
	}{
		{name: "lease name", owner: "worker-a", leaseTTL: 2 * time.Minute, leaseTimeout: 10 * time.Second, payoutTimeout: time.Minute},
		{name: "owner", leaseName: "work-payout", leaseTTL: 2 * time.Minute, leaseTimeout: 10 * time.Second, payoutTimeout: time.Minute},
		{name: "lease window", leaseName: "work-payout", owner: "worker-a", leaseTTL: 70 * time.Second, leaseTimeout: 10 * time.Second, payoutTimeout: time.Minute},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newWorkPayoutWorker(&workPayoutWorkerRepository{}, fakemongo.NewSchedulerLeaseStore(), newFakeWorkPayoutScheduler(), test.leaseName, test.owner, test.leaseTTL, test.leaseTimeout, test.payoutTimeout, time.Now, discardWorkPayoutLogger())
			if !errors.Is(err, ErrInvalidWorkPayoutWorker) {
				t.Fatalf("expected invalid worker, got %v", err)
			}
		})
	}
}

func TestLegacyRoundedWorkPayoutUnix(t *testing.T) {
	if got := domain.LegacyRoundedWorkPayoutUnix(time.Unix(100, 499*time.Millisecond.Nanoseconds())); got != 100 {
		t.Fatalf("round down = %d", got)
	}
	if got := domain.LegacyRoundedWorkPayoutUnix(time.Unix(100, 500*time.Millisecond.Nanoseconds())); got != 101 {
		t.Fatalf("round half up = %d", got)
	}
}

type workPayoutWorkerRepository struct {
	mu       sync.Mutex
	runs     int
	nowUnix  int64
	result   domain.WorkPayoutResult
	err      error
	block    bool
	started  chan struct{}
	startOne sync.Once
}

func (r *workPayoutWorkerRepository) PreviewWorkPayout(context.Context, int64) (domain.WorkPayoutResult, error) {
	return domain.WorkPayoutResult{}, nil
}

func (r *workPayoutWorkerRepository) RunWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error) {
	r.mu.Lock()
	r.runs++
	r.nowUnix = nowUnix
	block := r.block
	started := r.started
	result := r.result
	err := r.err
	r.mu.Unlock()
	if started != nil {
		r.startOne.Do(func() { close(started) })
	}
	if block {
		<-ctx.Done()
		return domain.WorkPayoutResult{}, ctx.Err()
	}
	return result, err
}

func (r *workPayoutWorkerRepository) runCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runs
}

func (r *workPayoutWorkerRepository) lastNowUnix() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nowUnix
}

type fakeWorkPayoutScheduler struct {
	mu      sync.Mutex
	next    cron.EntryID
	spec    string
	entries map[cron.EntryID]func()
	wg      sync.WaitGroup
	started bool
	stopped bool
}

func newFakeWorkPayoutScheduler() *fakeWorkPayoutScheduler {
	return &fakeWorkPayoutScheduler{entries: map[cron.EntryID]func(){}}
}

func (s *fakeWorkPayoutScheduler) AddFunc(spec string, fn func()) (cron.EntryID, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(spec); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	s.spec = spec
	s.entries[s.next] = fn
	return s.next, nil
}

func (s *fakeWorkPayoutScheduler) Start() {
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
}

func (s *fakeWorkPayoutScheduler) Stop() context.Context {
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s.wg.Wait()
		cancel()
	}()
	return ctx
}

func (s *fakeWorkPayoutScheduler) triggerAll() {
	s.mu.Lock()
	callbacks := make([]func(), 0, len(s.entries))
	for _, callback := range s.entries {
		callbacks = append(callbacks, callback)
	}
	s.mu.Unlock()
	for _, callback := range callbacks {
		s.wg.Add(1)
		go func(fn func()) {
			defer s.wg.Done()
			fn()
		}(callback)
	}
}

func (s *fakeWorkPayoutScheduler) wait() {
	s.wg.Wait()
}

func stopWorkPayoutWorker(t *testing.T, worker *WorkPayoutWorker) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := worker.Stop(ctx); err != nil {
		t.Fatalf("stop worker: %v", err)
	}
}

func waitForWorkPayout(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for work payout worker")
}

func discardWorkPayoutLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
