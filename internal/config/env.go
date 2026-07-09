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
	cfg := Config{
		Env:                                  getString(lookup, "MHCAT_ENV", DefaultEnv),
		LogLevel:                             getString(lookup, "MHCAT_LOG_LEVEL", DefaultLogLevel),
		LogFormat:                            getString(lookup, "MHCAT_LOG_FORMAT", DefaultLogFormat),
		DiscordEnableGateway:                 DefaultDiscordEnableGateway,
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
		FeatureEconomyQueryEnabled:           DefaultFeatureEconomyQueryEnabled,
		FeatureEconomySignInEnabled:          DefaultFeatureEconomySignInEnabled,
		FeatureEconomySettingsEnabled:        DefaultFeatureEconomySettingsEnabled,
		FeatureWorkEnabled:                   DefaultFeatureWorkEnabled,
		FeatureWarningsEnabled:               DefaultFeatureWarningsEnabled,
		FeatureTranslateEnabled:              DefaultFeatureTranslateEnabled,
		FeatureBalanceQueryEnabled:           DefaultFeatureBalanceQueryEnabled,
		FeatureAutoChatConfigEnabled:         DefaultFeatureAutoChatConfigEnabled,
		FeatureAutoNotificationConfigEnabled: DefaultFeatureAutoNotificationConfigEnabled,
		FeatureAntiScamConfigEnabled:         DefaultFeatureAntiScamConfigEnabled,
		FeatureAntiScamReportEnabled:         DefaultFeatureAntiScamReportEnabled,
		FeatureLoggingConfigEnabled:          DefaultFeatureLoggingConfigEnabled,
		FeatureGachaPrizeListEnabled:         DefaultFeatureGachaPrizeListEnabled,
		FeatureLotteryDisabledCommandEnabled: DefaultFeatureLotteryDisabledCommandEnabled,
		FeatureStatsQueryEnabled:             DefaultFeatureStatsQueryEnabled,
		FeatureBirthdayConfigEnabled:         DefaultFeatureBirthdayConfigEnabled,
		FeatureAnnouncementConfigEnabled:     DefaultFeatureAnnouncementConfigEnabled,
		FeatureAnnouncementSendEnabled:       DefaultFeatureAnnouncementSendEnabled,
		FeatureAnnouncementRelayEnabled:      DefaultFeatureAnnouncementRelayEnabled,
		FeatureTextXPConfigEnabled:           DefaultFeatureTextXPConfigEnabled,
		FeatureVoiceXPConfigEnabled:          DefaultFeatureVoiceXPConfigEnabled,
		FeatureXPProfileDisabledEnabled:      DefaultFeatureXPProfileDisabledEnabled,
		FeatureVoiceRoomConfigEnabled:        DefaultFeatureVoiceRoomConfigEnabled,
		FeatureJoinRoleConfigEnabled:         DefaultFeatureJoinRoleConfigEnabled,
		FeatureJoinRoleAssignmentEnabled:     DefaultFeatureJoinRoleAssignmentEnabled,
		FeatureWelcomeMessageConfigEnabled:   DefaultFeatureWelcomeMessageConfigEnabled,
		FeatureWelcomeMessageDeliveryEnabled: DefaultFeatureWelcomeMessageDeliveryEnabled,
		FeatureLeaveMessageDeliveryEnabled:   DefaultFeatureLeaveMessageDeliveryEnabled,
		FeatureVerificationConfigEnabled:     DefaultFeatureVerificationConfigEnabled,
		FeatureVerificationFlowEnabled:       DefaultFeatureVerificationFlowEnabled,
		FeatureAccountAgeConfigEnabled:       DefaultFeatureAccountAgeConfigEnabled,
		FeatureAccountAgePolicyEnabled:       DefaultFeatureAccountAgePolicyEnabled,
		JobsDailyResetEnabled:                DefaultJobsDailyResetEnabled,
		MongoConnectTimeout:                  DefaultMongoConnectTimeout,
		MongoPingTimeout:                     DefaultMongoPingTimeout,
		ShutdownTimeout:                      DefaultShutdownTimeout,
	}

	cfg.DiscordToken = getAliasedString(lookup, "MHCAT_DISCORD_TOKEN", "TOKEN", &cfg)
	cfg.MongoDBURI = getAliasedString(lookup, "MHCAT_MONGODB_URI", "MONGOOSE_CONNECTION_STRING", &cfg)
	cfg.MongoDBDatabase = getString(lookup, "MHCAT_MONGODB_DATABASE", "")
	cfg.ReportWebhookURL = getAliasedString(lookup, "MHCAT_REPORT_WEBHOOK_URL", "REPORT_WEBHOOK", &cfg)

	var err error
	if cfg.DiscordEnableGateway, err = getBool(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY", DefaultDiscordEnableGateway); err != nil {
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
	if cfg.FeatureEconomyQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_QUERY_ENABLED", DefaultFeatureEconomyQueryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomySignInEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED", DefaultFeatureEconomySignInEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureEconomySettingsEnabled, err = getBool(lookup, "MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED", DefaultFeatureEconomySettingsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWorkEnabled, err = getBool(lookup, "MHCAT_FEATURE_WORK_ENABLED", DefaultFeatureWorkEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureWarningsEnabled, err = getBool(lookup, "MHCAT_FEATURE_WARNINGS_ENABLED", DefaultFeatureWarningsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureTranslateEnabled, err = getBool(lookup, "MHCAT_FEATURE_TRANSLATE_ENABLED", DefaultFeatureTranslateEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureBalanceQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_BALANCE_QUERY_ENABLED", DefaultFeatureBalanceQueryEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoChatConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED", DefaultFeatureAutoChatConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAutoNotificationConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED", DefaultFeatureAutoNotificationConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAntiScamConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED", DefaultFeatureAntiScamConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureAntiScamReportEnabled, err = getBool(lookup, "MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED", DefaultFeatureAntiScamReportEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLoggingConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOGGING_CONFIG_ENABLED", DefaultFeatureLoggingConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureGachaPrizeListEnabled, err = getBool(lookup, "MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED", DefaultFeatureGachaPrizeListEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureLotteryDisabledCommandEnabled, err = getBool(lookup, "MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED", DefaultFeatureLotteryDisabledCommandEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureStatsQueryEnabled, err = getBool(lookup, "MHCAT_FEATURE_STATS_QUERY_ENABLED", DefaultFeatureStatsQueryEnabled); err != nil {
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
	if cfg.FeatureVoiceXPConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED", DefaultFeatureVoiceXPConfigEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureXPProfileDisabledEnabled, err = getBool(lookup, "MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED", DefaultFeatureXPProfileDisabledEnabled); err != nil {
		return Config{}, err
	}
	if cfg.FeatureVoiceRoomConfigEnabled, err = getBool(lookup, "MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED", DefaultFeatureVoiceRoomConfigEnabled); err != nil {
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
	if cfg.JobsDailyResetEnabled, err = getBool(lookup, "MHCAT_JOBS_DAILY_RESET_ENABLED", DefaultJobsDailyResetEnabled); err != nil {
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
