package ports

import (
	"context"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type SchedulerLeaseStore interface {
	Inspect(ctx context.Context, name string, now time.Time) (domain.SchedulerLeaseStatus, error)
	TryAcquire(ctx context.Context, request domain.SchedulerLeaseRequest) (domain.SchedulerLease, error)
	Renew(ctx context.Context, lease domain.SchedulerLease, ttl time.Duration, now time.Time) (domain.SchedulerLease, error)
	Release(ctx context.Context, lease domain.SchedulerLease) error
}
