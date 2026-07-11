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

func TestNewDailyResetWorkerBuildsProductionScheduler(t *testing.T) {
	worker, err := NewDailyResetWorker(
		&dailyResetWorkerRepository{},
		fakemongo.NewSchedulerLeaseStore(),
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

func TestDailyResetWorkerSchedulesRunsUnderLeaseAndReleases(t *testing.T) {
	now := time.Now().UTC()
	repository := &dailyResetWorkerRepository{result: domain.DailyResetResult{CoinsModified: 2, WorkEnergyIncrements: 3}}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeDailyResetScheduler()
	worker, err := newDailyResetWorker(repository, leases, scheduler, "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardDailyResetLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if scheduler.spec != DailyResetSchedulerCronSpec {
		t.Fatalf("cron spec = %q", scheduler.spec)
	}
	if !worker.Start(context.Background()) {
		t.Fatal("worker did not start")
	}
	scheduler.triggerAll()
	waitForDailyReset(t, func() bool { return repository.runCount() == 1 })
	scheduler.wait()
	status, err := leases.Inspect(context.Background(), DailyResetSchedulerLeaseName, now)
	if err != nil {
		t.Fatalf("inspect lease: %v", err)
	}
	if status.Held {
		t.Fatalf("lease was not released: %#v", status)
	}
	stopDailyResetWorker(t, worker)
}

func TestDailyResetWorkerSkipsWhenAnotherOwnerHoldsLease(t *testing.T) {
	now := time.Now().UTC()
	leases := fakemongo.NewSchedulerLeaseStore()
	if _, err := leases.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: DailyResetSchedulerLeaseName, Owner: "worker-a", TTL: 2 * time.Minute, Now: now}); err != nil {
		t.Fatalf("seed lease: %v", err)
	}
	repository := &dailyResetWorkerRepository{}
	scheduler := newFakeDailyResetScheduler()
	worker, err := newDailyResetWorker(repository, leases, scheduler, "worker-b", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardDailyResetLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	scheduler.wait()
	if repository.runCount() != 0 {
		t.Fatalf("non-owner ran reset %d times", repository.runCount())
	}
	stopDailyResetWorker(t, worker)
}

func TestDailyResetWorkerReleasesLeaseAfterRepositoryFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &dailyResetWorkerRepository{err: errors.New("reset failed")}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeDailyResetScheduler()
	worker, err := newDailyResetWorker(repository, leases, scheduler, "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardDailyResetLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	scheduler.wait()
	status, err := leases.Inspect(context.Background(), DailyResetSchedulerLeaseName, now)
	if err != nil || status.Held {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	stopDailyResetWorker(t, worker)
}

func TestDailyResetWorkerShutdownCancelsRunAndReleasesLease(t *testing.T) {
	now := time.Now().UTC()
	repository := &dailyResetWorkerRepository{block: true, started: make(chan struct{})}
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeDailyResetScheduler()
	worker, err := newDailyResetWorker(repository, leases, scheduler, "worker-a", 2*time.Minute, 10*time.Second, time.Minute, func() time.Time { return now }, discardDailyResetLogger())
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.Start(context.Background())
	scheduler.triggerAll()
	select {
	case <-repository.started:
	case <-time.After(time.Second):
		t.Fatal("reset did not start")
	}
	stopDailyResetWorker(t, worker)
	status, err := leases.Inspect(context.Background(), DailyResetSchedulerLeaseName, now)
	if err != nil || status.Held {
		t.Fatalf("status=%#v err=%v", status, err)
	}
}

func TestDailyResetWorkerRejectsUnsafeLeaseWindow(t *testing.T) {
	_, err := newDailyResetWorker(&dailyResetWorkerRepository{}, fakemongo.NewSchedulerLeaseStore(), newFakeDailyResetScheduler(), "worker-a", 70*time.Second, 10*time.Second, time.Minute, time.Now, discardDailyResetLogger())
	if !errors.Is(err, ErrInvalidDailyResetWorker) {
		t.Fatalf("expected invalid worker, got %v", err)
	}
}

type dailyResetWorkerRepository struct {
	mu      sync.Mutex
	runs    int
	result  domain.DailyResetResult
	err     error
	block   bool
	started chan struct{}
}

func (r *dailyResetWorkerRepository) PreviewDailyReset(context.Context) (domain.DailyResetResult, error) {
	return domain.DailyResetResult{}, nil
}

func (r *dailyResetWorkerRepository) RunDailyReset(ctx context.Context) (domain.DailyResetResult, error) {
	r.mu.Lock()
	r.runs++
	block := r.block
	started := r.started
	result := r.result
	err := r.err
	r.mu.Unlock()
	if started != nil {
		select {
		case <-started:
		default:
			close(started)
		}
	}
	if block {
		<-ctx.Done()
		return domain.DailyResetResult{}, ctx.Err()
	}
	return result, err
}

func (r *dailyResetWorkerRepository) runCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runs
}

type fakeDailyResetScheduler struct {
	mu      sync.Mutex
	next    cron.EntryID
	spec    string
	entries map[cron.EntryID]func()
	wg      sync.WaitGroup
	started bool
	stopped bool
}

func newFakeDailyResetScheduler() *fakeDailyResetScheduler {
	return &fakeDailyResetScheduler{entries: map[cron.EntryID]func(){}}
}

func (s *fakeDailyResetScheduler) AddFunc(spec string, fn func()) (cron.EntryID, error) {
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

func (s *fakeDailyResetScheduler) Start() {
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
}

func (s *fakeDailyResetScheduler) Stop() context.Context {
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

func (s *fakeDailyResetScheduler) triggerAll() {
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

func (s *fakeDailyResetScheduler) wait() {
	s.wg.Wait()
}

func stopDailyResetWorker(t *testing.T, worker *DailyResetWorker) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := worker.Stop(ctx); err != nil {
		t.Fatalf("stop worker: %v", err)
	}
}

func waitForDailyReset(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for daily reset worker")
}

func discardDailyResetLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
