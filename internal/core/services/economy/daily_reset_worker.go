package economy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	DailyResetSchedulerLeaseName    = "daily-reset"
	DailyResetSchedulerCronSpec     = "0 0 * * *"
	DailyResetSchedulerLocationName = "Asia/Taipei"
)

var ErrInvalidDailyResetWorker = errors.New("invalid daily reset worker")

type dailyResetCronScheduler interface {
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	Start()
	Stop() context.Context
}

type DailyResetWorker struct {
	repository   ports.DailyResetRepository
	leases       ports.SchedulerLeaseStore
	scheduler    dailyResetCronScheduler
	owner        string
	leaseTTL     time.Duration
	leaseTimeout time.Duration
	resetTimeout time.Duration
	now          func() time.Time
	logger       *slog.Logger

	mu      sync.Mutex
	runCtx  context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	running bool
}

func NewDailyResetWorker(repository ports.DailyResetRepository, leases ports.SchedulerLeaseStore, owner string, leaseTTL time.Duration, leaseTimeout time.Duration, resetTimeout time.Duration, logger *slog.Logger) (*DailyResetWorker, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	scheduler := cron.New(
		cron.WithParser(parser),
		cron.WithLocation(time.FixedZone(DailyResetSchedulerLocationName, 8*60*60)),
	)
	return newDailyResetWorker(repository, leases, scheduler, owner, leaseTTL, leaseTimeout, resetTimeout, time.Now, logger)
}

func newDailyResetWorker(repository ports.DailyResetRepository, leases ports.SchedulerLeaseStore, scheduler dailyResetCronScheduler, owner string, leaseTTL time.Duration, leaseTimeout time.Duration, resetTimeout time.Duration, now func() time.Time, logger *slog.Logger) (*DailyResetWorker, error) {
	owner = strings.TrimSpace(owner)
	if repository == nil || leases == nil || scheduler == nil || owner == "" || leaseTTL <= 0 || leaseTimeout <= 0 || resetTimeout <= 0 || leaseTTL <= resetTimeout || leaseTTL-resetTimeout <= leaseTimeout || now == nil {
		return nil, ErrInvalidDailyResetWorker
	}
	if logger == nil {
		logger = slog.Default()
	}
	worker := &DailyResetWorker{
		repository:   repository,
		leases:       leases,
		scheduler:    scheduler,
		owner:        owner,
		leaseTTL:     leaseTTL,
		leaseTimeout: leaseTimeout,
		resetTimeout: resetTimeout,
		now:          now,
		logger:       logger,
	}
	if _, err := scheduler.AddFunc(DailyResetSchedulerCronSpec, worker.runSafe); err != nil {
		return nil, fmt.Errorf("%w: schedule reset: %v", ErrInvalidDailyResetWorker, err)
	}
	return worker, nil
}

func (w *DailyResetWorker) Start(ctx context.Context) bool {
	if w == nil || ctx == nil {
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
	go w.waitForStop(runCtx, w.done)
	return true
}

func (w *DailyResetWorker) Stop(ctx context.Context) error {
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

func (w *DailyResetWorker) waitForStop(ctx context.Context, done chan struct{}) {
	<-ctx.Done()
	<-w.scheduler.Stop().Done()
	w.mu.Lock()
	if w.done == done {
		w.running = false
		w.runCtx = nil
		w.cancel = nil
		close(done)
	}
	w.mu.Unlock()
}

func (w *DailyResetWorker) runSafe() {
	result, acquired, err := w.run()
	if err != nil {
		w.logger.Warn("daily reset scheduler run failed", "error", err.Error())
		return
	}
	if !acquired {
		w.logger.Info("daily reset scheduler skipped; lease held by another owner", "lease", DailyResetSchedulerLeaseName)
		return
	}
	w.logger.Info(
		"daily reset scheduler run completed",
		"coins_matched", result.CoinsMatched,
		"coins_modified", result.CoinsModified,
		"work_guilds", result.WorkGuilds,
		"work_energy_increments", result.WorkEnergyIncrements,
		"work_energy_clamps", result.WorkEnergyClamps,
	)
}

func (w *DailyResetWorker) run() (result domain.DailyResetResult, acquired bool, runErr error) {
	runCtx, ok := w.executionContext()
	if !ok {
		return domain.DailyResetResult{}, false, context.Canceled
	}
	now := w.now().UTC()
	leaseCtx, cancel := context.WithTimeout(runCtx, w.leaseTimeout)
	lease, err := w.leases.TryAcquire(leaseCtx, domain.SchedulerLeaseRequest{
		Name:  DailyResetSchedulerLeaseName,
		Owner: w.owner,
		TTL:   w.leaseTTL,
		Now:   now,
	})
	cancel()
	if err != nil {
		return domain.DailyResetResult{}, false, err
	}
	if !lease.Acquired {
		return domain.DailyResetResult{}, false, nil
	}
	if err := lease.ValidateHeld(); err != nil {
		return domain.DailyResetResult{}, false, err
	}
	acquired = true
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), w.leaseTimeout)
		defer cancel()
		if err := w.leases.Release(releaseCtx, lease); err != nil && !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
			runErr = errors.Join(runErr, fmt.Errorf("release daily reset lease: %w", err))
		}
	}()

	resetCtx, cancel := context.WithTimeout(runCtx, w.resetTimeout)
	if deadline, ok := resetCtx.Deadline(); !ok || lease.ExpiresAt.Before(deadline) {
		cancel()
		resetCtx, cancel = context.WithDeadline(runCtx, lease.ExpiresAt)
	}
	defer cancel()
	result, runErr = w.repository.RunDailyReset(resetCtx)
	return result, acquired, runErr
}

func (w *DailyResetWorker) executionContext() (context.Context, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running || w.runCtx == nil || w.runCtx.Err() != nil {
		return nil, false
	}
	return w.runCtx, true
}
