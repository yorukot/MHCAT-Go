package config

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidConfig = errors.New("invalid config")

func Validate(cfg Config) error {
	var missing []string
	if strings.TrimSpace(cfg.DiscordToken) == "" {
		missing = append(missing, "MHCAT_DISCORD_TOKEN")
	}
	if strings.TrimSpace(cfg.MongoDBURI) == "" {
		missing = append(missing, "MHCAT_MONGODB_URI")
	}
	if strings.TrimSpace(cfg.MongoDBDatabase) == "" {
		missing = append(missing, "MHCAT_MONGODB_DATABASE")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing required env: %s", ErrInvalidConfig, strings.Join(missing, ", "))
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_LEVEL must be debug, info, warn, or error", ErrInvalidConfig)
	}

	switch cfg.LogFormat {
	case "text", "json":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_FORMAT must be text or json", ErrInvalidConfig)
	}

	if cfg.MongoConnectTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_MONGO_CONNECT_TIMEOUT must be positive", ErrInvalidConfig)
	}
	if cfg.DiscordGatewayConnectTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT must be positive", ErrInvalidConfig)
	}
	if cfg.DiscordInteractionTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_DISCORD_INTERACTION_TIMEOUT must be positive", ErrInvalidConfig)
	}
	if cfg.DiscordGatewaySmokeTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT must be positive", ErrInvalidConfig)
	}
	if cfg.DiscordGatewaySmokeTest && !cfg.DiscordEnableGateway {
		return fmt.Errorf("%w: MHCAT_DISCORD_GATEWAY_SMOKE_TEST requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
	}
	if cfg.FeatureAnnouncementRelayEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureAutoChatFallbackEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureAutoChatPaidHandoffEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.AutoChatPaidOwnershipConfirmed {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED requires MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureAutoNotificationDelivery {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.SchedulerLeaseEnabled {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED requires MHCAT_SCHEDULER_LEASE_ENABLED=true", ErrInvalidConfig)
		}
		if strings.TrimSpace(cfg.SchedulerLeaseOwner) == "" {
			return fmt.Errorf("%w: MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED requires MHCAT_SCHEDULER_LEASE_OWNER", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTTL <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTimeout <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TIMEOUT must be positive", ErrInvalidConfig)
		}
	}
	if cfg.FeatureDailyResetSchedulerEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.JobsDailyResetEnabled {
			return fmt.Errorf("%w: MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED requires MHCAT_JOBS_DAILY_RESET_ENABLED=true", ErrInvalidConfig)
		}
		if !cfg.SchedulerLeaseEnabled {
			return fmt.Errorf("%w: MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED requires MHCAT_SCHEDULER_LEASE_ENABLED=true", ErrInvalidConfig)
		}
		if strings.TrimSpace(cfg.SchedulerLeaseOwner) == "" {
			return fmt.Errorf("%w: MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED requires MHCAT_SCHEDULER_LEASE_OWNER", ErrInvalidConfig)
		}
		if cfg.JobsDailyResetTimeout <= 0 {
			return fmt.Errorf("%w: MHCAT_JOBS_DAILY_RESET_TIMEOUT must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTTL <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTimeout <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TIMEOUT must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTTL <= cfg.JobsDailyResetTimeout || cfg.SchedulerLeaseTTL-cfg.JobsDailyResetTimeout <= cfg.SchedulerLeaseTimeout {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be greater than MHCAT_JOBS_DAILY_RESET_TIMEOUT plus MHCAT_SCHEDULER_LEASE_TIMEOUT", ErrInvalidConfig)
		}
	}
	if cfg.FeatureWorkPayoutSchedulerEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.JobsWorkPayoutEnabled {
			return fmt.Errorf("%w: MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED requires MHCAT_JOBS_WORK_PAYOUT_ENABLED=true", ErrInvalidConfig)
		}
		if !cfg.SchedulerLeaseEnabled {
			return fmt.Errorf("%w: MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED requires MHCAT_SCHEDULER_LEASE_ENABLED=true", ErrInvalidConfig)
		}
		if strings.TrimSpace(cfg.SchedulerLeaseOwner) == "" {
			return fmt.Errorf("%w: MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED requires MHCAT_SCHEDULER_LEASE_OWNER", ErrInvalidConfig)
		}
		if strings.TrimSpace(cfg.JobsWorkPayoutLeaseName) == "" {
			return fmt.Errorf("%w: MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME is required", ErrInvalidConfig)
		}
		if cfg.JobsWorkPayoutTimeout <= 0 {
			return fmt.Errorf("%w: MHCAT_JOBS_WORK_PAYOUT_TIMEOUT must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTTL <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTimeout <= 0 {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TIMEOUT must be positive", ErrInvalidConfig)
		}
		if cfg.SchedulerLeaseTTL <= cfg.JobsWorkPayoutTimeout || cfg.SchedulerLeaseTTL-cfg.JobsWorkPayoutTimeout <= cfg.SchedulerLeaseTimeout {
			return fmt.Errorf("%w: MHCAT_SCHEDULER_LEASE_TTL must be greater than MHCAT_JOBS_WORK_PAYOUT_TIMEOUT plus MHCAT_SCHEDULER_LEASE_TIMEOUT", ErrInvalidConfig)
		}
	}
	if cfg.FeatureXPResetEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_XP_RESET_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_XP_RESET_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_XP_RESET_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureTextXPAccrualEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureVoiceXPSessionsEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordVoiceStateIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED requires MHCAT_DISCORD_VOICE_STATE_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureEconomyCoinResetEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureJoinRoleAssignmentEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureLeaveMessageDeliveryEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureWelcomeMessageDeliveryEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureVoiceRoomLockEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordVoiceStateIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED requires MHCAT_DISCORD_VOICE_STATE_INTENT=true", ErrInvalidConfig)
		}
	}
	if err := validateLegacyWelcomeSpecialConfig(cfg); err != nil {
		return err
	}
	if cfg.FeatureAccountAgePolicyEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureRoleSelectionEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_ROLE_SELECTION_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageReactionsIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ROLE_SELECTION_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureAntiScamReportEnabled && strings.TrimSpace(cfg.ReportWebhookURL) == "" {
		return fmt.Errorf("%w: MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED requires MHCAT_REPORT_WEBHOOK_URL or REPORT_WEBHOOK", ErrInvalidConfig)
	}
	if cfg.FeatureAntiScamMessageDeleteEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureLoggingMessageEventsEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMessagesIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED requires MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true", ErrInvalidConfig)
		}
		if !cfg.DiscordMessageContentIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED requires MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureLoggingChannelEventsEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureLoggingVoiceEventsEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordVoiceStateIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED requires MHCAT_DISCORD_VOICE_STATE_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureStatsCreateEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_CREATE_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_CREATE_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureStatsRoleCountEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureStatsRenameWorkerEnabled {
		if !cfg.DiscordEnableGateway {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
		}
		if !cfg.DiscordGuildMembersIntent {
			return fmt.Errorf("%w: MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED requires MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true", ErrInvalidConfig)
		}
	}
	if cfg.FeatureLotteryComponentsEnabled && !cfg.DiscordEnableGateway {
		return fmt.Errorf("%w: MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED requires MHCAT_DISCORD_ENABLE_GATEWAY=true", ErrInvalidConfig)
	}
	if err := ValidateStagingGatewaySmoke(cfg.Staging, cfg.DiscordGatewaySmokeTest); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	if cfg.MongoPingTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_MONGO_PING_TIMEOUT must be positive", ErrInvalidConfig)
	}
	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("%w: MHCAT_SHUTDOWN_TIMEOUT must be positive", ErrInvalidConfig)
	}
	return nil
}

func validateLegacyWelcomeSpecialConfig(cfg Config) error {
	values := []string{
		strings.TrimSpace(cfg.LegacyWelcomeSpecialGuildID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialBotID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialChannelID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialChatChannelID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialHelpChannelID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialBugChannelID),
		strings.TrimSpace(cfg.LegacyWelcomeSpecialSupportChannelID),
	}
	set := 0
	for _, value := range values {
		if value != "" {
			set++
		}
	}
	if set != 0 && set != len(values) {
		return fmt.Errorf("%w: MHCAT_LEGACY_WELCOME_SPECIAL_* values must be set together", ErrInvalidConfig)
	}
	return nil
}
