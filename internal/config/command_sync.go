package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	CommandSyncScopeGuild  = "guild"
	CommandSyncScopeGlobal = "global"

	DefaultCommandSyncScope                         = CommandSyncScopeGuild
	DefaultCommandSyncDryRun                        = true
	DefaultCommandSyncAllowDelete                   = false
	DefaultCommandSyncAllowBulkOverwrite            = false
	DefaultCommandSyncStrict                        = true
	DefaultCommandSyncFormat                        = "text"
	DefaultCommandSyncIncludeTickets                = false
	DefaultCommandSyncIncludePolls                  = false
	DefaultCommandSyncIncludeEconomyQuery           = false
	DefaultCommandSyncIncludeEconomySignIn          = false
	DefaultCommandSyncIncludeEconomySettings        = false
	DefaultCommandSyncIncludeEconomyCoinAdmin       = false
	DefaultCommandSyncIncludeEconomyCoinRank        = false
	DefaultCommandSyncIncludeEconomyProfile         = false
	DefaultCommandSyncIncludeWork                   = false
	DefaultCommandSyncIncludeWarnings               = false
	DefaultCommandSyncIncludeWarningSettings        = false
	DefaultCommandSyncIncludeWarningRemoval         = false
	DefaultCommandSyncIncludeWarningIssue           = false
	DefaultCommandSyncIncludeMessageCleanup         = false
	DefaultCommandSyncIncludeDeleteData             = false
	DefaultCommandSyncIncludeTranslate              = false
	DefaultCommandSyncIncludeBalanceQuery           = false
	DefaultCommandSyncIncludeRedeem                 = false
	DefaultCommandSyncIncludeAutoChatConfig         = false
	DefaultCommandSyncIncludeAutoNotificationConfig = false
	DefaultCommandSyncIncludeAntiScamConfig         = false
	DefaultCommandSyncIncludeAntiScamReport         = false
	DefaultCommandSyncIncludeLoggingConfig          = false
	DefaultCommandSyncIncludeGachaPrizeList         = false
	DefaultCommandSyncIncludeGachaPrizeCreate       = false
	DefaultCommandSyncIncludeGachaPrizeDelete       = false
	DefaultCommandSyncIncludeLotteryDisabledCommand = false
	DefaultCommandSyncIncludeStatsQuery             = false
	DefaultCommandSyncIncludeStatsDelete            = false
	DefaultCommandSyncIncludeBirthdayConfig         = false
	DefaultCommandSyncIncludeAnnouncementConfig     = false
	DefaultCommandSyncIncludeAnnouncementSend       = false
	DefaultCommandSyncIncludeTextXPConfig           = false
	DefaultCommandSyncIncludeVoiceXPConfig          = false
	DefaultCommandSyncIncludeXPRoleConfig           = false
	DefaultCommandSyncIncludeXPProfileDisabled      = false
	DefaultCommandSyncIncludeVoiceRoomConfig        = false
	DefaultCommandSyncIncludeVoiceRoomLock          = false
	DefaultCommandSyncIncludeJoinRoleConfig         = false
	DefaultCommandSyncIncludeWelcomeMessageConfig   = false
	DefaultCommandSyncIncludeVerificationConfig     = false
	DefaultCommandSyncIncludeVerificationFlow       = false
	DefaultCommandSyncIncludeAccountAgeConfig       = false
)

var ErrInvalidCommandSyncConfig = errors.New("invalid command sync config")

type CommandSyncConfig struct {
	Env                           string
	LogLevel                      string
	LogFormat                     string
	DiscordToken                  string
	ApplicationID                 string
	Scope                         string
	GuildID                       string
	DryRun                        bool
	AllowDelete                   bool
	AllowBulkOverwrite            bool
	Strict                        bool
	Format                        string
	IncludeTickets                bool
	IncludePolls                  bool
	IncludeEconomyQuery           bool
	IncludeEconomySignIn          bool
	IncludeEconomySettings        bool
	IncludeEconomyCoinAdmin       bool
	IncludeEconomyCoinRank        bool
	IncludeEconomyProfile         bool
	IncludeWork                   bool
	IncludeWarnings               bool
	IncludeWarningSettings        bool
	IncludeWarningRemoval         bool
	IncludeWarningIssue           bool
	IncludeMessageCleanup         bool
	IncludeDeleteData             bool
	IncludeTranslate              bool
	IncludeBalanceQuery           bool
	IncludeRedeem                 bool
	IncludeAutoChatConfig         bool
	IncludeAutoNotificationConfig bool
	IncludeAntiScamConfig         bool
	IncludeAntiScamReport         bool
	IncludeLoggingConfig          bool
	IncludeGachaPrizeList         bool
	IncludeGachaPrizeCreate       bool
	IncludeGachaPrizeDelete       bool
	IncludeLotteryDisabledCommand bool
	IncludeStatsQuery             bool
	IncludeStatsDelete            bool
	IncludeBirthdayConfig         bool
	IncludeAnnouncementConfig     bool
	IncludeAnnouncementSend       bool
	IncludeTextXPConfig           bool
	IncludeVoiceXPConfig          bool
	IncludeXPRoleConfig           bool
	IncludeXPProfileDisabled      bool
	IncludeVoiceRoomConfig        bool
	IncludeVoiceRoomLock          bool
	IncludeJoinRoleConfig         bool
	IncludeWelcomeMessageConfig   bool
	IncludeVerificationConfig     bool
	IncludeVerificationFlow       bool
	IncludeAccountAgeConfig       bool
	Staging                       StagingConfig
	AliasWarnings                 []AliasWarning
}

func LoadCommandSync() (CommandSyncConfig, error) {
	return LoadCommandSyncWithLookup(os.LookupEnv)
}

func LoadCommandSyncWithLookup(lookup LookupFunc) (CommandSyncConfig, error) {
	cfg, err := LoadCommandSyncRawWithLookup(lookup)
	if err != nil {
		return CommandSyncConfig{}, err
	}
	if err := ValidateCommandSync(cfg); err != nil {
		return CommandSyncConfig{}, err
	}
	return cfg, nil
}

func LoadCommandSyncRawWithLookup(lookup LookupFunc) (CommandSyncConfig, error) {
	cfg := CommandSyncConfig{
		Env:                           getString(lookup, "MHCAT_ENV", DefaultEnv),
		LogLevel:                      getString(lookup, "MHCAT_LOG_LEVEL", DefaultLogLevel),
		LogFormat:                     getString(lookup, "MHCAT_LOG_FORMAT", DefaultLogFormat),
		ApplicationID:                 getString(lookup, "MHCAT_DISCORD_APPLICATION_ID", ""),
		Scope:                         getString(lookup, "MHCAT_COMMAND_SYNC_SCOPE", DefaultCommandSyncScope),
		GuildID:                       getString(lookup, "MHCAT_COMMAND_SYNC_GUILD_ID", ""),
		DryRun:                        DefaultCommandSyncDryRun,
		AllowDelete:                   DefaultCommandSyncAllowDelete,
		AllowBulkOverwrite:            DefaultCommandSyncAllowBulkOverwrite,
		Strict:                        DefaultCommandSyncStrict,
		Format:                        DefaultCommandSyncFormat,
		IncludeTickets:                DefaultCommandSyncIncludeTickets,
		IncludePolls:                  DefaultCommandSyncIncludePolls,
		IncludeEconomyQuery:           DefaultCommandSyncIncludeEconomyQuery,
		IncludeEconomySignIn:          DefaultCommandSyncIncludeEconomySignIn,
		IncludeEconomySettings:        DefaultCommandSyncIncludeEconomySettings,
		IncludeEconomyCoinAdmin:       DefaultCommandSyncIncludeEconomyCoinAdmin,
		IncludeEconomyCoinRank:        DefaultCommandSyncIncludeEconomyCoinRank,
		IncludeEconomyProfile:         DefaultCommandSyncIncludeEconomyProfile,
		IncludeWork:                   DefaultCommandSyncIncludeWork,
		IncludeWarnings:               DefaultCommandSyncIncludeWarnings,
		IncludeWarningSettings:        DefaultCommandSyncIncludeWarningSettings,
		IncludeWarningRemoval:         DefaultCommandSyncIncludeWarningRemoval,
		IncludeWarningIssue:           DefaultCommandSyncIncludeWarningIssue,
		IncludeMessageCleanup:         DefaultCommandSyncIncludeMessageCleanup,
		IncludeDeleteData:             DefaultCommandSyncIncludeDeleteData,
		IncludeTranslate:              DefaultCommandSyncIncludeTranslate,
		IncludeBalanceQuery:           DefaultCommandSyncIncludeBalanceQuery,
		IncludeRedeem:                 DefaultCommandSyncIncludeRedeem,
		IncludeAutoChatConfig:         DefaultCommandSyncIncludeAutoChatConfig,
		IncludeAutoNotificationConfig: DefaultCommandSyncIncludeAutoNotificationConfig,
		IncludeAntiScamConfig:         DefaultCommandSyncIncludeAntiScamConfig,
		IncludeAntiScamReport:         DefaultCommandSyncIncludeAntiScamReport,
		IncludeLoggingConfig:          DefaultCommandSyncIncludeLoggingConfig,
		IncludeGachaPrizeList:         DefaultCommandSyncIncludeGachaPrizeList,
		IncludeGachaPrizeCreate:       DefaultCommandSyncIncludeGachaPrizeCreate,
		IncludeGachaPrizeDelete:       DefaultCommandSyncIncludeGachaPrizeDelete,
		IncludeLotteryDisabledCommand: DefaultCommandSyncIncludeLotteryDisabledCommand,
		IncludeStatsQuery:             DefaultCommandSyncIncludeStatsQuery,
		IncludeStatsDelete:            DefaultCommandSyncIncludeStatsDelete,
		IncludeBirthdayConfig:         DefaultCommandSyncIncludeBirthdayConfig,
		IncludeAnnouncementConfig:     DefaultCommandSyncIncludeAnnouncementConfig,
		IncludeAnnouncementSend:       DefaultCommandSyncIncludeAnnouncementSend,
		IncludeTextXPConfig:           DefaultCommandSyncIncludeTextXPConfig,
		IncludeVoiceXPConfig:          DefaultCommandSyncIncludeVoiceXPConfig,
		IncludeXPRoleConfig:           DefaultCommandSyncIncludeXPRoleConfig,
		IncludeXPProfileDisabled:      DefaultCommandSyncIncludeXPProfileDisabled,
		IncludeVoiceRoomConfig:        DefaultCommandSyncIncludeVoiceRoomConfig,
		IncludeVoiceRoomLock:          DefaultCommandSyncIncludeVoiceRoomLock,
		IncludeJoinRoleConfig:         DefaultCommandSyncIncludeJoinRoleConfig,
		IncludeWelcomeMessageConfig:   DefaultCommandSyncIncludeWelcomeMessageConfig,
		IncludeVerificationConfig:     DefaultCommandSyncIncludeVerificationConfig,
		IncludeVerificationFlow:       DefaultCommandSyncIncludeVerificationFlow,
		IncludeAccountAgeConfig:       DefaultCommandSyncIncludeAccountAgeConfig,
	}
	cfg.DiscordToken = getAliasedCommandSyncString(lookup, "MHCAT_DISCORD_TOKEN", "TOKEN", &cfg)

	var err error
	if cfg.DryRun, err = getBool(lookup, "MHCAT_COMMAND_SYNC_DRY_RUN", DefaultCommandSyncDryRun); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.AllowDelete, err = getBool(lookup, "MHCAT_COMMAND_SYNC_ALLOW_DELETE", DefaultCommandSyncAllowDelete); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.AllowBulkOverwrite, err = getBool(lookup, "MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE", DefaultCommandSyncAllowBulkOverwrite); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.Strict, err = getBool(lookup, "MHCAT_COMMAND_SYNC_STRICT", DefaultCommandSyncStrict); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeTickets, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TICKETS", DefaultCommandSyncIncludeTickets); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludePolls, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_POLLS", DefaultCommandSyncIncludePolls); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomyQuery, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY", DefaultCommandSyncIncludeEconomyQuery); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomySignIn, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN", DefaultCommandSyncIncludeEconomySignIn); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomySettings, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS", DefaultCommandSyncIncludeEconomySettings); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomyCoinAdmin, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN", DefaultCommandSyncIncludeEconomyCoinAdmin); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomyCoinRank, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK", DefaultCommandSyncIncludeEconomyCoinRank); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeEconomyProfile, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE", DefaultCommandSyncIncludeEconomyProfile); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWork, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WORK", DefaultCommandSyncIncludeWork); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWarnings, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS", DefaultCommandSyncIncludeWarnings); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWarningSettings, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS", DefaultCommandSyncIncludeWarningSettings); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWarningRemoval, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL", DefaultCommandSyncIncludeWarningRemoval); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWarningIssue, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE", DefaultCommandSyncIncludeWarningIssue); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeMessageCleanup, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP", DefaultCommandSyncIncludeMessageCleanup); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeDeleteData, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA", DefaultCommandSyncIncludeDeleteData); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeTranslate, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE", DefaultCommandSyncIncludeTranslate); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeBalanceQuery, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY", DefaultCommandSyncIncludeBalanceQuery); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeRedeem, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_REDEEM", DefaultCommandSyncIncludeRedeem); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAutoChatConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG", DefaultCommandSyncIncludeAutoChatConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAutoNotificationConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG", DefaultCommandSyncIncludeAutoNotificationConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAntiScamConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG", DefaultCommandSyncIncludeAntiScamConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAntiScamReport, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT", DefaultCommandSyncIncludeAntiScamReport); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeLoggingConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG", DefaultCommandSyncIncludeLoggingConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeGachaPrizeList, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST", DefaultCommandSyncIncludeGachaPrizeList); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeGachaPrizeCreate, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE", DefaultCommandSyncIncludeGachaPrizeCreate); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeGachaPrizeDelete, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE", DefaultCommandSyncIncludeGachaPrizeDelete); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeLotteryDisabledCommand, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND", DefaultCommandSyncIncludeLotteryDisabledCommand); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeStatsQuery, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY", DefaultCommandSyncIncludeStatsQuery); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeStatsDelete, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE", DefaultCommandSyncIncludeStatsDelete); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeBirthdayConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG", DefaultCommandSyncIncludeBirthdayConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAnnouncementConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG", DefaultCommandSyncIncludeAnnouncementConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAnnouncementSend, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND", DefaultCommandSyncIncludeAnnouncementSend); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeTextXPConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG", DefaultCommandSyncIncludeTextXPConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeVoiceXPConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG", DefaultCommandSyncIncludeVoiceXPConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeXPRoleConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG", DefaultCommandSyncIncludeXPRoleConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeXPProfileDisabled, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS", DefaultCommandSyncIncludeXPProfileDisabled); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeVoiceRoomConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG", DefaultCommandSyncIncludeVoiceRoomConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeVoiceRoomLock, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK", DefaultCommandSyncIncludeVoiceRoomLock); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeJoinRoleConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG", DefaultCommandSyncIncludeJoinRoleConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeWelcomeMessageConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG", DefaultCommandSyncIncludeWelcomeMessageConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeVerificationConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG", DefaultCommandSyncIncludeVerificationConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeVerificationFlow, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW", DefaultCommandSyncIncludeVerificationFlow); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.IncludeAccountAgeConfig, err = getBool(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG", DefaultCommandSyncIncludeAccountAgeConfig); err != nil {
		return CommandSyncConfig{}, err
	}
	if cfg.Staging, err = loadStagingWithLookup(lookup); err != nil {
		return CommandSyncConfig{}, err
	}

	return cfg, nil
}

func ValidateCommandSync(cfg CommandSyncConfig) error {
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_LEVEL must be debug, info, warn, or error", ErrInvalidCommandSyncConfig)
	}
	switch cfg.LogFormat {
	case "text", "json":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_FORMAT must be text or json", ErrInvalidCommandSyncConfig)
	}
	switch cfg.Format {
	case "text", "json":
	default:
		return fmt.Errorf("%w: command sync output format must be text or json", ErrInvalidCommandSyncConfig)
	}
	if cfg.DiscordToken == "" {
		return fmt.Errorf("%w: missing required env: MHCAT_DISCORD_TOKEN", ErrInvalidCommandSyncConfig)
	}
	if cfg.ApplicationID == "" {
		return fmt.Errorf("%w: missing required env: MHCAT_DISCORD_APPLICATION_ID", ErrInvalidCommandSyncConfig)
	}
	switch cfg.Scope {
	case CommandSyncScopeGuild:
		if cfg.GuildID == "" {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_GUILD_ID is required when scope is guild", ErrInvalidCommandSyncConfig)
		}
	case CommandSyncScopeGlobal:
		if cfg.GuildID != "" {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_GUILD_ID must be empty when scope is global", ErrInvalidCommandSyncConfig)
		}
	default:
		return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_SCOPE must be guild or global", ErrInvalidCommandSyncConfig)
	}
	if cfg.AllowDelete && cfg.DryRun {
		return fmt.Errorf("%w: allow-delete requires apply mode", ErrInvalidCommandSyncConfig)
	}
	if cfg.AllowBulkOverwrite && cfg.DryRun {
		return fmt.Errorf("%w: allow-bulk-overwrite requires apply mode", ErrInvalidCommandSyncConfig)
	}
	if cfg.IncludeTickets {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TICKETS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TICKETS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludePolls {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_POLLS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_POLLS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomyQuery {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomySignIn {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomySettings {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomyCoinAdmin {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomyCoinRank {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeEconomyProfile {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWork {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WORK requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WORK requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWarnings {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWarningSettings {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWarningRemoval {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWarningIssue {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeMessageCleanup {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeDeleteData {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeTranslate {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeBalanceQuery {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeRedeem {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_REDEEM requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_REDEEM requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAutoChatConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAutoNotificationConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAntiScamConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAntiScamReport {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeLoggingConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeGachaPrizeList {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeGachaPrizeCreate {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeGachaPrizeDelete {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeLotteryDisabledCommand {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeStatsQuery {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeStatsDelete {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeBirthdayConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAnnouncementConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAnnouncementSend {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeTextXPConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeVoiceXPConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeXPRoleConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeXPProfileDisabled {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeVoiceRoomConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeVoiceRoomLock {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeJoinRoleConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeWelcomeMessageConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeVerificationConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeVerificationFlow {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if cfg.IncludeAccountAgeConfig {
		if !cfg.Staging.Mode {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG requires MHCAT_STAGING_MODE=true", ErrInvalidCommandSyncConfig)
		}
		if cfg.Scope != CommandSyncScopeGuild {
			return fmt.Errorf("%w: MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG requires guild scope", ErrInvalidCommandSyncConfig)
		}
	}
	if err := ValidateStagingCommandSync(cfg.Staging, cfg); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCommandSyncConfig, err)
	}
	return nil
}

func getAliasedCommandSyncString(lookup LookupFunc, primary, alias string, cfg *CommandSyncConfig) string {
	primaryValue, primaryOK := lookup(primary)
	aliasValue, aliasOK := lookup(alias)
	primaryValue = strings.TrimSpace(primaryValue)
	aliasValue = strings.TrimSpace(aliasValue)
	if primaryOK {
		if aliasOK && primaryValue != aliasValue {
			cfg.AliasWarnings = append(cfg.AliasWarnings, AliasWarning{
				Primary:      primary,
				Alias:        alias,
				PrimaryValue: primaryValue,
				AliasValue:   aliasValue,
			})
		}
		return primaryValue
	}
	if aliasOK {
		return aliasValue
	}
	return ""
}
