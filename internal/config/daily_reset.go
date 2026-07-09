package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	DefaultDailyResetDryRun  = true
	DefaultDailyResetTimeout = 60 * time.Second
)

var ErrInvalidDailyResetConfig = errors.New("invalid daily reset config")

type DailyResetConfig struct {
	MongoAdminConfig
	Enabled bool
	DryRun  bool
	Timeout time.Duration
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
		MongoAdminConfig: mongoCfg,
		Enabled:          DefaultJobsDailyResetEnabled,
		DryRun:           DefaultDailyResetDryRun,
		Timeout:          DefaultDailyResetTimeout,
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
	return cfg, nil
}

func ValidateDailyReset(cfg DailyResetConfig) error {
	if err := ValidateMongoAdmin(cfg.MongoAdminConfig); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDailyResetConfig, err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("%w: MHCAT_JOBS_DAILY_RESET_TIMEOUT must be positive", ErrInvalidDailyResetConfig)
	}
	if !cfg.DryRun && !cfg.Enabled {
		return fmt.Errorf("%w: apply requires MHCAT_JOBS_DAILY_RESET_ENABLED=true", ErrInvalidDailyResetConfig)
	}
	return nil
}
