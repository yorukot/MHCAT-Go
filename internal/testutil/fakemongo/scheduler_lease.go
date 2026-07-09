package fakemongo

import (
	"context"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type SchedulerLeaseStore struct {
	mu     sync.Mutex
	leases map[string]domain.SchedulerLease
}

func NewSchedulerLeaseStore() *SchedulerLeaseStore {
	return &SchedulerLeaseStore{leases: map[string]domain.SchedulerLease{}}
}

func (s *SchedulerLeaseStore) Inspect(ctx context.Context, name string, now time.Time) (domain.SchedulerLeaseStatus, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLeaseStatus{}, err
	}
	if name == "" || now.IsZero() {
		return domain.SchedulerLeaseStatus{}, domain.ErrInvalidSchedulerLease
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, ok := s.leases[name]
	if !ok {
		return domain.SchedulerLeaseStatus{Name: name, Held: false}, nil
	}
	return domain.SchedulerLeaseStatus{
		Name:      lease.Name,
		Owner:     lease.Owner,
		Fence:     lease.Fence,
		ExpiresAt: lease.ExpiresAt,
		Held:      lease.Owner != "" && lease.ExpiresAt.After(now),
	}, nil
}

func (s *SchedulerLeaseStore) TryAcquire(ctx context.Context, request domain.SchedulerLeaseRequest) (domain.SchedulerLease, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	if err := request.Validate(); err != nil {
		return domain.SchedulerLease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.leases[request.Name]
	if ok && current.Owner != request.Owner && current.ExpiresAt.After(request.Now) {
		return domain.SchedulerLease{Name: request.Name, Owner: request.Owner, Acquired: false}, nil
	}
	fence := current.Fence + 1
	if fence <= 0 {
		fence = 1
	}
	lease := domain.SchedulerLease{
		Name:      request.Name,
		Owner:     request.Owner,
		Fence:     fence,
		Acquired:  true,
		ExpiresAt: request.Now.Add(request.TTL),
	}
	s.leases[request.Name] = lease
	return lease, nil
}

func (s *SchedulerLeaseStore) Renew(ctx context.Context, lease domain.SchedulerLease, ttl time.Duration, now time.Time) (domain.SchedulerLease, error) {
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	if err := lease.ValidateHeld(); err != nil {
		return domain.SchedulerLease{}, err
	}
	if ttl <= 0 || now.IsZero() {
		return domain.SchedulerLease{}, domain.ErrInvalidSchedulerLease
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.leases[lease.Name]
	if !ok || current.Owner != lease.Owner || current.Fence != lease.Fence || !current.ExpiresAt.After(now) {
		return domain.SchedulerLease{}, domain.ErrSchedulerLeaseNotHeld
	}
	current.ExpiresAt = now.Add(ttl)
	current.Acquired = true
	s.leases[lease.Name] = current
	return current, nil
}

func (s *SchedulerLeaseStore) Release(ctx context.Context, lease domain.SchedulerLease) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := lease.ValidateHeld(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.leases[lease.Name]
	if !ok || current.Owner != lease.Owner || current.Fence != lease.Fence {
		return domain.ErrSchedulerLeaseNotHeld
	}
	current.Owner = ""
	current.Acquired = false
	current.ExpiresAt = time.Time{}
	s.leases[lease.Name] = current
	return nil
}
