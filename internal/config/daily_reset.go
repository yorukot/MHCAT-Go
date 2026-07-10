package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DefaultDailyResetDryRun  = true
	DefaultDailyResetTimeout = 60 * time.Second
)

var ErrInvalidDailyResetConfig = errors.New("invalid daily reset config")

type DailyResetConfig struct {
	MongoAdminConfig
	Enabled               bool
	DryRun                bool
	Timeout               time.Duration
	SchedulerLeaseGate    bool
	SchedulerLeaseOwner   string
	SchedulerLeaseTTL     time.Duration
	SchedulerLeaseTimeout time.Duration
}

func LoadDailyReset() (DailyResetConfig, error) {
	return LoadDailyResetWithLookup(os.LookupEnv)
}

func LoadDailyResetWithLookup(lookup LookupFunc) (DailyResetConfig, error) {
	cfg, err := LoadDailyResetRawWithLookup(lookup)
	if err != nil {
		return DailyResetConfig{}, err
	}
	if err := ValidateDailyReset(cfg); err != nil {
		return DailyResetConfig{}, err
	}
	return cfg, nil
}

func LoadDailyResetRawWithLookup(lookup LookupFunc) (DailyResetConfig, error) {
	mongoCfg, err := LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		return DailyResetConfig{}, err
	}
	cfg := DailyResetConfig{
		MongoAdminConfig:      mongoCfg,
		Enabled:               DefaultJobsDailyResetEnabled,
		DryRun:                DefaultDailyResetDryRun,
		Timeout:               DefaultDailyResetTimeout,
		SchedulerLeaseGate:    DefaultSchedulerLeaseEnabled,
		SchedulerLeaseOwner:   getString(lookup, "MHCAT_SCHEDULER_LEASE_OWNER", ""),
		SchedulerLeaseTTL:     DefaultSchedulerLeaseTTL,
		SchedulerLeaseTimeout: DefaultSchedulerLeaseTimeout,
	}
	if cfg.Enabled, err = getBool(lookup, "MHCAT_JOBS_DAILY_RESET_ENABLED", DefaultJobsDailyResetEnabled); err != nil {
		return DailyResetConfig{}, err
	}
	if cfg.DryRun, err = getBool(lookup, "MHCAT_JOBS_DAILY_RESET_DRY_RUN", DefaultDailyResetDryRun); err != nil {
		return DailyResetConfig{}, err
	}
	if cfg.Timeout, err = getDuration(lookup, "MHCAT_JOBS_DAILY_RESET_TIMEOUT", DefaultDailyResetTimeout); err != nil {
		return DailyResetConfig{}, err
	}
	if cfg.SchedulerLeaseGate, err = getBool(lookup, "MHCAT_SCHEDULER_LEASE_ENABLED", DefaultSchedulerLeaseEnabled); err != nil {
		return DailyResetConfig{}, err
	}
	if cfg.SchedulerLeaseTTL, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TTL", DefaultSchedulerLeaseTTL); err != nil {
		return DailyResetConfig{}, err
	}
	if cfg.SchedulerLeaseTimeout, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TIMEOUT", DefaultSchedulerLeaseTimeout); err != nil {
		return DailyResetConfig{}, err
	}
	return cfg, nil
}

func ValidateDailyReset(cfg DailyResetConfig) error {
	if err := ValidateMongoAdmin(cfg.MongoAdminConfig); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDailyResetConfig, err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("%w: MHCAT_JOBS_DAILY_RESET_TIMEOUT must be positive", ErrInvalidDailyResetConfig)
	}
	if cfg.SchedulerLeaseTTL <= 0 {
		return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidDailyResetConfig)
	}
	if cfg.SchedulerLeaseTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TIMEOUT must be positive", ErrInvalidDailyResetConfig)
	}
	if !cfg.DryRun {
		if !cfg.Enabled {
			return fmt.Errorf("%w: apply requires MHCAT_JOBS_DAILY_RESET_ENABLED=true", ErrInvalidDailyResetConfig)
		}
		if !cfg.SchedulerLeaseGate {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_ENABLED=true", ErrInvalidDailyResetConfig)
		}
		if strings.TrimSpace(cfg.SchedulerLeaseOwner) == "" {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_OWNER", ErrInvalidDailyResetConfig)
		}
		if cfg.SchedulerLeaseTTL <= cfg.Timeout || cfg.SchedulerLeaseTTL-cfg.Timeout <= cfg.SchedulerLeaseTimeout {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_TTL greater than MHCAT_JOBS_DAILY_RESET_TIMEOUT plus MHCAT_SCHEDULER_LEASE_TIMEOUT", ErrInvalidDailyResetConfig)
		}
	}
	return nil
}
