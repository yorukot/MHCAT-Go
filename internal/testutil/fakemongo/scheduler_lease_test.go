package fakemongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestSchedulerLeaseStoreAcquireBlocksOtherOwnerUntilExpired(t *testing.T) {
	store := NewSchedulerLeaseStore()
	now := time.Unix(100, 0).UTC()
	first, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: now})
	if err != nil {
		t.Fatalf("acquire first: %v", err)
	}
	if !first.Acquired || first.Fence != 1 {
		t.Fatalf("first lease = %#v", first)
	}
	second, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-b", TTL: time.Minute, Now: now.Add(30 * time.Second)})
	if err != nil {
		t.Fatalf("acquire second: %v", err)
	}
	if second.Acquired {
		t.Fatalf("second owner should not acquire active lease: %#v", second)
	}
	third, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-b", TTL: time.Minute, Now: now.Add(61 * time.Second)})
	if err != nil {
		t.Fatalf("acquire after expiry: %v", err)
	}
	if !third.Acquired || third.Owner != "worker-b" || third.Fence != 2 {
		t.Fatalf("third lease = %#v", third)
	}
	status, err := store.Inspect(context.Background(), "daily-reset", now.Add(62*time.Second))
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !status.Held || status.Owner != "worker-b" || status.Fence != 2 {
		t.Fatalf("status = %#v", status)
	}
}

func TestSchedulerLeaseStoreRenewAndRelease(t *testing.T) {
	store := NewSchedulerLeaseStore()
	now := time.Unix(100, 0).UTC()
	lease, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "work-payout", Owner: "worker-a", TTL: time.Minute, Now: now})
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	renewed, err := store.Renew(context.Background(), lease, 2*time.Minute, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if !renewed.ExpiresAt.Equal(now.Add(150 * time.Second)) {
		t.Fatalf("renewed expiry = %v", renewed.ExpiresAt)
	}
	if err := store.Release(context.Background(), renewed); err != nil {
		t.Fatalf("release: %v", err)
	}
	if err := store.Release(context.Background(), renewed); !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
		t.Fatalf("expected not held after release, got %v", err)
	}
	next, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "work-payout", Owner: "worker-b", TTL: time.Minute, Now: now.Add(45 * time.Second)})
	if err != nil {
		t.Fatalf("reacquire after release: %v", err)
	}
	if !next.Acquired || next.Fence != 2 {
		t.Fatalf("expected monotonic fence after release, got %#v", next)
	}
}

func TestSchedulerLeaseStoreRenewExpiredFails(t *testing.T) {
	store := NewSchedulerLeaseStore()
	now := time.Unix(100, 0).UTC()
	lease, err := store.TryAcquire(context.Background(), domain.SchedulerLeaseRequest{Name: "work-payout", Owner: "worker-a", TTL: time.Minute, Now: now})
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if _, err := store.Renew(context.Background(), lease, time.Minute, now.Add(2*time.Minute)); !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
		t.Fatalf("expected expired lease not held, got %v", err)
	}
}

func TestSchedulerLeaseStoreContextCancellation(t *testing.T) {
	store := NewSchedulerLeaseStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: time.Unix(100, 0)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
