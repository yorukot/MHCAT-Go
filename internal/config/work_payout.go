package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DefaultWorkPayoutEnabled   = DefaultJobsWorkPayoutEnabled
	DefaultWorkPayoutDryRun    = true
	DefaultWorkPayoutTimeout   = 60 * time.Second
	DefaultWorkPayoutLeaseName = "work-payout"
)

var ErrInvalidWorkPayoutConfig = errors.New("invalid work payout config")

type WorkPayoutConfig struct {
	MongoAdminConfig
	Enabled             bool
	DryRun              bool
	Timeout             time.Duration
	LeaseName           string
	SchedulerLeaseGate  bool
	SchedulerLeaseOwner string
	SchedulerLeaseTTL   time.Duration
}

func LoadWorkPayout() (WorkPayoutConfig, error) {
	return LoadWorkPayoutWithLookup(os.LookupEnv)
}

func LoadWorkPayoutWithLookup(lookup LookupFunc) (WorkPayoutConfig, error) {
	cfg, err := LoadWorkPayoutRawWithLookup(lookup)
	if err != nil {
		return WorkPayoutConfig{}, err
	}
	if err := ValidateWorkPayout(cfg); err != nil {
		return WorkPayoutConfig{}, err
	}
	return cfg, nil
}

func LoadWorkPayoutRawWithLookup(lookup LookupFunc) (WorkPayoutConfig, error) {
	mongoCfg, err := LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		return WorkPayoutConfig{}, err
	}
	cfg := WorkPayoutConfig{
		MongoAdminConfig:    mongoCfg,
		Enabled:             DefaultWorkPayoutEnabled,
		DryRun:              DefaultWorkPayoutDryRun,
		Timeout:             DefaultWorkPayoutTimeout,
		LeaseName:           getString(lookup, "MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME", DefaultWorkPayoutLeaseName),
		SchedulerLeaseGate:  DefaultSchedulerLeaseEnabled,
		SchedulerLeaseOwner: getString(lookup, "MHCAT_SCHEDULER_LEASE_OWNER", ""),
		SchedulerLeaseTTL:   DefaultSchedulerLeaseTTL,
	}
	if cfg.Enabled, err = getBool(lookup, "MHCAT_JOBS_WORK_PAYOUT_ENABLED", DefaultWorkPayoutEnabled); err != nil {
		return WorkPayoutConfig{}, err
	}
	if cfg.DryRun, err = getBool(lookup, "MHCAT_JOBS_WORK_PAYOUT_DRY_RUN", DefaultWorkPayoutDryRun); err != nil {
		return WorkPayoutConfig{}, err
	}
	if cfg.Timeout, err = getDuration(lookup, "MHCAT_JOBS_WORK_PAYOUT_TIMEOUT", DefaultWorkPayoutTimeout); err != nil {
		return WorkPayoutConfig{}, err
	}
	if cfg.SchedulerLeaseGate, err = getBool(lookup, "MHCAT_SCHEDULER_LEASE_ENABLED", DefaultSchedulerLeaseEnabled); err != nil {
		return WorkPayoutConfig{}, err
	}
	if cfg.SchedulerLeaseTTL, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TTL", DefaultSchedulerLeaseTTL); err != nil {
		return WorkPayoutConfig{}, err
	}
	return cfg, nil
}

func ValidateWorkPayout(cfg WorkPayoutConfig) error {
	if err := ValidateMongoAdmin(cfg.MongoAdminConfig); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidWorkPayoutConfig, err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("%w: MHCAT_JOBS_WORK_PAYOUT_TIMEOUT must be positive", ErrInvalidWorkPayoutConfig)
	}
	if strings.TrimSpace(cfg.LeaseName) == "" {
		return fmt.Errorf("%w: MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME is required", ErrInvalidWorkPayoutConfig)
	}
	if cfg.SchedulerLeaseTTL <= 0 {
		return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidWorkPayoutConfig)
	}
	if !cfg.DryRun {
		if !cfg.Enabled {
			return fmt.Errorf("%w: apply requires MHCAT_JOBS_WORK_PAYOUT_ENABLED=true", ErrInvalidWorkPayoutConfig)
		}
		if !cfg.SchedulerLeaseGate {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_ENABLED=true", ErrInvalidWorkPayoutConfig)
		}
		if strings.TrimSpace(cfg.SchedulerLeaseOwner) == "" {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_OWNER", ErrInvalidWorkPayoutConfig)
		}
		if cfg.SchedulerLeaseTTL <= cfg.Timeout {
			return fmt.Errorf("%w: apply requires MHCAT_SCHEDULER_LEASE_TTL greater than MHCAT_JOBS_WORK_PAYOUT_TIMEOUT", ErrInvalidWorkPayoutConfig)
		}
	}
	return nil
}
