package config

import "time"

const (
	DefaultEnv                                  = "development"
	DefaultLogLevel                             = "info"
	DefaultLogFormat                            = "text"
	DefaultDiscordEnableGateway                 = false
	DefaultMessageContentIntent                 = false
	DefaultGuildMembersIntent                   = false
	DefaultGuildMessagesIntent                  = false
	DefaultMessageReactionsIntent               = false
	DefaultVoiceStateIntent                     = false
	DefaultGatewayConnectTimeout                = 15 * time.Second
	DefaultInteractionTimeout                   = 2500 * time.Millisecond
	DefaultGatewaySmokeTest                     = false
	DefaultGatewaySmokeTimeout                  = 30 * time.Second
	DefaultFeatureTicketsEnabled                = false
	DefaultFeaturePollsEnabled                  = false
	DefaultFeatureEconomyQueryEnabled           = false
	DefaultFeatureEconomySignInEnabled          = false
	DefaultFeatureEconomySettingsEnabled        = false
	DefaultFeatureWorkEnabled                   = false
	DefaultFeatureWarningsEnabled               = false
	DefaultFeatureWarningSettingsEnabled        = false
	DefaultFeatureWarningRemovalEnabled         = false
	DefaultFeatureWarningIssueEnabled           = false
	DefaultFeatureMessageCleanupEnabled         = false
	DefaultFeatureTranslateEnabled              = false
	DefaultFeatureBalanceQueryEnabled           = false
	DefaultFeatureRedeemEnabled                 = false
	DefaultFeatureAutoChatConfigEnabled         = false
	DefaultFeatureAutoNotificationConfigEnabled = false
	DefaultFeatureAntiScamConfigEnabled         = false
	DefaultFeatureAntiScamReportEnabled         = false
	DefaultFeatureLoggingConfigEnabled          = false
	DefaultFeatureGachaPrizeListEnabled         = false
	DefaultFeatureLotteryDisabledCommandEnabled = false
	DefaultFeatureStatsQueryEnabled             = false
	DefaultFeatureBirthdayConfigEnabled         = false
	DefaultFeatureAnnouncementConfigEnabled     = false
	DefaultFeatureAnnouncementSendEnabled       = false
	DefaultFeatureAnnouncementRelayEnabled      = false
	DefaultFeatureTextXPConfigEnabled           = false
	DefaultFeatureVoiceXPConfigEnabled          = false
	DefaultFeatureXPProfileDisabledEnabled      = false
	DefaultFeatureVoiceRoomConfigEnabled        = false
	DefaultFeatureJoinRoleConfigEnabled         = false
	DefaultFeatureJoinRoleAssignmentEnabled     = false
	DefaultFeatureWelcomeMessageConfigEnabled   = false
	DefaultFeatureWelcomeMessageDeliveryEnabled = false
	DefaultFeatureLeaveMessageDeliveryEnabled   = false
	DefaultFeatureVerificationConfigEnabled     = false
	DefaultFeatureVerificationFlowEnabled       = false
	DefaultFeatureAccountAgeConfigEnabled       = false
	DefaultFeatureAccountAgePolicyEnabled       = false
	DefaultJobsDailyResetEnabled                = false
	DefaultMongoConnectTimeout                  = 10 * time.Second
	DefaultMongoPingTimeout                     = 5 * time.Second
	DefaultShutdownTimeout                      = 10 * time.Second
)

type Config struct {
	Env                                  string
	LogLevel                             string
	LogFormat                            string
	DiscordToken                         string
	DiscordEnableGateway                 bool
	DiscordMessageContentIntent          bool
	DiscordGuildMembersIntent            bool
	DiscordGuildMessagesIntent           bool
	DiscordMessageReactionsIntent        bool
	DiscordVoiceStateIntent              bool
	DiscordGatewayConnectTimeout         time.Duration
	DiscordInteractionTimeout            time.Duration
	DiscordGatewaySmokeTest              bool
	DiscordGatewaySmokeTimeout           time.Duration
	FeatureTicketsEnabled                bool
	FeaturePollsEnabled                  bool
	FeatureEconomyQueryEnabled           bool
	FeatureEconomySignInEnabled          bool
	FeatureEconomySettingsEnabled        bool
	FeatureWorkEnabled                   bool
	FeatureWarningsEnabled               bool
	FeatureWarningSettingsEnabled        bool
	FeatureWarningRemovalEnabled         bool
	FeatureWarningIssueEnabled           bool
	FeatureMessageCleanupEnabled         bool
	FeatureTranslateEnabled              bool
	FeatureBalanceQueryEnabled           bool
	FeatureRedeemEnabled                 bool
	FeatureAutoChatConfigEnabled         bool
	FeatureAutoNotificationConfigEnabled bool
	FeatureAntiScamConfigEnabled         bool
	FeatureAntiScamReportEnabled         bool
	FeatureLoggingConfigEnabled          bool
	FeatureGachaPrizeListEnabled         bool
	FeatureLotteryDisabledCommandEnabled bool
	FeatureStatsQueryEnabled             bool
	FeatureBirthdayConfigEnabled         bool
	FeatureAnnouncementConfigEnabled     bool
	FeatureAnnouncementSendEnabled       bool
	FeatureAnnouncementRelayEnabled      bool
	FeatureTextXPConfigEnabled           bool
	FeatureVoiceXPConfigEnabled          bool
	FeatureXPProfileDisabledEnabled      bool
	FeatureVoiceRoomConfigEnabled        bool
	FeatureJoinRoleConfigEnabled         bool
	FeatureJoinRoleAssignmentEnabled     bool
	FeatureWelcomeMessageConfigEnabled   bool
	FeatureWelcomeMessageDeliveryEnabled bool
	FeatureLeaveMessageDeliveryEnabled   bool
	FeatureVerificationConfigEnabled     bool
	FeatureVerificationFlowEnabled       bool
	FeatureAccountAgeConfigEnabled       bool
	FeatureAccountAgePolicyEnabled       bool
	JobsDailyResetEnabled                bool
	Staging                              StagingConfig
	MongoDBURI                           string
	MongoDBDatabase                      string
	ReportWebhookURL                     string
	MongoConnectTimeout                  time.Duration
	MongoPingTimeout                     time.Duration
	ShutdownTimeout                      time.Duration
	LegacyWelcomeSpecialGuildID          string
	LegacyWelcomeSpecialBotID            string
	LegacyWelcomeSpecialChannelID        string
	LegacyWelcomeSpecialChatChannelID    string
	LegacyWelcomeSpecialHelpChannelID    string
	LegacyWelcomeSpecialBugChannelID     string
	LegacyWelcomeSpecialSupportChannelID string
	AliasWarnings                        []AliasWarning
}

type AliasWarning struct {
	Primary      string
	Alias        string
	PrimaryValue string
	AliasValue   string
}

func (w AliasWarning) Message() string {
	return w.Primary + " differs from legacy alias " + w.Alias
}

func (w AliasWarning) RedactedFields() map[string]string {
	return map[string]string{
		w.Primary: RedactValue(w.Primary, w.PrimaryValue),
		w.Alias:   RedactValue(w.Alias, w.AliasValue),
	}
}
