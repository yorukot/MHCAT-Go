package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidSchedulerLease = errors.New("invalid scheduler lease")
	ErrSchedulerLeaseNotHeld = errors.New("scheduler lease is not held")
)

type SchedulerLeaseRequest struct {
	Name  string
	Owner string
	TTL   time.Duration
	Now   time.Time
}

type SchedulerLease struct {
	Name      string
	Owner     string
	Fence     int64
	Acquired  bool
	ExpiresAt time.Time
}

type SchedulerLeaseStatus struct {
	Name      string
	Owner     string
	Fence     int64
	ExpiresAt time.Time
	Held      bool
}

func (r SchedulerLeaseRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Owner) == "" || r.TTL <= 0 || r.Now.IsZero() {
		return ErrInvalidSchedulerLease
	}
	return nil
}

func (l SchedulerLease) ValidateHeld() error {
	if strings.TrimSpace(l.Name) == "" || strings.TrimSpace(l.Owner) == "" || l.Fence <= 0 || l.ExpiresAt.IsZero() || !l.Acquired {
		return ErrSchedulerLeaseNotHeld
	}
	return nil
}
