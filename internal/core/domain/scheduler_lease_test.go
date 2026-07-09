package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSchedulerLeaseRequestValidate(t *testing.T) {
	valid := SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: time.Unix(100, 0)}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request: %v", err)
	}
	cases := []SchedulerLeaseRequest{
		{Name: "", Owner: "worker-a", TTL: time.Minute, Now: time.Unix(100, 0)},
		{Name: "daily-reset", Owner: "", TTL: time.Minute, Now: time.Unix(100, 0)},
		{Name: "daily-reset", Owner: "worker-a", TTL: 0, Now: time.Unix(100, 0)},
		{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute},
	}
	for _, tc := range cases {
		if err := tc.Validate(); !errors.Is(err, ErrInvalidSchedulerLease) {
			t.Fatalf("expected invalid request error for %#v, got %v", tc, err)
		}
	}
}

func TestSchedulerLeaseValidateHeld(t *testing.T) {
	valid := SchedulerLease{Name: "daily-reset", Owner: "worker-a", Fence: 1, Acquired: true, ExpiresAt: time.Unix(160, 0)}
	if err := valid.ValidateHeld(); err != nil {
		t.Fatalf("valid held lease: %v", err)
	}
	cases := []SchedulerLease{
		{Name: "", Owner: "worker-a", Fence: 1, Acquired: true, ExpiresAt: time.Unix(160, 0)},
		{Name: "daily-reset", Owner: "", Fence: 1, Acquired: true, ExpiresAt: time.Unix(160, 0)},
		{Name: "daily-reset", Owner: "worker-a", Fence: 0, Acquired: true, ExpiresAt: time.Unix(160, 0)},
		{Name: "daily-reset", Owner: "worker-a", Fence: 1, Acquired: false, ExpiresAt: time.Unix(160, 0)},
		{Name: "daily-reset", Owner: "worker-a", Fence: 1, Acquired: true},
	}
	for _, tc := range cases {
		if err := tc.ValidateHeld(); !errors.Is(err, ErrSchedulerLeaseNotHeld) {
			t.Fatalf("expected not held error for %#v, got %v", tc, err)
		}
	}
}
