package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	DefaultSchedulerLeaseEnabled = false
	DefaultSchedulerLeaseTTL     = 2 * time.Minute
	DefaultSchedulerLeaseTimeout = 10 * time.Second
)

var ErrInvalidSchedulerLeaseConfig = errors.New("invalid scheduler lease config")

type SchedulerLeaseConfig struct {
	MongoAdminConfig
	Enabled bool
	Owner   string
	TTL     time.Duration
	Timeout time.Duration
}

func LoadSchedulerLease() (SchedulerLeaseConfig, error) {
	return LoadSchedulerLeaseWithLookup(os.LookupEnv)
}

func LoadSchedulerLeaseWithLookup(lookup LookupFunc) (SchedulerLeaseConfig, error) {
	cfg, err := LoadSchedulerLeaseRawWithLookup(lookup)
	if err != nil {
		return SchedulerLeaseConfig{}, err
	}
	if err := ValidateSchedulerLease(cfg); err != nil {
		return SchedulerLeaseConfig{}, err
	}
	return cfg, nil
}

func LoadSchedulerLeaseRawWithLookup(lookup LookupFunc) (SchedulerLeaseConfig, error) {
	mongoCfg, err := LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		return SchedulerLeaseConfig{}, err
	}
	cfg := SchedulerLeaseConfig{
		MongoAdminConfig: mongoCfg,
		Enabled:          DefaultSchedulerLeaseEnabled,
		Owner:            getString(lookup, "MHCAT_SCHEDULER_LEASE_OWNER", ""),
		TTL:              DefaultSchedulerLeaseTTL,
		Timeout:          DefaultSchedulerLeaseTimeout,
	}
	if cfg.Enabled, err = getBool(lookup, "MHCAT_SCHEDULER_LEASE_ENABLED", DefaultSchedulerLeaseEnabled); err != nil {
		return SchedulerLeaseConfig{}, err
	}
	if cfg.TTL, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TTL", DefaultSchedulerLeaseTTL); err != nil {
		return SchedulerLeaseConfig{}, err
	}
	if cfg.Timeout, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TIMEOUT", DefaultSchedulerLeaseTimeout); err != nil {
		return SchedulerLeaseConfig{}, err
	}
	return cfg, nil
}

func ValidateSchedulerLease(cfg SchedulerLeaseConfig) error {
	if err := ValidateMongoAdmin(cfg.MongoAdminConfig); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchedulerLeaseConfig, err)
	}
	if cfg.TTL <= 0 {
		return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidSchedulerLeaseConfig)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TIMEOUT must be positive", ErrInvalidSchedulerLeaseConfig)
	}
	return nil
}
