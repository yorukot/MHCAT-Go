package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMissingRequiredEnvFailsValidation(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(nil))
	if err == nil {
		t.Fatal("expected missing env error")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLegacyAliasesWork(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"TOKEN":                      "legacy-token",
		"MONGOOSE_CONNECTION_STRING": "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":     "mhcat",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DiscordToken != "legacy-token" {
		t.Fatalf("expected legacy token, got %q", cfg.DiscordToken)
	}
	if cfg.MongoDBURI != "mongodb://localhost:27017/mhcat" {
		t.Fatalf("expected legacy mongo uri, got %q", cfg.MongoDBURI)
	}
}

func TestNewEnvOverridesLegacyAlias(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":        "new-token",
		"TOKEN":                      "legacy-token",
		"MHCAT_MONGODB_URI":          "mongodb://localhost:27017/new",
		"MONGOOSE_CONNECTION_STRING": "mongodb://localhost:27017/legacy",
		"MHCAT_MONGODB_DATABASE":     "mhcat",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DiscordToken != "new-token" {
		t.Fatalf("expected new token, got %q", cfg.DiscordToken)
	}
	if cfg.MongoDBURI != "mongodb://localhost:27017/new" {
		t.Fatalf("expected new mongo uri, got %q", cfg.MongoDBURI)
	}
	if len(cfg.AliasWarnings) != 2 {
		t.Fatalf("expected 2 alias warnings, got %d", len(cfg.AliasWarnings))
	}
}

func TestInvalidDurationFails(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":         "token",
		"MHCAT_MONGODB_URI":           "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":      "mhcat",
		"MHCAT_MONGO_CONNECT_TIMEOUT": "not-a-duration",
	}))
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}

func TestDefaultsAreSafe(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":           "token",
		"MHCAT_MONGODB_URI":             "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":        "mhcat",
		"ENABLE_MESSAGE_CONTENT_INTENT": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Env != DefaultEnv {
		t.Fatalf("expected default env %q, got %q", DefaultEnv, cfg.Env)
	}
	if cfg.LogLevel != DefaultLogLevel {
		t.Fatalf("expected default log level %q, got %q", DefaultLogLevel, cfg.LogLevel)
	}
	if cfg.LogFormat != DefaultLogFormat {
		t.Fatalf("expected default log format %q, got %q", DefaultLogFormat, cfg.LogFormat)
	}
	if cfg.DiscordEnableGateway {
		t.Fatal("gateway must be disabled by default")
	}
	if cfg.DiscordMessageContentIntent {
		t.Fatal("message content intent must be disabled by default")
	}
	if cfg.DiscordGuildMembersIntent {
		t.Fatal("guild members intent must be disabled by default")
	}
	if cfg.DiscordGuildMessagesIntent || cfg.DiscordMessageReactionsIntent || cfg.DiscordVoiceStateIntent {
		t.Fatalf("event intents must be disabled by default: %#v", cfg)
	}
	if cfg.DiscordGatewayConnectTimeout != 15*time.Second {
		t.Fatalf("unexpected gateway connect timeout: %v", cfg.DiscordGatewayConnectTimeout)
	}
	if cfg.DiscordInteractionTimeout != 2500*time.Millisecond {
		t.Fatalf("unexpected interaction timeout: %v", cfg.DiscordInteractionTimeout)
	}
	if cfg.DiscordGatewaySmokeTest {
		t.Fatal("gateway smoke test must be disabled by default")
	}
	if cfg.FeatureTicketsEnabled {
		t.Fatal("ticket feature must be disabled by default")
	}
	if cfg.FeaturePollsEnabled {
		t.Fatal("poll feature must be disabled by default")
	}
	if cfg.FeatureEconomyQueryEnabled {
		t.Fatal("economy query feature must be disabled by default")
	}
	if cfg.FeatureEconomySignInEnabled {
		t.Fatal("economy sign-in feature must be disabled by default")
	}
	if cfg.FeatureEconomySettingsEnabled {
		t.Fatal("economy settings feature must be disabled by default")
	}
	if cfg.FeatureEconomyCoinAdminEnabled {
		t.Fatal("economy coin-admin feature must be disabled by default")
	}
	if cfg.FeatureEconomyCoinRankEnabled {
		t.Fatal("economy coin-rank feature must be disabled by default")
	}
	if cfg.FeatureEconomyRPSEnabled {
		t.Fatal("economy RPS feature must be disabled by default")
	}
	if cfg.FeatureEconomyProfileEnabled {
		t.Fatal("economy profile feature must be disabled by default")
	}
	if cfg.FeatureWorkEnabled {
		t.Fatal("work feature must be disabled by default")
	}
	if cfg.FeatureWarningsEnabled {
		t.Fatal("warnings feature must be disabled by default")
	}
	if cfg.FeatureWarningSettingsEnabled {
		t.Fatal("warning settings feature must be disabled by default")
	}
	if cfg.FeatureWarningRemovalEnabled {
		t.Fatal("warning removal feature must be disabled by default")
	}
	if cfg.FeatureWarningIssueEnabled {
		t.Fatal("warning issue feature must be disabled by default")
	}
	if cfg.FeatureMessageCleanupEnabled {
		t.Fatal("message cleanup feature must be disabled by default")
	}
	if cfg.FeatureDeleteDataEnabled {
		t.Fatal("delete data feature must be disabled by default")
	}
	if cfg.FeatureTranslateEnabled {
		t.Fatal("translate feature must be disabled by default")
	}
	if cfg.FeatureBalanceQueryEnabled {
		t.Fatal("balance query feature must be disabled by default")
	}
	if cfg.FeatureRedeemEnabled {
		t.Fatal("redeem feature must be disabled by default")
	}
	if cfg.FeatureAutoChatConfigEnabled {
		t.Fatal("autochat config feature must be disabled by default")
	}
	if cfg.FeatureAutoNotificationConfigEnabled {
		t.Fatal("auto-notification config feature must be disabled by default")
	}
	if cfg.FeatureAntiScamConfigEnabled {
		t.Fatal("anti-scam config feature must be disabled by default")
	}
	if cfg.FeatureAntiScamReportEnabled {
		t.Fatal("anti-scam report feature must be disabled by default")
	}
	if cfg.FeatureLoggingConfigEnabled {
		t.Fatal("logging config feature must be disabled by default")
	}
	if cfg.FeatureGachaPrizeListEnabled {
		t.Fatal("gacha prize-list feature must be disabled by default")
	}
	if cfg.FeatureGachaDrawEnabled {
		t.Fatal("gacha draw feature must be disabled by default")
	}
	if cfg.FeatureGachaPrizeCreateEnabled {
		t.Fatal("gacha prize-create feature must be disabled by default")
	}
	if cfg.FeatureGachaPrizeEditEnabled {
		t.Fatal("gacha prize-edit feature must be disabled by default")
	}
	if cfg.FeatureGachaPrizeDeleteEnabled {
		t.Fatal("gacha prize-delete feature must be disabled by default")
	}
	if cfg.FeatureLotteryDisabledCommandEnabled {
		t.Fatal("lottery disabled-command feature must be disabled by default")
	}
	if cfg.FeatureStatsQueryEnabled {
		t.Fatal("stats query feature must be disabled by default")
	}
	if cfg.FeatureStatsCreateEnabled {
		t.Fatal("stats create feature must be disabled by default")
	}
	if cfg.FeatureStatsRoleCountEnabled {
		t.Fatal("stats role-count feature must be disabled by default")
	}
	if cfg.FeatureStatsDeleteEnabled {
		t.Fatal("stats delete feature must be disabled by default")
	}
	if cfg.FeatureBirthdayConfigEnabled {
		t.Fatal("birthday config feature must be disabled by default")
	}
	if cfg.FeatureAnnouncementConfigEnabled {
		t.Fatal("announcement config feature must be disabled by default")
	}
	if cfg.FeatureAnnouncementSendEnabled {
		t.Fatal("announcement send feature must be disabled by default")
	}
	if cfg.FeatureAnnouncementRelayEnabled {
		t.Fatal("announcement relay feature must be disabled by default")
	}
	if cfg.FeatureTextXPConfigEnabled {
		t.Fatal("text XP config feature must be disabled by default")
	}
	if cfg.FeatureVoiceXPConfigEnabled {
		t.Fatal("voice XP config feature must be disabled by default")
	}
	if cfg.FeatureXPRoleConfigEnabled {
		t.Fatal("XP role config feature must be disabled by default")
	}
	if cfg.FeatureXPProfileDisabledEnabled {
		t.Fatal("XP profile disabled commands feature must be disabled by default")
	}
	if cfg.FeatureXPAdminEnabled {
		t.Fatal("XP admin feature must be disabled by default")
	}
	if cfg.FeatureVoiceRoomConfigEnabled {
		t.Fatal("voice-room config feature must be disabled by default")
	}
	if cfg.FeatureVoiceRoomLockEnabled {
		t.Fatal("voice-room lock feature must be disabled by default")
	}
	if cfg.FeatureJoinRoleConfigEnabled {
		t.Fatal("join-role config feature must be disabled by default")
	}
	if cfg.FeatureJoinRoleAssignmentEnabled {
		t.Fatal("join-role assignment feature must be disabled by default")
	}
	if cfg.FeatureWelcomeMessageConfigEnabled {
		t.Fatal("welcome-message config feature must be disabled by default")
	}
	if cfg.FeatureWelcomeMessageDeliveryEnabled {
		t.Fatal("welcome-message delivery feature must be disabled by default")
	}
	if cfg.FeatureLeaveMessageDeliveryEnabled {
		t.Fatal("leave-message delivery feature must be disabled by default")
	}
	if cfg.FeatureVerificationConfigEnabled {
		t.Fatal("verification config feature must be disabled by default")
	}
	if cfg.FeatureVerificationFlowEnabled {
		t.Fatal("verification flow feature must be disabled by default")
	}
	if cfg.FeatureAccountAgeConfigEnabled {
		t.Fatal("account-age config feature must be disabled by default")
	}
	if cfg.FeatureAccountAgePolicyEnabled {
		t.Fatal("account-age policy feature must be disabled by default")
	}
	if cfg.JobsDailyResetEnabled {
		t.Fatal("daily reset job gate must be disabled by default")
	}
	if cfg.DiscordGatewaySmokeTimeout != 30*time.Second {
		t.Fatalf("unexpected gateway smoke timeout: %v", cfg.DiscordGatewaySmokeTimeout)
	}
	if cfg.MongoConnectTimeout != 10*time.Second {
		t.Fatalf("unexpected mongo connect timeout: %v", cfg.MongoConnectTimeout)
	}
	if cfg.MongoPingTimeout != 5*time.Second {
		t.Fatalf("unexpected mongo ping timeout: %v", cfg.MongoPingTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected shutdown timeout: %v", cfg.ShutdownTimeout)
	}
}

func TestFeatureTextXPConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_MONGODB_URI":                    "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":               "mhcat",
		"MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureTextXPConfigEnabled {
		t.Fatal("expected text XP config feature to be enabled explicitly")
	}
}

func TestFeatureVoiceXPConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureVoiceXPConfigEnabled {
		t.Fatal("expected voice XP config feature to be enabled explicitly")
	}
}

func TestFeatureXPProfileDisabledParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                "token",
		"MHCAT_MONGODB_URI":                                  "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                             "mhcat",
		"MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureXPProfileDisabledEnabled {
		t.Fatal("expected XP profile disabled commands feature to be enabled explicitly")
	}
}

func TestFeatureXPAdminParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":            "token",
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_FEATURE_XP_ADMIN_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureXPAdminEnabled {
		t.Fatal("expected XP admin feature to be enabled explicitly")
	}
}

func TestFeatureVoiceRoomConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_MONGODB_URI":                       "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                  "mhcat",
		"MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureVoiceRoomConfigEnabled {
		t.Fatal("expected voice-room config feature to be enabled explicitly")
	}
}

func TestFeatureVoiceRoomLockRequiresGatewayAndVoiceStateIntent(t *testing.T) {
	base := map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":          "true",
		"MHCAT_DISCORD_VOICE_STATE_INTENT":      "true",
	}
	cfg, err := LoadWithLookup(mapLookup(base))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureVoiceRoomLockEnabled {
		t.Fatal("expected voice-room lock feature to be enabled explicitly")
	}

	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":     "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_VOICE_STATE_INTENT": "MHCAT_DISCORD_VOICE_STATE_INTENT=true",
	} {
		env := map[string]string{}
		for k, v := range base {
			env[k] = v
		}
		env[key] = "false"
		_, err := LoadWithLookup(mapLookup(env))
		if !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("expected ErrInvalidConfig for %s, got %v", key, err)
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %q, got %v", want, err)
		}
	}
}

func TestFeatureRedeemParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":          "token",
		"MHCAT_MONGODB_URI":            "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":       "mhcat",
		"MHCAT_FEATURE_REDEEM_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureRedeemEnabled {
		t.Fatal("expected redeem feature to be enabled explicitly")
	}
}

func TestFeatureAntiScamConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAntiScamConfigEnabled {
		t.Fatal("expected anti-scam config feature to be enabled explicitly")
	}
}

func TestFeatureAntiScamReportRequiresWebhook(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED": "true",
	}))
	if err == nil {
		t.Fatal("expected anti-scam report without webhook to fail")
	}
	if !errors.Is(err, ErrInvalidConfig) || !strings.Contains(err.Error(), "MHCAT_REPORT_WEBHOOK_URL") {
		t.Fatalf("expected webhook validation error, got %v", err)
	}
}

func TestFeatureAntiScamReportParsesWithLegacyWebhookAlias(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED": "true",
		"REPORT_WEBHOOK":                         "https://example.test/webhook",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAntiScamReportEnabled {
		t.Fatal("expected anti-scam report feature to be enabled explicitly")
	}
	if cfg.ReportWebhookURL != "https://example.test/webhook" {
		t.Fatalf("report webhook = %q", cfg.ReportWebhookURL)
	}
}

func TestFeatureJoinRoleConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureJoinRoleConfigEnabled {
		t.Fatal("expected join-role config feature to be enabled explicitly")
	}
}

func TestFeatureJoinRoleAssignmentRequiresGatewayAndGuildMembers(t *testing.T) {
	base := map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_MONGODB_URI":                          "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                     "mhcat",
		"MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":               "true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":         "true",
	}
	cfg, err := LoadWithLookup(mapLookup(base))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureJoinRoleAssignmentEnabled {
		t.Fatal("expected join-role assignment feature to be enabled explicitly")
	}

	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":       "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT": "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true",
	} {
		env := map[string]string{}
		for k, v := range base {
			env[k] = v
		}
		env[key] = "false"
		_, err := LoadWithLookup(mapLookup(env))
		if err == nil {
			t.Fatalf("expected %s validation error", key)
		}
		if !errors.Is(err, ErrInvalidConfig) || !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %q, got %v", want, err)
		}
	}
}

func TestFeatureLeaveMessageDeliveryRequiresGatewayAndGuildMembers(t *testing.T) {
	base := map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_MONGODB_URI":                            "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                       "mhcat",
		"MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":                 "true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":           "true",
	}
	cfg, err := LoadWithLookup(mapLookup(base))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureLeaveMessageDeliveryEnabled {
		t.Fatal("expected leave-message delivery feature to be enabled explicitly")
	}

	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":       "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT": "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true",
	} {
		env := map[string]string{}
		for k, v := range base {
			env[k] = v
		}
		env[key] = "false"
		_, err := LoadWithLookup(mapLookup(env))
		if err == nil {
			t.Fatalf("expected %s validation error", key)
		}
		if !errors.Is(err, ErrInvalidConfig) || !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %q, got %v", want, err)
		}
	}
}

func TestFeatureWelcomeMessageConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_MONGODB_URI":                            "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                       "mhcat",
		"MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWelcomeMessageConfigEnabled {
		t.Fatal("expected welcome-message config feature to be enabled explicitly")
	}
}

func TestFeatureVerificationConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_MONGODB_URI":                         "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                    "mhcat",
		"MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureVerificationConfigEnabled {
		t.Fatal("expected verification config feature to be enabled explicitly")
	}
}

func TestFeatureVerificationFlowParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_MONGODB_URI":                       "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                  "mhcat",
		"MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureVerificationFlowEnabled {
		t.Fatal("expected verification flow feature to be enabled explicitly")
	}
}

func TestFeatureGachaPrizeListConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureGachaPrizeListEnabled {
		t.Fatal("expected gacha prize-list feature to be enabled explicitly")
	}
}

func TestFeatureGachaDrawConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_MONGODB_URI":                "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":           "mhcat",
		"MHCAT_FEATURE_GACHA_DRAW_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureGachaDrawEnabled {
		t.Fatal("expected gacha draw feature to be enabled explicitly")
	}
}

func TestFeatureGachaPrizeDeleteConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureGachaPrizeDeleteEnabled {
		t.Fatal("expected gacha prize-delete feature to be enabled explicitly")
	}
}

func TestFeatureGachaPrizeCreateConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureGachaPrizeCreateEnabled {
		t.Fatal("expected gacha prize-create feature to be enabled explicitly")
	}
}

func TestFeatureGachaPrizeEditConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureGachaPrizeEditEnabled {
		t.Fatal("expected gacha prize-edit feature to be enabled explicitly")
	}
}

func TestFeatureLotteryDisabledCommandConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_MONGODB_URI":                              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                         "mhcat",
		"MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureLotteryDisabledCommandEnabled {
		t.Fatal("expected lottery disabled-command feature to be enabled explicitly")
	}
}

func TestFeatureStatsQueryConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_MONGODB_URI":                 "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":            "mhcat",
		"MHCAT_FEATURE_STATS_QUERY_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureStatsQueryEnabled {
		t.Fatal("expected stats query feature to be enabled explicitly")
	}
}

func TestFeatureStatsDeleteConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                "token",
		"MHCAT_MONGODB_URI":                  "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":             "mhcat",
		"MHCAT_FEATURE_STATS_DELETE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureStatsDeleteEnabled {
		t.Fatal("expected stats delete feature to be enabled explicitly")
	}
}

func TestFeatureStatsCreateConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                "token",
		"MHCAT_MONGODB_URI":                  "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":             "mhcat",
		"MHCAT_FEATURE_STATS_CREATE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureStatsCreateEnabled {
		t.Fatal("expected stats create feature to be enabled explicitly")
	}
}

func TestFeatureStatsRoleCountConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureStatsRoleCountEnabled {
		t.Fatal("expected stats role-count feature to be enabled explicitly")
	}
}

func TestFeatureXPRoleConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_MONGODB_URI":                    "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":               "mhcat",
		"MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureXPRoleConfigEnabled {
		t.Fatal("expected XP role config feature to be enabled explicitly")
	}
}

func TestFeatureAnnouncementConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_MONGODB_URI":                         "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                    "mhcat",
		"MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAnnouncementConfigEnabled {
		t.Fatal("expected announcement config feature to be enabled explicitly")
	}
}

func TestFeatureBirthdayConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureBirthdayConfigEnabled {
		t.Fatal("expected birthday config feature to be enabled explicitly")
	}
}

func TestFeatureAutoChatConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAutoChatConfigEnabled {
		t.Fatal("expected autochat config feature to be enabled explicitly")
	}
}

func TestFeatureAutoNotificationConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_MONGODB_URI":                              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                         "mhcat",
		"MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAutoNotificationConfigEnabled {
		t.Fatal("expected auto-notification config feature to be enabled explicitly")
	}
}

func TestFeatureAnnouncementSendParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_MONGODB_URI":                       "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                  "mhcat",
		"MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAnnouncementSendEnabled {
		t.Fatal("expected announcement send feature to be enabled explicitly")
	}
}

func TestFeatureAnnouncementRelayRequiresGatewayMessagesAndContent(t *testing.T) {
	base := map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":             "true",
		"MHCAT_DISCORD_GUILD_MESSAGES_INTENT":      "true",
		"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT":     "true",
	}
	cfg, err := LoadWithLookup(mapLookup(base))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAnnouncementRelayEnabled {
		t.Fatal("expected announcement relay feature to be enabled explicitly")
	}

	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":         "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MESSAGES_INTENT":  "MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true",
		"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT": "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true",
	} {
		env := map[string]string{}
		for k, v := range base {
			env[k] = v
		}
		env[key] = "false"
		_, err := LoadWithLookup(mapLookup(env))
		if err == nil {
			t.Fatalf("expected %s validation error", key)
		}
		if !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("expected ErrInvalidConfig for %s, got %v", key, err)
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %q, got %v", want, err)
		}
	}
}

func TestFeatureTicketsConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":           "token",
		"MHCAT_MONGODB_URI":             "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":        "mhcat",
		"MHCAT_FEATURE_TICKETS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureTicketsEnabled {
		t.Fatal("expected ticket feature to be enabled explicitly")
	}
}

func TestFeaturePollsConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":         "token",
		"MHCAT_MONGODB_URI":           "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":      "mhcat",
		"MHCAT_FEATURE_POLLS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeaturePollsEnabled {
		t.Fatal("expected poll feature to be enabled explicitly")
	}
}

func TestFeatureEconomyQueryConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_MONGODB_URI":                   "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":              "mhcat",
		"MHCAT_FEATURE_ECONOMY_QUERY_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomyQueryEnabled {
		t.Fatal("expected economy query feature to be enabled explicitly")
	}
}

func TestFeatureEconomySignInConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_MONGODB_URI":                    "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":               "mhcat",
		"MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomySignInEnabled {
		t.Fatal("expected economy sign-in feature to be enabled explicitly")
	}
}

func TestFeatureEconomySettingsConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomySettingsEnabled {
		t.Fatal("expected economy settings feature to be enabled explicitly")
	}
}

func TestFeatureWorkConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":        "token",
		"MHCAT_MONGODB_URI":          "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":     "mhcat",
		"MHCAT_FEATURE_WORK_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWorkEnabled {
		t.Fatal("expected work feature to be enabled explicitly")
	}
}

func TestFeatureWarningsConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":            "token",
		"MHCAT_MONGODB_URI":              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":         "mhcat",
		"MHCAT_FEATURE_WARNINGS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWarningsEnabled {
		t.Fatal("expected warnings feature to be enabled explicitly")
	}
}

func TestFeatureWarningSettingsConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_MONGODB_URI":                      "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                 "mhcat",
		"MHCAT_FEATURE_WARNING_SETTINGS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWarningSettingsEnabled {
		t.Fatal("expected warning settings feature to be enabled explicitly")
	}
}

func TestFeatureWarningRemovalConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_WARNING_REMOVAL_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWarningRemovalEnabled {
		t.Fatal("expected warning removal feature to be enabled explicitly")
	}
}

func TestFeatureWarningIssueConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_MONGODB_URI":                   "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":              "mhcat",
		"MHCAT_FEATURE_WARNING_ISSUE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWarningIssueEnabled {
		t.Fatal("expected warning issue feature to be enabled explicitly")
	}
}

func TestFeatureMessageCleanupConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureMessageCleanupEnabled {
		t.Fatal("expected message cleanup feature to be enabled explicitly")
	}
}

func TestFeatureDeleteDataConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_MONGODB_URI":                 "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":            "mhcat",
		"MHCAT_FEATURE_DELETE_DATA_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureDeleteDataEnabled {
		t.Fatal("expected delete data feature to be enabled explicitly")
	}
}

func TestFeatureEconomyCoinAdminConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomyCoinAdminEnabled {
		t.Fatal("expected economy coin-admin feature to be enabled explicitly")
	}
}

func TestFeatureEconomyCoinRankConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_MONGODB_URI":                       "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                  "mhcat",
		"MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomyCoinRankEnabled {
		t.Fatal("expected economy coin-rank feature to be enabled explicitly")
	}
}

func TestFeatureEconomyRPSConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_MONGODB_URI":                 "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":            "mhcat",
		"MHCAT_FEATURE_ECONOMY_RPS_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomyRPSEnabled {
		t.Fatal("expected economy RPS feature to be enabled explicitly")
	}
}

func TestFeatureEconomyProfileConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureEconomyProfileEnabled {
		t.Fatal("expected economy profile feature to be enabled explicitly")
	}
}

func TestFeatureTranslateConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":             "token",
		"MHCAT_MONGODB_URI":               "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":          "mhcat",
		"MHCAT_FEATURE_TRANSLATE_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureTranslateEnabled {
		t.Fatal("expected translate feature to be enabled explicitly")
	}
}

func TestFeatureBalanceQueryConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_MONGODB_URI":                   "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":              "mhcat",
		"MHCAT_FEATURE_BALANCE_QUERY_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureBalanceQueryEnabled {
		t.Fatal("expected balance query feature to be enabled explicitly")
	}
}

func TestFeatureLoggingConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_MONGODB_URI":                    "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":               "mhcat",
		"MHCAT_FEATURE_LOGGING_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureLoggingConfigEnabled {
		t.Fatal("expected logging config feature to be enabled explicitly")
	}
}

func TestGatewaySmokeRequiresGatewayEnabled(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_MONGODB_URI":                "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":           "mhcat",
		"MHCAT_DISCORD_GATEWAY_SMOKE_TEST": "true",
	}))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestGatewayRuntimeConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_DISCORD_ENABLE_GATEWAY":             "true",
		"MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT":    "20s",
		"MHCAT_DISCORD_INTERACTION_TIMEOUT":        "3s",
		"MHCAT_DISCORD_GATEWAY_SMOKE_TEST":         "true",
		"MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT": "45s",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_STAGING_ALLOW_GATEWAY_SMOKE":        "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.DiscordEnableGateway || !cfg.DiscordGatewaySmokeTest {
		t.Fatalf("gateway config not enabled: %#v", cfg)
	}
	if cfg.DiscordGatewayConnectTimeout != 20*time.Second || cfg.DiscordInteractionTimeout != 3*time.Second || cfg.DiscordGatewaySmokeTimeout != 45*time.Second {
		t.Fatalf("unexpected durations: %#v", cfg)
	}
}

func TestExplicitPrivilegedIntentConfig(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_MONGODB_URI":                    "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":               "mhcat",
		"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT": "true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":   "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.DiscordMessageContentIntent {
		t.Fatal("expected message content intent to be enabled explicitly")
	}
	if !cfg.DiscordGuildMembersIntent {
		t.Fatal("expected guild members intent to be enabled explicitly")
	}
}

func TestExplicitEventIntentConfig(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_MONGODB_URI":                            "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                       "mhcat",
		"MHCAT_DISCORD_GUILD_MESSAGES_INTENT":          "true",
		"MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT": "true",
		"MHCAT_DISCORD_VOICE_STATE_INTENT":             "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.DiscordGuildMessagesIntent || !cfg.DiscordMessageReactionsIntent || !cfg.DiscordVoiceStateIntent {
		t.Fatalf("event intents not enabled: %#v", cfg)
	}
	if cfg.DiscordMessageContentIntent {
		t.Fatal("message content should not be enabled by event intents")
	}
}

func TestFeatureAccountAgeConfigParses(t *testing.T) {
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED": "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAccountAgeConfigEnabled {
		t.Fatal("expected account-age config feature to be enabled explicitly")
	}
}

func TestFeatureWelcomeMessageDeliveryRequiresGatewayAndGuildMembersIntent(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_MONGODB_URI":                              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                         "mhcat",
		"MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED": "true",
	}))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_MONGODB_URI":                              "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                         "mhcat",
		"MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":                   "true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":             "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureWelcomeMessageDeliveryEnabled {
		t.Fatal("expected welcome-message delivery feature to be enabled explicitly")
	}
}

func TestLegacyWelcomeSpecialConfigMustBeComplete(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_MONGODB_URI":                     "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                "mhcat",
		"MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID": "guild",
	}))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                             "token",
		"MHCAT_MONGODB_URI":                               "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                          "mhcat",
		"MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID":           "guild",
		"MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID":             "bot",
		"MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID":         "channel",
		"MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID":    "chat",
		"MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID":    "help",
		"MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID":     "bug",
		"MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID": "support",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.LegacyWelcomeSpecialGuildID != "guild" ||
		cfg.LegacyWelcomeSpecialBotID != "bot" ||
		cfg.LegacyWelcomeSpecialChannelID != "channel" ||
		cfg.LegacyWelcomeSpecialChatChannelID != "chat" ||
		cfg.LegacyWelcomeSpecialHelpChannelID != "help" ||
		cfg.LegacyWelcomeSpecialBugChannelID != "bug" ||
		cfg.LegacyWelcomeSpecialSupportChannelID != "support" {
		t.Fatalf("special welcome config = %#v", cfg)
	}
}

func TestFeatureAccountAgePolicyRequiresGatewayAndGuildMembersIntent(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED": "true",
	}))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
	cfg, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_MONGODB_URI":                        "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":                   "mhcat",
		"MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED": "true",
		"MHCAT_DISCORD_ENABLE_GATEWAY":             "true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT":       "true",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeatureAccountAgePolicyEnabled {
		t.Fatal("expected account-age policy feature to be enabled explicitly")
	}
}

func mapLookup(values map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
