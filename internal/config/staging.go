package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultStagingMode              = false
	DefaultStagingRequireGuildScope = true
	DefaultStagingAllowCommandApply = false
	DefaultStagingAllowGatewaySmoke = false
	DefaultStagingSmokeTimeout      = 60 * time.Second
	DefaultStagingExpectedCommands  = "help,ping,info"
)

var ErrInvalidStagingConfig = errors.New("invalid staging config")

type StagingConfig struct {
	Mode                 bool
	GuildID              string
	AllowedApplicationID string
	RequireGuildScope    bool
	AllowCommandApply    bool
	AllowGatewaySmoke    bool
	SmokeTimeout         time.Duration
	ExpectedCommands     []string
}

func loadStagingWithLookup(lookup LookupFunc) (StagingConfig, error) {
	cfg := StagingConfig{
		Mode:                 DefaultStagingMode,
		GuildID:              getString(lookup, "MHCAT_STAGING_GUILD_ID", ""),
		AllowedApplicationID: getString(lookup, "MHCAT_STAGING_ALLOWED_APPLICATION_ID", ""),
		RequireGuildScope:    DefaultStagingRequireGuildScope,
		AllowCommandApply:    DefaultStagingAllowCommandApply,
		AllowGatewaySmoke:    DefaultStagingAllowGatewaySmoke,
		SmokeTimeout:         DefaultStagingSmokeTimeout,
		ExpectedCommands:     parseCSV(DefaultStagingExpectedCommands),
	}
	var err error
	if cfg.Mode, err = getBool(lookup, "MHCAT_STAGING_MODE", DefaultStagingMode); err != nil {
		return StagingConfig{}, err
	}
	if cfg.RequireGuildScope, err = getBool(lookup, "MHCAT_STAGING_REQUIRE_GUILD_SCOPE", DefaultStagingRequireGuildScope); err != nil {
		return StagingConfig{}, err
	}
	if cfg.AllowCommandApply, err = getBool(lookup, "MHCAT_STAGING_ALLOW_COMMAND_APPLY", DefaultStagingAllowCommandApply); err != nil {
		return StagingConfig{}, err
	}
	if cfg.AllowGatewaySmoke, err = getBool(lookup, "MHCAT_STAGING_ALLOW_GATEWAY_SMOKE", DefaultStagingAllowGatewaySmoke); err != nil {
		return StagingConfig{}, err
	}
	if cfg.SmokeTimeout, err = getDuration(lookup, "MHCAT_STAGING_SMOKE_TIMEOUT", DefaultStagingSmokeTimeout); err != nil {
		return StagingConfig{}, err
	}
	cfg.ExpectedCommands = parseCSV(getString(lookup, "MHCAT_STAGING_EXPECTED_COMMANDS", DefaultStagingExpectedCommands))
	if len(cfg.ExpectedCommands) == 0 {
		return StagingConfig{}, fmt.Errorf("%w: MHCAT_STAGING_EXPECTED_COMMANDS must not be empty", ErrInvalidStagingConfig)
	}
	if cfg.AllowCommandApply && !cfg.Mode {
		return StagingConfig{}, fmt.Errorf("%w: MHCAT_STAGING_ALLOW_COMMAND_APPLY requires MHCAT_STAGING_MODE=true", ErrInvalidStagingConfig)
	}
	if cfg.AllowCommandApply && strings.TrimSpace(cfg.GuildID) == "" {
		return StagingConfig{}, fmt.Errorf("%w: MHCAT_STAGING_GUILD_ID is required when command apply is allowed", ErrInvalidStagingConfig)
	}
	if cfg.AllowGatewaySmoke && !cfg.Mode {
		return StagingConfig{}, fmt.Errorf("%w: MHCAT_STAGING_ALLOW_GATEWAY_SMOKE requires MHCAT_STAGING_MODE=true", ErrInvalidStagingConfig)
	}
	return cfg, nil
}

func ValidateStagingGatewaySmoke(staging StagingConfig, smokeEnabled bool) error {
	if !smokeEnabled {
		return nil
	}
	if !staging.Mode {
		return fmt.Errorf("%w: gateway smoke requires MHCAT_STAGING_MODE=true", ErrInvalidStagingConfig)
	}
	if !staging.AllowGatewaySmoke {
		return fmt.Errorf("%w: gateway smoke requires MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true", ErrInvalidStagingConfig)
	}
	return nil
}

func ValidateStagingCommandSync(staging StagingConfig, cfg CommandSyncConfig) error {
	if staging.AllowedApplicationID != "" && cfg.ApplicationID != "" && cfg.ApplicationID != staging.AllowedApplicationID {
		return fmt.Errorf("%w: MHCAT_DISCORD_APPLICATION_ID does not match MHCAT_STAGING_ALLOWED_APPLICATION_ID", ErrInvalidStagingConfig)
	}
	if staging.Mode && staging.RequireGuildScope && cfg.Scope != CommandSyncScopeGuild {
		return fmt.Errorf("%w: staging command sync requires guild scope", ErrInvalidStagingConfig)
	}
	if staging.Mode && staging.RequireGuildScope && strings.TrimSpace(cfg.GuildID) == "" {
		return fmt.Errorf("%w: staging command sync requires guild id", ErrInvalidStagingConfig)
	}
	if staging.Mode && staging.GuildID != "" && cfg.GuildID != "" && cfg.GuildID != staging.GuildID {
		return fmt.Errorf("%w: command sync guild id does not match staging guild id", ErrInvalidStagingConfig)
	}
	if cfg.DryRun {
		return nil
	}
	if !staging.Mode {
		return fmt.Errorf("%w: apply mode requires MHCAT_STAGING_MODE=true", ErrInvalidStagingConfig)
	}
	if !staging.AllowCommandApply {
		return fmt.Errorf("%w: apply mode requires MHCAT_STAGING_ALLOW_COMMAND_APPLY=true", ErrInvalidStagingConfig)
	}
	if cfg.Scope != CommandSyncScopeGuild {
		return fmt.Errorf("%w: staging apply rejects global scope", ErrInvalidStagingConfig)
	}
	if strings.TrimSpace(cfg.GuildID) == "" {
		return fmt.Errorf("%w: staging apply requires guild id", ErrInvalidStagingConfig)
	}
	if cfg.AllowDelete {
		return fmt.Errorf("%w: staging apply rejects command deletion", ErrInvalidStagingConfig)
	}
	if cfg.AllowBulkOverwrite {
		return fmt.Errorf("%w: staging apply rejects bulk overwrite", ErrInvalidStagingConfig)
	}
	return nil
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
