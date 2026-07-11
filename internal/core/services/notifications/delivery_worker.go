package notifications

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	AutoNotificationDeliveryLeaseName        = "auto-notification-delivery"
	AutoNotificationDeliveryLocationName     = "Asia/Taipei"
	DefaultAutoNotificationReconcileInterval = 30 * time.Second
)

type deliveryCronScheduler interface {
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	Remove(id cron.EntryID)
	Start()
	Stop() context.Context
}

type scheduledAutoNotification struct {
	entryID cron.EntryID
	spec    string
}

type DeliveryWorker struct {
	service          DeliveryService
	leases           ports.SchedulerLeaseStore
	scheduler        deliveryCronScheduler
	owner            string
	leaseTTL         time.Duration
	operationTimeout time.Duration
	interval         time.Duration
	now              func() time.Time
	logger           *slog.Logger

	mu      sync.Mutex
	lease   domain.SchedulerLease
	runCtx  context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	running bool
	entries map[string]scheduledAutoNotification
}

func NewDeliveryWorker(service DeliveryService, leases ports.SchedulerLeaseStore, owner string, leaseTTL time.Duration, operationTimeout time.Duration, logger *slog.Logger) (*DeliveryWorker, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	scheduler := cron.New(
		cron.WithParser(parser),
		cron.WithLocation(time.FixedZone(AutoNotificationDeliveryLocationName, 8*60*60)),
	)
	interval := DefaultAutoNotificationReconcileInterval
	if leaseTTL > 0 && leaseTTL/3 < interval {
		interval = leaseTTL / 3
	}
	return newDeliveryWorker(service, leases, scheduler, owner, leaseTTL, operationTimeout, interval, time.Now, logger)
}

func newDeliveryWorker(service DeliveryService, leases ports.SchedulerLeaseStore, scheduler deliveryCronScheduler, owner string, leaseTTL time.Duration, operationTimeout time.Duration, interval time.Duration, now func() time.Time, logger *slog.Logger) (*DeliveryWorker, error) {
	owner = strings.TrimSpace(owner)
	if service.Repository == nil || service.Messages == nil || service.Channels == nil || leases == nil || scheduler == nil || owner == "" || leaseTTL <= 0 || operationTimeout <= 0 || interval <= 0 || now == nil {
		return nil, domain.ErrInvalidAutoNotificationSchedule
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &DeliveryWorker{
		service:          service,
		leases:           leases,
		scheduler:        scheduler,
		owner:            owner,
		leaseTTL:         leaseTTL,
		operationTimeout: operationTimeout,
		interval:         interval,
		now:              now,
		logger:           logger,
		entries:          map[string]scheduledAutoNotification{},
	}, nil
}

func (w *DeliveryWorker) Start(ctx context.Context) bool {
	if w == nil {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return false
	}
	runCtx, cancel := context.WithCancel(ctx)
	w.runCtx = runCtx
	w.cancel = cancel
	w.done = make(chan struct{})
	w.running = true
	w.scheduler.Start()
	go w.loop(runCtx)
	return true
}

func (w *DeliveryWorker) Stop(ctx context.Context) error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	cancel := w.cancel
	done := w.done
	w.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *DeliveryWorker) loop(ctx context.Context) {
	defer w.finish()
	w.runCycleSafe(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runCycleSafe(ctx)
		}
	}
}

func (w *DeliveryWorker) runCycleSafe(ctx context.Context) {
	if err := w.runCycle(ctx); err != nil && ctx.Err() == nil {
		w.logger.WarnContext(ctx, "auto-notification delivery reconciliation failed", "error", err.Error())
	}
}

func (w *DeliveryWorker) runCycle(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := w.now().UTC()
	lease := w.currentLease()
	if !lease.Acquired || !lease.ExpiresAt.After(now) {
		w.clearLeaseAndEntries()
		operationCtx, cancel := context.WithTimeout(ctx, w.operationTimeout)
		acquired, err := w.leases.TryAcquire(operationCtx, domain.SchedulerLeaseRequest{
			Name:  AutoNotificationDeliveryLeaseName,
			Owner: w.owner,
			TTL:   w.leaseTTL,
			Now:   now,
		})
		cancel()
		if err != nil {
			return err
		}
		if !acquired.Acquired {
			return nil
		}
		if err := acquired.ValidateHeld(); err != nil {
			return err
		}
		w.setLease(acquired)
		lease = acquired
	}
	if lease.ExpiresAt.Sub(now) <= w.leaseTTL/2 {
		operationCtx, cancel := context.WithTimeout(ctx, w.operationTimeout)
		renewed, err := w.leases.Renew(operationCtx, lease, w.leaseTTL, now)
		cancel()
		if err != nil {
			w.clearLeaseAndEntries()
			return err
		}
		if err := renewed.ValidateHeld(); err != nil {
			w.clearLeaseAndEntries()
			return err
		}
		w.setLease(renewed)
	}
	return w.reconcile(ctx)
}

func (w *DeliveryWorker) reconcile(ctx context.Context) error {
	operationCtx, cancel := context.WithTimeout(ctx, w.operationTimeout)
	schedules, err := w.service.List(operationCtx)
	cancel()
	if err != nil {
		return err
	}
	type desiredSchedule struct {
		guildID string
		id      string
		spec    string
	}
	desired := make(map[string]desiredSchedule, len(schedules))
	for _, schedule := range schedules {
		schedule = schedule.Normalized()
		if err := domain.ValidateAutoNotificationDelivery(schedule); err != nil {
			w.logger.WarnContext(ctx, "skip invalid auto-notification delivery", "guild_id", schedule.GuildID, "schedule_id", schedule.ID)
			continue
		}
		key := autoNotificationDeliveryKey(schedule.GuildID, schedule.ID)
		if _, exists := desired[key]; exists {
			w.logger.WarnContext(ctx, "skip duplicate auto-notification delivery", "guild_id", schedule.GuildID, "schedule_id", schedule.ID)
			continue
		}
		desired[key] = desiredSchedule{guildID: schedule.GuildID, id: schedule.ID, spec: NormalizeLegacyAutoNotificationCron(schedule.Cron)}
	}
	for key, current := range w.entries {
		next, exists := desired[key]
		if exists && next.spec == current.spec {
			delete(desired, key)
			continue
		}
		w.scheduler.Remove(current.entryID)
		delete(w.entries, key)
	}
	for key, schedule := range desired {
		guildID := schedule.guildID
		id := schedule.id
		entryID, err := w.scheduler.AddFunc(schedule.spec, func() {
			w.runDeliverySafe(guildID, id)
		})
		if err != nil {
			w.logger.WarnContext(ctx, "skip auto-notification delivery with invalid cron", "guild_id", guildID, "schedule_id", id, "cron", schedule.spec, "error", err.Error())
			continue
		}
		w.entries[key] = scheduledAutoNotification{entryID: entryID, spec: schedule.spec}
	}
	return ctx.Err()
}

func (w *DeliveryWorker) runDeliverySafe(guildID string, id string) {
	ctx, cancel, ok := w.deliveryContext()
	if !ok {
		return
	}
	defer cancel()
	if err := w.service.Deliver(ctx, guildID, id); err != nil && !errors.Is(err, ports.ErrAutoNotificationScheduleMissing) && ctx.Err() == nil {
		w.logger.WarnContext(ctx, "auto-notification delivery failed", "guild_id", guildID, "schedule_id", id, "error", err.Error())
	}
}

func (w *DeliveryWorker) deliveryContext() (context.Context, context.CancelFunc, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.now().UTC()
	if !w.lease.Acquired || !w.lease.ExpiresAt.After(now) || w.runCtx == nil || w.runCtx.Err() != nil {
		return nil, nil, false
	}
	deadline := now.Add(w.operationTimeout)
	if w.lease.ExpiresAt.Before(deadline) {
		deadline = w.lease.ExpiresAt
	}
	ctx, cancel := context.WithDeadline(w.runCtx, deadline)
	return ctx, cancel, true
}

func (w *DeliveryWorker) finish() {
	w.clearEntries()
	stopCtx := w.scheduler.Stop()
	select {
	case <-stopCtx.Done():
	case <-time.After(w.operationTimeout):
	}
	lease := w.takeLease()
	if lease.Acquired {
		releaseCtx, cancel := context.WithTimeout(context.Background(), w.operationTimeout)
		if err := w.leases.Release(releaseCtx, lease); err != nil && !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
			w.logger.WarnContext(releaseCtx, "auto-notification delivery lease release failed", "error", err.Error())
		}
		cancel()
	}
	w.mu.Lock()
	w.running = false
	w.cancel = nil
	w.runCtx = nil
	close(w.done)
	w.mu.Unlock()
}

func (w *DeliveryWorker) currentLease() domain.SchedulerLease {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lease
}

func (w *DeliveryWorker) setLease(lease domain.SchedulerLease) {
	w.mu.Lock()
	w.lease = lease
	w.mu.Unlock()
}

func (w *DeliveryWorker) takeLease() domain.SchedulerLease {
	w.mu.Lock()
	defer w.mu.Unlock()
	lease := w.lease
	w.lease = domain.SchedulerLease{}
	return lease
}

func (w *DeliveryWorker) clearLeaseAndEntries() {
	w.takeLease()
	w.clearEntries()
}

func (w *DeliveryWorker) clearEntries() {
	for key, entry := range w.entries {
		w.scheduler.Remove(entry.entryID)
		delete(w.entries, key)
	}
}

func autoNotificationDeliveryKey(guildID string, id string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(id)
}
