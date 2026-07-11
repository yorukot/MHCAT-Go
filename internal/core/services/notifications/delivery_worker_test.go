package notifications

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
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestDeliveryWorkerAcquiresLeaseReconcilesAndReloadsBeforeSend(t *testing.T) {
	now := time.Now().UTC()
	repo := newDeliveryRepository(deliveryFixture("*/30 * * * *"))
	messages := newAutoNotificationDeliverySideEffects()
	scheduler := newFakeDeliveryScheduler()
	leases := fakemongo.NewSchedulerLeaseStore()
	worker, err := newDeliveryWorker(
		DeliveryService{Repository: repo, Messages: messages, Channels: messages},
		leases,
		scheduler,
		"worker-a",
		2*time.Minute,
		time.Second,
		time.Hour,
		func() time.Time { return now },
		discardNotificationLogger(),
	)
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if !worker.Start(context.Background()) {
		t.Fatal("worker did not start")
	}
	waitForNotification(t, func() bool { return scheduler.count() == 1 })
	status, err := leases.Inspect(context.Background(), AutoNotificationDeliveryLeaseName, now)
	if err != nil || !status.Held || status.Owner != "worker-a" {
		t.Fatalf("lease status=%#v err=%v", status, err)
	}

	scheduler.triggerAll()
	waitForNotification(t, func() bool { return len(messages.Sent) == 1 })
	if messages.Sent[0].ChannelID != "channel-1" || messages.Sent[0].Message.Content != "hello" {
		t.Fatalf("sent = %#v", messages.Sent)
	}

	repo.delete("guild-1", "schedule-1")
	scheduler.triggerAll()
	time.Sleep(10 * time.Millisecond)
	if len(messages.Sent) != 1 {
		t.Fatalf("deleted schedule sent again: %#v", messages.Sent)
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := worker.Stop(stopCtx); err != nil {
		t.Fatalf("stop worker: %v", err)
	}
	status, err = leases.Inspect(context.Background(), AutoNotificationDeliveryLeaseName, now)
	if err != nil || status.Held {
		t.Fatalf("lease was not released: status=%#v err=%v", status, err)
	}
}

func TestDeliveryWorkerDoesNotScheduleWithoutLeaseOwnership(t *testing.T) {
	now := time.Now().UTC()
	leases := fakemongo.NewSchedulerLeaseStore()
	if _, err := leases.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{
		Name: AutoNotificationDeliveryLeaseName, Owner: "worker-a", TTL: time.Minute, Now: now,
	}); err != nil {
		t.Fatalf("seed lease: %v", err)
	}
	scheduler := newFakeDeliveryScheduler()
	worker, err := newDeliveryWorker(
		autoNotificationDeliveryService(newDeliveryRepository(deliveryFixture("*/30 * * * *"))),
		leases,
		scheduler,
		"worker-b",
		time.Minute,
		time.Second,
		time.Hour,
		func() time.Time { return now },
		discardNotificationLogger(),
	)
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if err := worker.runCycle(context.Background()); err != nil {
		t.Fatalf("run cycle: %v", err)
	}
	if scheduler.count() != 0 || worker.currentLease().Acquired {
		t.Fatalf("non-owner scheduled entries=%d lease=%#v", scheduler.count(), worker.currentLease())
	}
}

func TestDeliveryWorkerRemovesEntriesAfterLeaseLoss(t *testing.T) {
	now := time.Now().UTC()
	current := now
	leases := fakemongo.NewSchedulerLeaseStore()
	scheduler := newFakeDeliveryScheduler()
	worker, err := newDeliveryWorker(
		autoNotificationDeliveryService(newDeliveryRepository(deliveryFixture("*/30 * * * *"))),
		leases,
		scheduler,
		"worker-a",
		2*time.Minute,
		time.Second,
		time.Hour,
		func() time.Time { return current },
		discardNotificationLogger(),
	)
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if err := worker.runCycle(context.Background()); err != nil {
		t.Fatalf("first cycle: %v", err)
	}
	lease := worker.currentLease()
	if scheduler.count() != 1 || !lease.Acquired {
		t.Fatalf("entries=%d lease=%#v", scheduler.count(), lease)
	}
	if err := leases.Release(context.Background(), lease); err != nil {
		t.Fatalf("release first owner: %v", err)
	}
	if _, err := leases.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{
		Name: AutoNotificationDeliveryLeaseName, Owner: "worker-b", TTL: 2 * time.Minute, Now: now.Add(time.Second),
	}); err != nil {
		t.Fatalf("acquire second owner: %v", err)
	}
	current = now.Add(70 * time.Second)
	if err := worker.runCycle(context.Background()); !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
		t.Fatalf("expected lost lease error, got %v", err)
	}
	if scheduler.count() != 0 || worker.currentLease().Acquired {
		t.Fatalf("lost owner retained entries=%d lease=%#v", scheduler.count(), worker.currentLease())
	}
}

func TestDeliveryWorkerReconcilesCronChangesAndInvalidRows(t *testing.T) {
	now := time.Now().UTC()
	repo := newDeliveryRepository(
		deliveryFixture("*/30 * * * *"),
		domain.AutoNotificationSchedule{GuildID: "guild-1", ID: "invalid", Cron: "not a cron", ChannelID: "channel-2", Message: domain.AutoNotificationMessage{Content: "bad"}},
	)
	scheduler := newFakeDeliveryScheduler()
	worker, err := newDeliveryWorker(
		autoNotificationDeliveryService(repo),
		fakemongo.NewSchedulerLeaseStore(),
		scheduler,
		"worker-a",
		2*time.Minute,
		time.Second,
		time.Hour,
		func() time.Time { return now },
		discardNotificationLogger(),
	)
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if err := worker.runCycle(context.Background()); err != nil {
		t.Fatalf("first cycle: %v", err)
	}
	if scheduler.count() != 1 || scheduler.onlySpec() != "*/30 * * * *" {
		t.Fatalf("entries=%d spec=%q", scheduler.count(), scheduler.onlySpec())
	}
	updated := deliveryFixture("15 * * * *")
	repo.set(updated)
	if err := worker.runCycle(context.Background()); err != nil {
		t.Fatalf("second cycle: %v", err)
	}
	if scheduler.count() != 1 || scheduler.onlySpec() != "15 * * * *" || scheduler.removedCount() != 1 {
		t.Fatalf("entries=%d spec=%q removed=%d", scheduler.count(), scheduler.onlySpec(), scheduler.removedCount())
	}
}

func TestDeliveryWorkerSchedulesLegacySundaySeven(t *testing.T) {
	now := time.Now().UTC()
	repo := newDeliveryRepository(deliveryFixture("0 9 * * 5-7"))
	scheduler := newFakeDeliveryScheduler()
	worker, err := newDeliveryWorker(
		autoNotificationDeliveryService(repo),
		fakemongo.NewSchedulerLeaseStore(),
		scheduler,
		"worker-a",
		2*time.Minute,
		time.Second,
		time.Hour,
		func() time.Time { return now },
		discardNotificationLogger(),
	)
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if err := worker.runCycle(context.Background()); err != nil {
		t.Fatalf("run cycle: %v", err)
	}
	if scheduler.count() != 1 || scheduler.onlySpec() != "0 9 * * 0,5,6" {
		t.Fatalf("entries=%d spec=%q", scheduler.count(), scheduler.onlySpec())
	}
}

func TestDeliveryWorkerConstructorRejectsIncompleteOwnership(t *testing.T) {
	_, err := newDeliveryWorker(DeliveryService{}, nil, nil, "", 0, 0, 0, nil, nil)
	if !errors.Is(err, domain.ErrInvalidAutoNotificationSchedule) {
		t.Fatalf("expected invalid worker error, got %v", err)
	}
}

func deliveryFixture(spec string) domain.AutoNotificationSchedule {
	return domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      spec,
		ChannelID: "channel-1",
		Message:   domain.AutoNotificationMessage{Content: "hello"},
	}
}

func autoNotificationDeliveryService(repo ports.AutoNotificationDeliveryRepository) DeliveryService {
	sideEffects := newAutoNotificationDeliverySideEffects()
	return DeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}
}

type fakeDeliveryScheduler struct {
	mu      sync.Mutex
	next    cron.EntryID
	entries map[cron.EntryID]fakeDeliveryEntry
	removed int
	started bool
	stopped bool
}

type fakeDeliveryEntry struct {
	spec string
	fn   func()
}

func newFakeDeliveryScheduler() *fakeDeliveryScheduler {
	return &fakeDeliveryScheduler{entries: map[cron.EntryID]fakeDeliveryEntry{}}
}

func (s *fakeDeliveryScheduler) AddFunc(spec string, fn func()) (cron.EntryID, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(spec); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	s.entries[s.next] = fakeDeliveryEntry{spec: spec, fn: fn}
	return s.next, nil
}

func (s *fakeDeliveryScheduler) Remove(id cron.EntryID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[id]; ok {
		delete(s.entries, id)
		s.removed++
	}
}

func (s *fakeDeliveryScheduler) Start() {
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
}

func (s *fakeDeliveryScheduler) Stop() context.Context {
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func (s *fakeDeliveryScheduler) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func (s *fakeDeliveryScheduler) removedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.removed
}

func (s *fakeDeliveryScheduler) onlySpec() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.entries {
		return entry.spec
	}
	return ""
}

func (s *fakeDeliveryScheduler) triggerAll() {
	s.mu.Lock()
	callbacks := make([]func(), 0, len(s.entries))
	for _, entry := range s.entries {
		callbacks = append(callbacks, entry.fn)
	}
	s.mu.Unlock()
	for _, callback := range callbacks {
		callback()
	}
}

func waitForNotification(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for auto-notification worker")
}

func discardNotificationLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
