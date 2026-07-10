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
	DefaultFeatureEconomyCoinAdminEnabled       = false
	DefaultFeatureEconomyCoinRankEnabled        = false
	DefaultFeatureEconomyCoinResetEnabled       = false
	DefaultFeatureEconomyRPSEnabled             = false
	DefaultFeatureEconomyGameEnabled            = false
	DefaultFeatureEconomyShopEnabled            = false
	DefaultFeatureEconomyProfileEnabled         = false
	DefaultFeatureWorkEnabled                   = false
	DefaultFeatureWarningsEnabled               = false
	DefaultFeatureWarningSettingsEnabled        = false
	DefaultFeatureWarningRemovalEnabled         = false
	DefaultFeatureWarningIssueEnabled           = false
	DefaultFeatureMessageCleanupEnabled         = false
	DefaultFeatureDeleteDataEnabled             = false
	DefaultFeatureTranslateEnabled              = false
	DefaultFeatureBalanceQueryEnabled           = false
	DefaultFeatureRedeemEnabled                 = false
	DefaultFeatureAutoChatConfigEnabled         = false
	DefaultFeatureAutoNotificationConfigEnabled = false
	DefaultFeatureAntiScamConfigEnabled         = false
	DefaultFeatureAntiScamReportEnabled         = false
	DefaultFeatureAntiScamMessageDeleteEnabled  = false
	DefaultFeatureLoggingConfigEnabled          = false
	DefaultFeatureLoggingMessageEventsEnabled   = false
	DefaultFeatureGachaPrizeListEnabled         = false
	DefaultFeatureGachaDrawEnabled              = false
	DefaultFeatureGachaPrizeCreateEnabled       = false
	DefaultFeatureGachaPrizeEditEnabled         = false
	DefaultFeatureGachaPrizeDeleteEnabled       = false
	DefaultFeatureLotteryDisabledCommandEnabled = false
	DefaultFeatureStatsQueryEnabled             = false
	DefaultFeatureStatsCreateEnabled            = false
	DefaultFeatureStatsRoleCountEnabled         = false
	DefaultFeatureStatsDeleteEnabled            = false
	DefaultFeatureStatsRenameWorkerEnabled      = false
	DefaultFeatureBirthdayConfigEnabled         = false
	DefaultFeatureAnnouncementConfigEnabled     = false
	DefaultFeatureAnnouncementSendEnabled       = false
	DefaultFeatureAnnouncementRelayEnabled      = false
	DefaultFeatureTextXPConfigEnabled           = false
	DefaultFeatureTextXPAccrualEnabled          = false
	DefaultFeatureVoiceXPConfigEnabled          = false
	DefaultFeatureVoiceXPSessionsEnabled        = false
	DefaultFeatureXPRoleConfigEnabled           = false
	DefaultFeatureXPProfileDisabledEnabled      = false
	DefaultFeatureXPAdminEnabled                = false
	DefaultFeatureXPResetEnabled                = false
	DefaultFeatureXPRankEnabled                 = false
	DefaultFeatureVoiceRoomConfigEnabled        = false
	DefaultFeatureVoiceRoomLockEnabled          = false
	DefaultFeatureJoinRoleConfigEnabled         = false
	DefaultFeatureJoinRoleAssignmentEnabled     = false
	DefaultFeatureWelcomeMessageConfigEnabled   = false
	DefaultFeatureWelcomeMessageDeliveryEnabled = false
	DefaultFeatureLeaveMessageDeliveryEnabled   = false
	DefaultFeatureVerificationConfigEnabled     = false
	DefaultFeatureVerificationFlowEnabled       = false
	DefaultFeatureAccountAgeConfigEnabled       = false
	DefaultFeatureAccountAgePolicyEnabled       = false
	DefaultFeatureRoleSelectionEnabled          = false
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
	FeatureEconomyCoinAdminEnabled       bool
	FeatureEconomyCoinRankEnabled        bool
	FeatureEconomyCoinResetEnabled       bool
	FeatureEconomyRPSEnabled             bool
	FeatureEconomyGameEnabled            bool
	FeatureEconomyShopEnabled            bool
	FeatureEconomyProfileEnabled         bool
	FeatureWorkEnabled                   bool
	FeatureWarningsEnabled               bool
	FeatureWarningSettingsEnabled        bool
	FeatureWarningRemovalEnabled         bool
	FeatureWarningIssueEnabled           bool
	FeatureMessageCleanupEnabled         bool
	FeatureDeleteDataEnabled             bool
	FeatureTranslateEnabled              bool
	FeatureBalanceQueryEnabled           bool
	FeatureRedeemEnabled                 bool
	FeatureAutoChatConfigEnabled         bool
	FeatureAutoNotificationConfigEnabled bool
	FeatureAntiScamConfigEnabled         bool
	FeatureAntiScamReportEnabled         bool
	FeatureAntiScamMessageDeleteEnabled  bool
	FeatureLoggingConfigEnabled          bool
	FeatureLoggingMessageEventsEnabled   bool
	FeatureGachaPrizeListEnabled         bool
	FeatureGachaDrawEnabled              bool
	FeatureGachaPrizeCreateEnabled       bool
	FeatureGachaPrizeEditEnabled         bool
	FeatureGachaPrizeDeleteEnabled       bool
	FeatureLotteryDisabledCommandEnabled bool
	FeatureStatsQueryEnabled             bool
	FeatureStatsCreateEnabled            bool
	FeatureStatsRoleCountEnabled         bool
	FeatureStatsDeleteEnabled            bool
	FeatureStatsRenameWorkerEnabled      bool
	FeatureBirthdayConfigEnabled         bool
	FeatureAnnouncementConfigEnabled     bool
	FeatureAnnouncementSendEnabled       bool
	FeatureAnnouncementRelayEnabled      bool
	FeatureTextXPConfigEnabled           bool
	FeatureTextXPAccrualEnabled          bool
	FeatureVoiceXPConfigEnabled          bool
	FeatureVoiceXPSessionsEnabled        bool
	FeatureXPRoleConfigEnabled           bool
	FeatureXPProfileDisabledEnabled      bool
	FeatureXPAdminEnabled                bool
	FeatureXPResetEnabled                bool
	FeatureXPRankEnabled                 bool
	FeatureVoiceRoomConfigEnabled        bool
	FeatureVoiceRoomLockEnabled          bool
	FeatureJoinRoleConfigEnabled         bool
	FeatureJoinRoleAssignmentEnabled     bool
	FeatureWelcomeMessageConfigEnabled   bool
	FeatureWelcomeMessageDeliveryEnabled bool
	FeatureLeaveMessageDeliveryEnabled   bool
	FeatureVerificationConfigEnabled     bool
	FeatureVerificationFlowEnabled       bool
	FeatureAccountAgeConfigEnabled       bool
	FeatureAccountAgePolicyEnabled       bool
	FeatureRoleSelectionEnabled          bool
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
