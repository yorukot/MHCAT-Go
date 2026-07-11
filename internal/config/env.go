package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type LookupFunc func(string) (string, bool)

func Load() (Config, error) {
	return LoadWithLookup(os.LookupEnv)
}

func LoadWithLookup(lookup LookupFunc) (Config, error) {
	environment := getString(lookup, "MHCAT_ENV", DefaultEnv)
	lookup = withEnvironmentDefaults(lookup, environment)
	cfg := Config{
		Env:                                  environment,
		LogLevel:                             getString(lookup, "MHCAT_LOG_LEVEL", DefaultLogLevel),
		LogFormat:                            getString(lookup, "MHCAT_LOG_FORMAT", DefaultLogFormat),
		DiscordEnableGateway:                 DefaultDiscordEnableGateway,
		DiscordShardID:                       DefaultDiscordShardID,
		DiscordShardCount:                    DefaultDiscordShardCount,
		DiscordMessageContentIntent:          DefaultMessageContentIntent,
		DiscordGuildMembersIntent:            DefaultGuildMembersIntent,
		DiscordGuildMessagesIntent:           DefaultGuildMessagesIntent,
		DiscordMessageReactionsIntent:        DefaultMessageReactionsIntent,
		DiscordVoiceStateIntent:              DefaultVoiceStateIntent,
		DiscordGatewayConnectTimeout:         DefaultGatewayConnectTimeout,
		DiscordInteractionTimeout:            DefaultInteractionTimeout,
		DiscordGatewaySmokeTest:              DefaultGatewaySmokeTest,
		DiscordGatewaySmokeTimeout:           DefaultGatewaySmokeTimeout,
		FeatureTicketsEnabled:                DefaultFeatureTicketsEnabled,
		FeaturePollsEnabled:                  DefaultFeaturePollsEnabled,
		FeatureUsageTrackingEnabled:          DefaultFeatureUsageTrackingEnabled,
		FeatureEconomyQueryEnabled:           DefaultFeatureEconomyQueryEnabled,
		FeatureEconomySignInEnabled:          DefaultFeatureEconomySignInEnabled,
		FeatureEconomySettingsEnabled:        DefaultFeatureEconomySettingsEnabled,
		FeatureEconomyCoinAdminEnabled:       DefaultFeatureEconomyCoinAdminEnabled,
		FeatureEconomyCoinRankEnabled:        DefaultFeatureEconomyCoinRankEnabled,
		FeatureEconomyCoinResetEnabled:       DefaultFeatureEconomyCoinResetEnabled,
		FeatureEconomyRPSEnabled:             DefaultFeatureEconomyRPSEnabled,
		FeatureEconomyGameEnabled:            DefaultFeatureEconomyGameEnabled,
		FeatureEconomyShopEnabled:            DefaultFeatureEconomyShopEnabled,
		FeatureEconomyProfileEnabled:         DefaultFeatureEconomyProfileEnabled,
		FeatureWorkEnabled:                   DefaultFeatureWorkEnabled,
		FeatureWarningsEnabled:               DefaultFeatureWarningsEnabled,
		FeatureWarningSettingsEnabled:        DefaultFeatureWarningSettingsEnabled,
		FeatureWarningRemovalEnabled:         DefaultFeatureWarningRemovalEnabled,
		FeatureWarningIssueEnabled:           DefaultFeatureWarningIssueEnabled,
		FeatureMessageCleanupEnabled:         DefaultFeatureMessageCleanupEnabled,
		FeatureDeleteDataEnabled:             DefaultFeatureDeleteDataEnabled,
		FeatureTranslateEnabled:              DefaultFeatureTranslateEnabled,
		FeatureBalanceQueryEnabled:           DefaultFeatureBalanceQueryEnabled,
		FeatureRedeemEnabled:                 DefaultFeatureRedeemEnabled,
		FeatureAutoChatConfigEnabled:         DefaultFeatureAutoChatConfigEnabled,
		FeatureAutoChatFallbackEnabled:       DefaultFeatureAutoChatFallbackEnabled,
		FeatureAutoChatPaidHandoffEnabled:    DefaultFeatureAutoChatPaidHandoffEnabled,
		AutoChatPaidOwnershipConfirmed:       DefaultAutoChatPaidOwnershipConfirmed,
		FeatureAutoNotificationConfigEnabled: DefaultFeatureAutoNotificationConfigEnabled,
		FeatureAutoNotificationDelivery:      DefaultFeatureAutoNotificationDelivery,
		FeatureDailyResetSchedulerEnabled:    DefaultFeatureDailyResetSchedulerEnabled,
		FeatureWorkPayoutSchedulerEnabled:    DefaultFeatureWorkPayoutSchedulerEnabled,
		FeatureAntiScamConfigEnabled:         DefaultFeatureAntiScamConfigEnabled,
		FeatureAntiScamReportEnabled:         DefaultFeatureAntiScamReportEnabled,
		FeatureAntiScamMessageDeleteEnabled:  DefaultFeatureAntiScamMessageDeleteEnabled,
		FeatureLoggingConfigEnabled:          DefaultFeatureLoggingConfigEnabled,
		FeatureLoggingMessageEventsEnabled:   DefaultFeatureLoggingMessageEventsEnabled,
		FeatureLoggingChannelEventsEnabled:   DefaultFeatureLoggingChannelEventsEnabled,
		FeatureLoggingVoiceEventsEnabled:     DefaultFeatureLoggingVoiceEventsEnabled,
		FeatureGachaPrizeListEnabled:         DefaultFeatureGachaPrizeListEnabled,
		FeatureGachaDrawEnabled:              DefaultFeatureGachaDrawEnabled,
		FeatureGachaPrizeCreateEnabled:       DefaultFeatureGachaPrizeCreateEnabled,
		FeatureGachaPrizeEditEnabled:         DefaultFeatureGachaPrizeEditEnabled,
		FeatureGachaPrizeDeleteEnabled:       DefaultFeatureGachaPrizeDeleteEnabled,
		FeatureLotteryDisabledCommandEnabled: DefaultFeatureLotteryDisabledCommandEnabled,
		FeatureLotteryComponentsEnabled:      DefaultFeatureLotteryComponentsEnabled,
		FeatureStatsQueryEnabled:             DefaultFeatureStatsQueryEnabled,
		FeatureStatsCreateEnabled:            DefaultFeatureStatsCreateEnabled,
		FeatureStatsRoleCountEnabled:         DefaultFeatureStatsRoleCountEnabled,
		FeatureStatsDeleteEnabled:            DefaultFeatureStatsDeleteEnabled,
		FeatureStatsRenameWorkerEnabled:      DefaultFeatureStatsRenameWorkerEnabled,
		FeatureBirthdayConfigEnabled:         DefaultFeatureBirthdayConfigEnabled,
		FeatureAnnouncementConfigEnabled:     DefaultFeatureAnnouncementConfigEnabled,
		FeatureAnnouncementSendEnabled:       DefaultFeatureAnnouncementSendEnabled,
		FeatureAnnouncementRelayEnabled:      DefaultFeatureAnnouncementRelayEnabled,
		FeatureTextXPConfigEnabled:           DefaultFeatureTextXPConfigEnabled,
		FeatureTextXPAccrualEnabled:          DefaultFeatureTextXPAccrualEnabled,
		FeatureVoiceXPConfigEnabled:          DefaultFeatureVoiceXPConfigEnabled,
		FeatureVoiceXPSessionsEnabled:        DefaultFeatureVoiceXPSessionsEnabled,
		FeatureXPRoleConfigEnabled:           DefaultFeatureXPRoleConfigEnabled,
		FeatureXPProfileDisabledEnabled:      DefaultFeatureXPProfileDisabledEnabled,
		FeatureXPAdminEnabled:                DefaultFeatureXPAdminEnabled,
		FeatureXPResetEnabled:                DefaultFeatureXPResetEnabled,
		FeatureXPRankEnabled:                 DefaultFeatureXPRankEnabled,
		FeatureVoiceRoomConfigEnabled:        DefaultFeatureVoiceRoomConfigEnabled,
		FeatureVoiceRoomLockEnabled:          DefaultFeatureVoiceRoomLockEnabled,
		FeatureJoinRoleConfigEnabled:         DefaultFeatureJoinRoleConfigEnabled,
		FeatureJoinRoleAssignmentEnabled:     DefaultFeatureJoinRoleAssignmentEnabled,
		FeatureWelcomeMessageConfigEnabled:   DefaultFeatureWelcomeMessageConfigEnabled,
		FeatureWelcomeMessageDeliveryEnabled: DefaultFeatureWelcomeMessageDeliveryEnabled,
		FeatureLeaveMessageDeliveryEnabled:   DefaultFeatureLeaveMessageDeliveryEnabled,
		FeatureVerificationConfigEnabled:     DefaultFeatureVerificationConfigEnabled,
		FeatureVerificationFlowEnabled:       DefaultFeatureVerificationFlowEnabled,
		FeatureAccountAgeConfigEnabled:       DefaultFeatureAccountAgeConfigEnabled,
		FeatureAccountAgePolicyEnabled:       DefaultFeatureAccountAgePolicyEnabled,
		FeatureRoleSelectionEnabled:          DefaultFeatureRoleSelectionEnabled,
		JobsDailyResetEnabled:                DefaultJobsDailyResetEnabled,
		JobsDailyResetTimeout:                DefaultDailyResetTimeout,
		JobsWorkPayoutEnabled:                DefaultJobsWorkPayoutEnabled,
		JobsWorkPayoutTimeout:                DefaultWorkPayoutTimeout,
		JobsWorkPayoutLeaseName:              DefaultWorkPayoutLeaseName,
		SchedulerLeaseEnabled:                DefaultSchedulerLeaseEnabled,
		SchedulerLeaseTTL:                    DefaultSchedulerLeaseTTL,
		SchedulerLeaseTimeout:                DefaultSchedulerLeaseTimeout,
		MongoConnectTimeout:                  DefaultMongoConnectTimeout,
		MongoPingTimeout:                     DefaultMongoPingTimeout,
		ShutdownTimeout:                      DefaultShutdownTimeout,
	}

	cfg.DiscordToken = getAliasedString(lookup, "MHCAT_DISCORD_TOKEN", "TOKEN", &cfg)
	cfg.MongoDBURI = getAliasedString(lookup, "MHCAT_MONGODB_URI", "MONGOOSE_CONNECTION_STRING", &cfg)
	cfg.MongoDBDatabase = getString(lookup, "MHCAT_MONGODB_DATABASE", "")
	cfg.ReportWebhookURL = getAliasedString(lookup, "MHCAT_REPORT_WEBHOOK_URL", "REPORT_WEBHOOK", &cfg)
	cfg.SchedulerLeaseOwner = getString(lookup, "MHCAT_SCHEDULER_LEASE_OWNER", "")

	var err error
	if cfg.DiscordEnableGateway, err = getBool(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY", DefaultDiscordEnableGateway); err != nil {
		return Config{}, err
	}
	if cfg.DiscordShardID, err = getInt(lookup, "MHCAT_DISCORD_SHARD_ID", DefaultDiscordShardID); err != nil {
		return Config{}, err
	}
	if cfg.DiscordShardCount, err = getInt(lookup, "MHCAT_DISCORD_SHARD_COUNT", DefaultDiscordShardCount); err != nil {
		return Config{}, err
	}
	if cfg.DiscordMessageContentIntent, err = getBool(lookup, "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT", DefaultMessageContentIntent); err != nil {
		return Config{}, err
	}
	if cfg.DiscordGuildMembersIntent, err = getBool(lookup, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT", DefaultGuildMembersIntent); err != nil {
		return Config{}, err
	}
	if cfg.DiscordGuildMessagesIntent, err = getBool(lookup, "MHCAT_DISCORD_GUILD_MESSAGES_INTENT", DefaultGuildMessagesIntent); err != nil {
		return Config{}, err
	}
	if cfg.DiscordMessageReactionsIntent, err = getBool(lookup, "MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT", DefaultMessageReactionsIntent); err != nil {
		return Config{}, err
	}
	if cfg.DiscordVoiceStateIntent, err = getBool(lookup, "MHCAT_DISCORD_VOICE_STATE_INTENT", DefaultVoiceStateIntent); err != nil {
		return Config{}, err
	}
	if cfg.DiscordGatewayConnectTimeout, err = getDuration(lookup, "MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT", DefaultGatewayConnectTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DiscordInteractionTimeout, err = getDuration(lookup, "MHCAT_DISCORD_INTERACTION_TIMEOUT", DefaultInteractionTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DiscordGatewaySmokeTest, err = getBool(lookup, "MHCAT_DISCORD_GATEWAY_SMOKE_TEST", DefaultGatewaySmokeTest); err != nil {
		return Config{}, err
	}
	if cfg.DiscordGatewaySmokeTimeout, err = getDuration(lookup, "MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT", DefaultGatewaySmokeTimeout); err != nil {
		return Config{}, err
	}
	if cfg.FeatureTicketsEnabled, err = getBool(lookup, "MHCAT_FEATURE_TICKETS_ENABLED", DefaultFeatureTicketsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeaturePollsEnabled, err = getBool(lookup, "MHCAT_FEATURE_POLLS_ENABLED", DefaultFeaturePollsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureUsageTrackingEnabled, err = getBool(lookup, "MHCAT_FEATURE_USAGE_TRACKING_ENABLED", DefaultFeatureUsageTrackingEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_QUERY_ENABLED", DefaultFeatureEconomyQueryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomySignInEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED", DefaultFeatureEconomySignInEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomySettingsEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED", DefaultFeatureEconomySettingsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyCoinAdminEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED", DefaultFeatureEconomyCoinAdminEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyCoinRankEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED", DefaultFeatureEconomyCoinRankEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyCoinResetEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED", DefaultFeatureEconomyCoinResetEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyRPSEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_RPS_ENABLED", DefaultFeatureEconomyRPSEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyGameEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_GAME_ENABLED", DefaultFeatureEconomyGameEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyShopEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_SHOP_ENABLED", DefaultFeatureEconomyShopEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomyProfileEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED", DefaultFeatureEconomyProfileEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWorkEnabled, err = getBool(lookup, "MHCAT_FEATURE_WORK_ENABLED", DefaultFeatureWorkEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWarningsEnabled, err = getBool(lookup, "MHCAT_FEATURE_WARNINGS_ENABLED", DefaultFeatureWarningsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWarningSettingsEnabled, err = getBool(lookup, "MHCAT_FEATURE_WARNING_SETTINGS_ENABLED", DefaultFeatureWarningSettingsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWarningRemovalEnabled, err = getBool(lookup, "MHCAT_FEATURE_WARNING_REMOVAL_ENABLED", DefaultFeatureWarningRemovalEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWarningIssueEnabled, err = getBool(lookup, "MHCAT_FEATURE_WARNING_ISSUE_ENABLED", DefaultFeatureWarningIssueEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureMessageCleanupEnabled, err = getBool(lookup, "MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED", DefaultFeatureMessageCleanupEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureDeleteDataEnabled, err = getBool(lookup, "MHCAT_FEATURE_DELETE_DATA_ENABLED", DefaultFeatureDeleteDataEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureTranslateEnabled, err = getBool(lookup, "MHCAT_FEATURE_TRANSLATE_ENABLED", DefaultFeatureTranslateEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureBalanceQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_BALANCE_QUERY_ENABLED", DefaultFeatureBalanceQueryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureRedeemEnabled, err = getBool(lookup, "MHCAT_FEATURE_REDEEM_ENABLED", DefaultFeatureRedeemEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoChatConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED", DefaultFeatureAutoChatConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoChatFallbackEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED", DefaultFeatureAutoChatFallbackEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoChatPaidHandoffEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED", DefaultFeatureAutoChatPaidHandoffEnabled); err != nil {
		return Config{}, err
	}
	if cfg.AutoChatPaidOwnershipConfirmed, err = getBool(lookup, "MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED", DefaultAutoChatPaidOwnershipConfirmed); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoNotificationConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED", DefaultFeatureAutoNotificationConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoNotificationDelivery, err = getBool(lookup, "MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED", DefaultFeatureAutoNotificationDelivery); err != nil {
		return Config{}, err
	}
	if cfg.FeatureDailyResetSchedulerEnabled, err = getBool(lookup, "MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED", DefaultFeatureDailyResetSchedulerEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWorkPayoutSchedulerEnabled, err = getBool(lookup, "MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED", DefaultFeatureWorkPayoutSchedulerEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAntiScamConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED", DefaultFeatureAntiScamConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAntiScamReportEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED", DefaultFeatureAntiScamReportEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAntiScamMessageDeleteEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED", DefaultFeatureAntiScamMessageDeleteEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLoggingConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOGGING_CONFIG_ENABLED", DefaultFeatureLoggingConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLoggingMessageEventsEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED", DefaultFeatureLoggingMessageEventsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLoggingChannelEventsEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED", DefaultFeatureLoggingChannelEventsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLoggingVoiceEventsEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED", DefaultFeatureLoggingVoiceEventsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaPrizeListEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED", DefaultFeatureGachaPrizeListEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaDrawEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_DRAW_ENABLED", DefaultFeatureGachaDrawEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaPrizeCreateEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED", DefaultFeatureGachaPrizeCreateEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaPrizeEditEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED", DefaultFeatureGachaPrizeEditEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaPrizeDeleteEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED", DefaultFeatureGachaPrizeDeleteEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLotteryDisabledCommandEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED", DefaultFeatureLotteryDisabledCommandEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLotteryComponentsEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED", DefaultFeatureLotteryComponentsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_QUERY_ENABLED", DefaultFeatureStatsQueryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsCreateEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_CREATE_ENABLED", DefaultFeatureStatsCreateEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsRoleCountEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED", DefaultFeatureStatsRoleCountEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsDeleteEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_DELETE_ENABLED", DefaultFeatureStatsDeleteEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsRenameWorkerEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED", DefaultFeatureStatsRenameWorkerEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureBirthdayConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED", DefaultFeatureBirthdayConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAnnouncementConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED", DefaultFeatureAnnouncementConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAnnouncementSendEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED", DefaultFeatureAnnouncementSendEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAnnouncementRelayEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED", DefaultFeatureAnnouncementRelayEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureTextXPConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED", DefaultFeatureTextXPConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureTextXPAccrualEnabled, err = getBool(lookup, "MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED", DefaultFeatureTextXPAccrualEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVoiceXPConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED", DefaultFeatureVoiceXPConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVoiceXPSessionsEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED", DefaultFeatureVoiceXPSessionsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPRoleConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED", DefaultFeatureXPRoleConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPProfileDisabledEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED", DefaultFeatureXPProfileDisabledEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPAdminEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_ADMIN_ENABLED", DefaultFeatureXPAdminEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPResetEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_RESET_ENABLED", DefaultFeatureXPResetEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPRankEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_RANK_ENABLED", DefaultFeatureXPRankEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVoiceRoomConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED", DefaultFeatureVoiceRoomConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVoiceRoomLockEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED", DefaultFeatureVoiceRoomLockEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureJoinRoleConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED", DefaultFeatureJoinRoleConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureJoinRoleAssignmentEnabled, err = getBool(lookup, "MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED", DefaultFeatureJoinRoleAssignmentEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWelcomeMessageConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED", DefaultFeatureWelcomeMessageConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWelcomeMessageDeliveryEnabled, err = getBool(lookup, "MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED", DefaultFeatureWelcomeMessageDeliveryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLeaveMessageDeliveryEnabled, err = getBool(lookup, "MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED", DefaultFeatureLeaveMessageDeliveryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVerificationConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED", DefaultFeatureVerificationConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVerificationFlowEnabled, err = getBool(lookup, "MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED", DefaultFeatureVerificationFlowEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAccountAgeConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED", DefaultFeatureAccountAgeConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAccountAgePolicyEnabled, err = getBool(lookup, "MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED", DefaultFeatureAccountAgePolicyEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureRoleSelectionEnabled, err = getBool(lookup, "MHCAT_FEATURE_ROLE_SELECTION_ENABLED", DefaultFeatureRoleSelectionEnabled); err != nil {
		return Config{}, err
	}
	if cfg.JobsDailyResetEnabled, err = getBool(lookup, "MHCAT_JOBS_DAILY_RESET_ENABLED", DefaultJobsDailyResetEnabled); err != nil {
		return Config{}, err
	}
	if cfg.JobsDailyResetTimeout, err = getDuration(lookup, "MHCAT_JOBS_DAILY_RESET_TIMEOUT", DefaultDailyResetTimeout); err != nil {
		return Config{}, err
	}
	if cfg.JobsWorkPayoutEnabled, err = getBool(lookup, "MHCAT_JOBS_WORK_PAYOUT_ENABLED", DefaultJobsWorkPayoutEnabled); err != nil {
		return Config{}, err
	}
	if cfg.JobsWorkPayoutTimeout, err = getDuration(lookup, "MHCAT_JOBS_WORK_PAYOUT_TIMEOUT", DefaultWorkPayoutTimeout); err != nil {
		return Config{}, err
	}
	cfg.JobsWorkPayoutLeaseName = getString(lookup, "MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME", DefaultWorkPayoutLeaseName)
	if cfg.SchedulerLeaseEnabled, err = getBool(lookup, "MHCAT_SCHEDULER_LEASE_ENABLED", DefaultSchedulerLeaseEnabled); err != nil {
		return Config{}, err
	}
	if cfg.SchedulerLeaseTTL, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TTL", DefaultSchedulerLeaseTTL); err != nil {
		return Config{}, err
	}
	if cfg.SchedulerLeaseTimeout, err = getDuration(lookup, "MHCAT_SCHEDULER_LEASE_TIMEOUT", DefaultSchedulerLeaseTimeout); err != nil {
		return Config{}, err
	}
	if cfg.Staging, err = loadStagingWithLookup(lookup); err != nil {
		return Config{}, err
	}
	if cfg.MongoConnectTimeout, err = getDuration(lookup, "MHCAT_MONGO_CONNECT_TIMEOUT", DefaultMongoConnectTimeout); err != nil {
		return Config{}, err
	}
	if cfg.MongoPingTimeout, err = getDuration(lookup, "MHCAT_MONGO_PING_TIMEOUT", DefaultMongoPingTimeout); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = getDuration(lookup, "MHCAT_SHUTDOWN_TIMEOUT", DefaultShutdownTimeout); err != nil {
		return Config{}, err
	}
	cfg.LegacyWelcomeSpecialGuildID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID", "")
	cfg.LegacyWelcomeSpecialBotID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID", "")
	cfg.LegacyWelcomeSpecialChannelID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID", "")
	cfg.LegacyWelcomeSpecialChatChannelID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID", "")
	cfg.LegacyWelcomeSpecialHelpChannelID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID", "")
	cfg.LegacyWelcomeSpecialBugChannelID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID", "")
	cfg.LegacyWelcomeSpecialSupportChannelID = getString(lookup, "MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID", "")

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func withEnvironmentDefaults(lookup LookupFunc, environment string) LookupFunc {
	if strings.TrimSpace(environment) != "production" {
		return lookup
	}
	return func(key string) (string, bool) {
		if value, ok := lookup(key); ok {
			return value, true
		}
		if _, ok := productionEnabledDefaults[key]; ok {
			return "true", true
		}
		return "", false
	}
}

var productionEnabledDefaults = map[string]struct{}{
	"MHCAT_DISCORD_ENABLE_GATEWAY":                       {},
	"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT":               {},
	"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":                 {},
	"MHCAT_DISCORD_GUILD_MESSAGES_INTENT":                {},
	"MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT":       {},
	"MHCAT_DISCORD_VOICE_STATE_INTENT":                   {},
	"MHCAT_FEATURE_TICKETS_ENABLED":                      {},
	"MHCAT_FEATURE_POLLS_ENABLED":                        {},
	"MHCAT_FEATURE_USAGE_TRACKING_ENABLED":               {},
	"MHCAT_FEATURE_ECONOMY_QUERY_ENABLED":                {},
	"MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED":               {},
	"MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED":             {},
	"MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED":           {},
	"MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED":            {},
	"MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED":           {},
	"MHCAT_FEATURE_ECONOMY_RPS_ENABLED":                  {},
	"MHCAT_FEATURE_ECONOMY_GAME_ENABLED":                 {},
	"MHCAT_FEATURE_ECONOMY_SHOP_ENABLED":                 {},
	"MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED":              {},
	"MHCAT_FEATURE_WORK_ENABLED":                         {},
	"MHCAT_FEATURE_WARNINGS_ENABLED":                     {},
	"MHCAT_FEATURE_WARNING_SETTINGS_ENABLED":             {},
	"MHCAT_FEATURE_WARNING_REMOVAL_ENABLED":              {},
	"MHCAT_FEATURE_WARNING_ISSUE_ENABLED":                {},
	"MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED":              {},
	"MHCAT_FEATURE_DELETE_DATA_ENABLED":                  {},
	"MHCAT_FEATURE_TRANSLATE_ENABLED":                    {},
	"MHCAT_FEATURE_BALANCE_QUERY_ENABLED":                {},
	"MHCAT_FEATURE_REDEEM_ENABLED":                       {},
	"MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED":              {},
	"MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED":            {},
	"MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED":        {},
	"MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED":            {},
	"MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED":     {},
	"MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED":   {},
	"MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED":        {},
	"MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED":        {},
	"MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED":             {},
	"MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED":             {},
	"MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED":     {},
	"MHCAT_FEATURE_LOGGING_CONFIG_ENABLED":               {},
	"MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED":       {},
	"MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED":       {},
	"MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED":         {},
	"MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED":             {},
	"MHCAT_FEATURE_GACHA_DRAW_ENABLED":                   {},
	"MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED":           {},
	"MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED":             {},
	"MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED":           {},
	"MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED":     {},
	"MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED":           {},
	"MHCAT_FEATURE_STATS_QUERY_ENABLED":                  {},
	"MHCAT_FEATURE_STATS_CREATE_ENABLED":                 {},
	"MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED":             {},
	"MHCAT_FEATURE_STATS_DELETE_ENABLED":                 {},
	"MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED":          {},
	"MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED":              {},
	"MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED":          {},
	"MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED":            {},
	"MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED":           {},
	"MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED":               {},
	"MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED":              {},
	"MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED":              {},
	"MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED":            {},
	"MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED":               {},
	"MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED": {},
	"MHCAT_FEATURE_XP_ADMIN_ENABLED":                     {},
	"MHCAT_FEATURE_XP_RESET_ENABLED":                     {},
	"MHCAT_FEATURE_XP_RANK_ENABLED":                      {},
	"MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED":            {},
	"MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED":              {},
	"MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED":             {},
	"MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED":         {},
	"MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED":       {},
	"MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED":     {},
	"MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED":       {},
	"MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED":          {},
	"MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED":            {},
	"MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED":           {},
	"MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED":           {},
	"MHCAT_FEATURE_ROLE_SELECTION_ENABLED":               {},
	"MHCAT_JOBS_DAILY_RESET_ENABLED":                     {},
	"MHCAT_JOBS_WORK_PAYOUT_ENABLED":                     {},
	"MHCAT_SCHEDULER_LEASE_ENABLED":                      {},
}

func getString(lookup LookupFunc, key, fallback string) string {
	if value, ok := lookup(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getAliasedString(lookup LookupFunc, primary, alias string, cfg *Config) string {
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

func getBool(lookup LookupFunc, key string, fallback bool) (bool, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("parse %s as bool: %w", key, err)
	}
	return parsed, nil
}

func getDuration(lookup LookupFunc, key string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s as duration: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return parsed, nil
}
