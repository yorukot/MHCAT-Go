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
	WorkPayoutSchedulerCronSpec     = "* * * * *"
	WorkPayoutSchedulerLocationName = "Asia/Taipei"
)

var ErrInvalidWorkPayoutWorker = errors.New("invalid work payout worker")

var errWorkPayoutRunInProgress = errors.New("work payout run already in progress")

type workPayoutCronScheduler interface {
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	Start()
	Stop() context.Context
}

type WorkPayoutWorker struct {
	repository    ports.WorkPayoutRepository
	leases        ports.SchedulerLeaseStore
	scheduler     workPayoutCronScheduler
	leaseName     string
	owner         string
	leaseTTL      time.Duration
	leaseTimeout  time.Duration
	payoutTimeout time.Duration
	now           func() time.Time
	logger        *slog.Logger

	mu        sync.Mutex
	runCtx    context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	running   bool
	executing bool
}

func NewWorkPayoutWorker(repository ports.WorkPayoutRepository, leases ports.SchedulerLeaseStore, leaseName string, owner string, leaseTTL time.Duration, leaseTimeout time.Duration, payoutTimeout time.Duration, logger *slog.Logger) (*WorkPayoutWorker, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	scheduler := cron.New(
		cron.WithParser(parser),
		cron.WithLocation(time.FixedZone(WorkPayoutSchedulerLocationName, 8*60*60)),
	)
	return newWorkPayoutWorker(repository, leases, scheduler, leaseName, owner, leaseTTL, leaseTimeout, payoutTimeout, time.Now, logger)
}

func newWorkPayoutWorker(repository ports.WorkPayoutRepository, leases ports.SchedulerLeaseStore, scheduler workPayoutCronScheduler, leaseName string, owner string, leaseTTL time.Duration, leaseTimeout time.Duration, payoutTimeout time.Duration, now func() time.Time, logger *slog.Logger) (*WorkPayoutWorker, error) {
	leaseName = strings.TrimSpace(leaseName)
	owner = strings.TrimSpace(owner)
	if repository == nil || leases == nil || scheduler == nil || leaseName == "" || owner == "" || leaseTTL <= 0 || leaseTimeout <= 0 || payoutTimeout <= 0 || leaseTTL <= payoutTimeout || leaseTTL-payoutTimeout <= leaseTimeout || now == nil {
		return nil, ErrInvalidWorkPayoutWorker
	}
	if logger == nil {
		logger = slog.Default()
	}
	worker := &WorkPayoutWorker{
		repository:    repository,
		leases:        leases,
		scheduler:     scheduler,
		leaseName:     leaseName,
		owner:         owner,
		leaseTTL:      leaseTTL,
		leaseTimeout:  leaseTimeout,
		payoutTimeout: payoutTimeout,
		now:           now,
		logger:        logger,
	}
	if _, err := scheduler.AddFunc(WorkPayoutSchedulerCronSpec, worker.runSafe); err != nil {
		return nil, fmt.Errorf("%w: schedule payout: %v", ErrInvalidWorkPayoutWorker, err)
	}
	return worker, nil
}

func (w *WorkPayoutWorker) Start(ctx context.Context) bool {
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

func (w *WorkPayoutWorker) Stop(ctx context.Context) error {
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

func (w *WorkPayoutWorker) waitForStop(ctx context.Context, done chan struct{}) {
	<-ctx.Done()
	<-w.scheduler.Stop().Done()
	w.mu.Lock()
	if w.done == done {
		w.running = false
		w.executing = false
		w.runCtx = nil
		w.cancel = nil
		close(done)
	}
	w.mu.Unlock()
}

func (w *WorkPayoutWorker) runSafe() {
	result, acquired, err := w.run()
	if errors.Is(err, errWorkPayoutRunInProgress) {
		w.logger.Info("work payout scheduler skipped; previous local run is still active", "lease", w.leaseName)
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		w.logger.Warn("work payout scheduler run failed", "lease", w.leaseName, "error", err.Error())
		return
	}
	if !acquired {
		w.logger.Info("work payout scheduler skipped; lease held by another owner", "lease", w.leaseName)
		return
	}
	w.logger.Info(
		"work payout scheduler run completed",
		"lease", w.leaseName,
		"eligible_jobs", result.EligibleJobs,
		"processed_jobs", result.ProcessedJobs,
		"idempotent_replays", result.IdempotentReplays,
		"coin_modified", result.CoinModified,
		"coin_upserted", result.CoinUpserted,
		"state_modified", result.StateModified,
		"skipped_invalid_jobs", result.SkippedInvalidJobs,
	)
}

func (w *WorkPayoutWorker) run() (result domain.WorkPayoutResult, acquired bool, runErr error) {
	runCtx, err := w.beginExecution()
	if err != nil {
		return domain.WorkPayoutResult{}, false, err
	}
	defer w.finishExecution()

	now := w.now().UTC()
	leaseCtx, cancel := context.WithTimeout(runCtx, w.leaseTimeout)
	lease, err := w.leases.TryAcquire(leaseCtx, domain.SchedulerLeaseRequest{
		Name:  w.leaseName,
		Owner: w.owner,
		TTL:   w.leaseTTL,
		Now:   now,
	})
	cancel()
	if err != nil {
		return domain.WorkPayoutResult{}, false, err
	}
	if !lease.Acquired {
		return domain.WorkPayoutResult{}, false, nil
	}
	if err := lease.ValidateHeld(); err != nil {
		return domain.WorkPayoutResult{}, false, err
	}
	acquired = true
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), w.leaseTimeout)
		defer cancel()
		if err := w.leases.Release(releaseCtx, lease); err != nil && !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
			runErr = errors.Join(runErr, fmt.Errorf("release work payout lease: %w", err))
		}
	}()

	payoutCtx, cancel := context.WithTimeout(runCtx, w.payoutTimeout)
	if deadline, ok := payoutCtx.Deadline(); !ok || lease.ExpiresAt.Before(deadline) {
		cancel()
		payoutCtx, cancel = context.WithDeadline(runCtx, lease.ExpiresAt)
	}
	defer cancel()
	result, runErr = w.repository.RunWorkPayout(payoutCtx, domain.LegacyRoundedWorkPayoutUnix(now))
	return result, acquired, runErr
}

func (w *WorkPayoutWorker) beginExecution() (context.Context, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running || w.runCtx == nil || w.runCtx.Err() != nil {
		return nil, context.Canceled
	}
	if w.executing {
		return nil, errWorkPayoutRunInProgress
	}
	w.executing = true
	return w.runCtx, nil
}

func (w *WorkPayoutWorker) finishExecution() {
	w.mu.Lock()
	w.executing = false
	w.mu.Unlock()
}
