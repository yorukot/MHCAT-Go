package config

import (
	"errors"
	"testing"
)

func TestCommandSyncMissingApplicationIDFails(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":         "token",
		"MHCAT_COMMAND_SYNC_GUILD_ID": "guild",
	}))
	if err == nil {
		t.Fatal("expected missing application id error")
	}
	if !errors.Is(err, ErrInvalidCommandSyncConfig) {
		t.Fatalf("expected ErrInvalidCommandSyncConfig, got %v", err)
	}
}

func TestCommandSyncGuildScopeRequiresGuildID(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":          "token",
		"MHCAT_DISCORD_APPLICATION_ID": "app",
	}))
	if err == nil {
		t.Fatal("expected missing guild id error")
	}
}

func TestCommandSyncDefaultsDryRunStrict(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":          "token",
		"MHCAT_DISCORD_APPLICATION_ID": "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":  "guild",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.DryRun {
		t.Fatal("dry-run must default true")
	}
	if cfg.AllowDelete {
		t.Fatal("allow-delete must default false")
	}
	if cfg.AllowBulkOverwrite {
		t.Fatal("allow-bulk-overwrite must default false")
	}
	if !cfg.Strict {
		t.Fatal("strict must default true")
	}
	if cfg.IncludeTickets {
		t.Fatal("include tickets must default false")
	}
	if cfg.IncludePolls {
		t.Fatal("include polls must default false")
	}
	if cfg.IncludeEconomyQuery {
		t.Fatal("include economy query must default false")
	}
	if cfg.IncludeEconomySignIn {
		t.Fatal("include economy sign-in must default false")
	}
	if cfg.IncludeEconomySettings {
		t.Fatal("include economy settings must default false")
	}
	if cfg.IncludeEconomyCoinAdmin {
		t.Fatal("include economy coin-admin must default false")
	}
	if cfg.IncludeEconomyCoinRank {
		t.Fatal("include economy coin-rank must default false")
	}
	if cfg.IncludeEconomyRPS {
		t.Fatal("include economy RPS must default false")
	}
	if cfg.IncludeEconomyProfile {
		t.Fatal("include economy profile must default false")
	}
	if cfg.IncludeWork {
		t.Fatal("include work must default false")
	}
	if cfg.IncludeWarnings {
		t.Fatal("include warnings must default false")
	}
	if cfg.IncludeWarningSettings {
		t.Fatal("include warning settings must default false")
	}
	if cfg.IncludeWarningRemoval {
		t.Fatal("include warning removal must default false")
	}
	if cfg.IncludeWarningIssue {
		t.Fatal("include warning issue must default false")
	}
	if cfg.IncludeMessageCleanup {
		t.Fatal("include message cleanup must default false")
	}
	if cfg.IncludeDeleteData {
		t.Fatal("include delete data must default false")
	}
	if cfg.IncludeBalanceQuery {
		t.Fatal("include balance query must default false")
	}
	if cfg.IncludeRedeem {
		t.Fatal("include redeem must default false")
	}
	if cfg.IncludeAutoChatConfig {
		t.Fatal("include autochat config must default false")
	}
	if cfg.IncludeAutoNotificationConfig {
		t.Fatal("include auto-notification config must default false")
	}
	if cfg.IncludeAntiScamConfig {
		t.Fatal("include anti-scam config must default false")
	}
	if cfg.IncludeAntiScamReport {
		t.Fatal("include anti-scam report must default false")
	}
	if cfg.IncludeLoggingConfig {
		t.Fatal("include logging config must default false")
	}
	if cfg.IncludeGachaPrizeList {
		t.Fatal("include gacha prize-list must default false")
	}
	if cfg.IncludeGachaDraw {
		t.Fatal("include gacha draw must default false")
	}
	if cfg.IncludeGachaPrizeCreate {
		t.Fatal("include gacha prize-create must default false")
	}
	if cfg.IncludeGachaPrizeEdit {
		t.Fatal("include gacha prize-edit must default false")
	}
	if cfg.IncludeGachaPrizeDelete {
		t.Fatal("include gacha prize-delete must default false")
	}
	if cfg.IncludeLotteryDisabledCommand {
		t.Fatal("include lottery disabled command must default false")
	}
	if cfg.IncludeStatsQuery {
		t.Fatal("include stats query must default false")
	}
	if cfg.IncludeStatsCreate {
		t.Fatal("include stats create must default false")
	}
	if cfg.IncludeStatsRoleCount {
		t.Fatal("include stats role count must default false")
	}
	if cfg.IncludeStatsDelete {
		t.Fatal("include stats delete must default false")
	}
	if cfg.IncludeBirthdayConfig {
		t.Fatal("include birthday config must default false")
	}
	if cfg.IncludeAnnouncementConfig {
		t.Fatal("include announcement config must default false")
	}
	if cfg.IncludeAnnouncementSend {
		t.Fatal("include announcement send must default false")
	}
	if cfg.IncludeTextXPConfig {
		t.Fatal("include text XP config must default false")
	}
	if cfg.IncludeVoiceXPConfig {
		t.Fatal("include voice XP config must default false")
	}
	if cfg.IncludeXPRoleConfig {
		t.Fatal("include XP role config must default false")
	}
	if cfg.IncludeXPProfileDisabled {
		t.Fatal("include XP profile disabled commands must default false")
	}
	if cfg.IncludeXPAdmin {
		t.Fatal("include XP admin must default false")
	}
	if cfg.IncludeXPReset {
		t.Fatal("include XP reset must default false")
	}
	if cfg.IncludeVoiceRoomConfig {
		t.Fatal("include voice-room config must default false")
	}
	if cfg.IncludeVoiceRoomLock {
		t.Fatal("include voice-room lock must default false")
	}
	if cfg.IncludeJoinRoleConfig {
		t.Fatal("include join-role config must default false")
	}
	if cfg.IncludeWelcomeMessageConfig {
		t.Fatal("include welcome-message config must default false")
	}
	if cfg.IncludeVerificationConfig {
		t.Fatal("include verification config must default false")
	}
	if cfg.IncludeVerificationFlow {
		t.Fatal("include verification flow must default false")
	}
	if cfg.IncludeAccountAgeConfig {
		t.Fatal("include account-age config must default false")
	}
	if cfg.Scope != CommandSyncScopeGuild {
		t.Fatalf("expected guild scope, got %q", cfg.Scope)
	}
}

func TestCommandSyncIncludeTextXPConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include text XP config without staging mode to fail")
	}
}

func TestCommandSyncIncludeTextXPConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                  "global",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include text XP config with global scope to fail")
	}
}

func TestCommandSyncIncludeTextXPConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_STAGING_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeTextXPConfig {
		t.Fatal("expected include text XP config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeVoiceXPConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice XP config without staging mode to fail")
	}
}

func TestCommandSyncIncludeVoiceXPConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice XP config with global scope to fail")
	}
}

func TestCommandSyncIncludeVoiceXPConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeVoiceXPConfig {
		t.Fatal("expected include voice XP config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeXPProfileDisabledRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":                            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                             "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP profile disabled commands without staging mode to fail")
	}
}

func TestCommandSyncIncludeXPProfileDisabledRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":                            "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                                "global",
		"MHCAT_STAGING_MODE":                                      "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP profile disabled commands with global scope to fail")
	}
}

func TestCommandSyncIncludeXPProfileDisabledStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":                            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                             "guild",
		"MHCAT_STAGING_MODE":                                      "true",
		"MHCAT_STAGING_GUILD_ID":                                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeXPProfileDisabled {
		t.Fatal("expected include XP profile disabled commands to be enabled explicitly")
	}
}

func TestCommandSyncIncludeXPAdminRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP admin without staging mode to fail")
	}
}

func TestCommandSyncIncludeXPAdminRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_SCOPE":            "global",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP admin with global scope to fail")
	}
}

func TestCommandSyncIncludeXPAdminStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_STAGING_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeXPAdmin {
		t.Fatal("expected include XP admin to be enabled explicitly")
	}
}

func TestCommandSyncIncludeXPResetRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP reset without staging mode to fail")
	}
}

func TestCommandSyncIncludeXPResetRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_SCOPE":            "global",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP reset with global scope to fail")
	}
}

func TestCommandSyncIncludeXPResetStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_STAGING_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeXPReset {
		t.Fatal("expected include XP reset to be enabled explicitly")
	}
}

func TestCommandSyncIncludeVoiceRoomConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice-room config without staging mode to fail")
	}
}

func TestCommandSyncIncludeVoiceRoomConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                     "global",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice-room config with global scope to fail")
	}
}

func TestCommandSyncIncludeVoiceRoomConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_STAGING_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeVoiceRoomConfig {
		t.Fatal("expected include voice-room config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeVoiceRoomLockRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice-room lock without staging mode to fail")
	}
}

func TestCommandSyncIncludeVoiceRoomLockRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK": "true",
	}))
	if err == nil {
		t.Fatal("expected include voice-room lock with global scope to fail")
	}
}

func TestCommandSyncIncludeVoiceRoomLockStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeVoiceRoomLock {
		t.Fatal("expected include voice-room lock to be enabled explicitly")
	}
}

func TestCommandSyncIncludeJoinRoleConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include join-role config without staging mode to fail")
	}
}

func TestCommandSyncIncludeJoinRoleConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include join-role config with global scope to fail")
	}
}

func TestCommandSyncIncludeJoinRoleConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeJoinRoleConfig {
		t.Fatal("expected include join-role config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeGachaPrizeListRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-list without staging mode to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeListRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-list with global scope to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeListStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeGachaPrizeList {
		t.Fatal("expected include gacha prize-list to be enabled explicitly")
	}
}

func TestCommandSyncIncludeGachaDrawRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_DISCORD_APPLICATION_ID":          "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":           "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha draw without staging mode to fail")
	}
}

func TestCommandSyncIncludeGachaDrawRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_DISCORD_APPLICATION_ID":          "app",
		"MHCAT_COMMAND_SYNC_SCOPE":              "global",
		"MHCAT_STAGING_MODE":                    "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha draw with global scope to fail")
	}
}

func TestCommandSyncIncludeGachaDrawStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                   "token",
		"MHCAT_DISCORD_APPLICATION_ID":          "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":           "guild",
		"MHCAT_STAGING_MODE":                    "true",
		"MHCAT_STAGING_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeGachaDraw {
		t.Fatal("expected include gacha draw to be enabled explicitly")
	}
}

func TestCommandSyncIncludeGachaPrizeDeleteRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-delete without staging mode to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeDeleteRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                      "global",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-delete with global scope to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeDeleteStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_STAGING_GUILD_ID":                        "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeGachaPrizeDelete {
		t.Fatal("expected include gacha prize-delete to be enabled explicitly")
	}
}

func TestCommandSyncIncludeGachaPrizeCreateRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-create without staging mode to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeCreateRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                      "global",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-create with global scope to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeCreateStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_STAGING_GUILD_ID":                        "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeGachaPrizeCreate {
		t.Fatal("expected include gacha prize-create to be enabled explicitly")
	}
}

func TestCommandSyncIncludeGachaPrizeEditRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-edit without staging mode to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeEditRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT": "true",
	}))
	if err == nil {
		t.Fatal("expected include gacha prize-edit with global scope to fail")
	}
}

func TestCommandSyncIncludeGachaPrizeEditStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeGachaPrizeEdit {
		t.Fatal("expected include gacha prize-edit to be enabled explicitly")
	}
}

func TestCommandSyncIncludeLotteryDisabledCommandRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND": "true",
	}))
	if err == nil {
		t.Fatal("expected include lottery disabled command without staging mode to fail")
	}
}

func TestCommandSyncIncludeLotteryDisabledCommandRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                            "global",
		"MHCAT_STAGING_MODE":                                  "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND": "true",
	}))
	if err == nil {
		t.Fatal("expected include lottery disabled command with global scope to fail")
	}
}

func TestCommandSyncIncludeLotteryDisabledCommandStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                         "guild",
		"MHCAT_STAGING_MODE":                                  "true",
		"MHCAT_STAGING_GUILD_ID":                              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeLotteryDisabledCommand {
		t.Fatal("expected include lottery disabled command to be enabled explicitly")
	}
}

func TestCommandSyncIncludeStatsQueryRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats query without staging mode to fail")
	}
}

func TestCommandSyncIncludeStatsQueryRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_SCOPE":               "global",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats query with global scope to fail")
	}
}

func TestCommandSyncIncludeStatsQueryStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_STAGING_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeStatsQuery {
		t.Fatal("expected include stats query to be enabled explicitly")
	}
}

func TestCommandSyncIncludeStatsDeleteRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":             "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats delete without staging mode to fail")
	}
}

func TestCommandSyncIncludeStatsDeleteRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                "global",
		"MHCAT_STAGING_MODE":                      "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats delete with global scope to fail")
	}
}

func TestCommandSyncIncludeStatsDeleteStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":             "guild",
		"MHCAT_STAGING_MODE":                      "true",
		"MHCAT_STAGING_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeStatsDelete {
		t.Fatal("expected include stats delete to be enabled explicitly")
	}
}

func TestCommandSyncIncludeStatsCreateRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":             "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats create without staging mode to fail")
	}
}

func TestCommandSyncIncludeStatsCreateRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                "global",
		"MHCAT_STAGING_MODE":                      "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats create with global scope to fail")
	}
}

func TestCommandSyncIncludeStatsCreateStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                     "token",
		"MHCAT_DISCORD_APPLICATION_ID":            "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":             "guild",
		"MHCAT_STAGING_MODE":                      "true",
		"MHCAT_STAGING_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeStatsCreate {
		t.Fatal("expected include stats create to be enabled explicitly")
	}
}

func TestCommandSyncIncludeStatsRoleCountRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats role-count without staging mode to fail")
	}
}

func TestCommandSyncIncludeStatsRoleCountRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT": "true",
	}))
	if err == nil {
		t.Fatal("expected include stats role-count with global scope to fail")
	}
}

func TestCommandSyncIncludeStatsRoleCountStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeStatsRoleCount {
		t.Fatal("expected include stats role-count to be enabled explicitly")
	}
}

func TestCommandSyncIncludeXPRoleConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP role config without staging mode to fail")
	}
}

func TestCommandSyncIncludeXPRoleConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                  "global",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include XP role config with global scope to fail")
	}
}

func TestCommandSyncIncludeXPRoleConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_STAGING_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeXPRoleConfig {
		t.Fatal("expected include XP role config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeRedeemRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_DISCORD_APPLICATION_ID":      "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_REDEEM": "true",
	}))
	if err == nil {
		t.Fatal("expected include redeem without staging mode to fail")
	}
}

func TestCommandSyncIncludeRedeemRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_DISCORD_APPLICATION_ID":      "app",
		"MHCAT_COMMAND_SYNC_SCOPE":          "global",
		"MHCAT_STAGING_MODE":                "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_REDEEM": "true",
	}))
	if err == nil {
		t.Fatal("expected include redeem with global scope to fail")
	}
}

func TestCommandSyncIncludeRedeemStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_DISCORD_APPLICATION_ID":      "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":       "guild",
		"MHCAT_STAGING_MODE":                "true",
		"MHCAT_STAGING_GUILD_ID":            "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_REDEEM": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeRedeem {
		t.Fatal("expected include redeem to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAnnouncementConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include announcement config without staging mode to fail")
	}
}

func TestCommandSyncIncludeBirthdayConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include birthday config without staging mode to fail")
	}
}

func TestCommandSyncIncludeBirthdayConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include birthday config with global scope to fail")
	}
}

func TestCommandSyncIncludeBirthdayConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeBirthdayConfig {
		t.Fatal("expected include birthday config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAnnouncementConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                       "global",
		"MHCAT_STAGING_MODE":                             "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include announcement config with global scope to fail")
	}
}

func TestCommandSyncIncludeAnnouncementConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                    "guild",
		"MHCAT_STAGING_MODE":                             "true",
		"MHCAT_STAGING_GUILD_ID":                         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAnnouncementConfig {
		t.Fatal("expected include announcement config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAnnouncementSendRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND": "true",
	}))
	if err == nil {
		t.Fatal("expected include announcement send without staging mode to fail")
	}
}

func TestCommandSyncIncludeAnnouncementSendRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                     "global",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND": "true",
	}))
	if err == nil {
		t.Fatal("expected include announcement send with global scope to fail")
	}
}

func TestCommandSyncIncludeAnnouncementSendStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_STAGING_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAnnouncementSend {
		t.Fatal("expected include announcement send to be enabled explicitly")
	}
}

func TestCommandSyncIncludeLoggingConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include logging config without staging mode to fail")
	}
}

func TestCommandSyncIncludeLoggingConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                  "global",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include logging config with global scope to fail")
	}
}

func TestCommandSyncIncludeLoggingConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_STAGING_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeLoggingConfig {
		t.Fatal("expected include logging config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeTicketsRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                "token",
		"MHCAT_DISCORD_APPLICATION_ID":       "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":        "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TICKETS": "true",
	}))
	if err == nil {
		t.Fatal("expected include tickets without staging mode to fail")
	}
}

func TestCommandSyncIncludeTicketsRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                "token",
		"MHCAT_DISCORD_APPLICATION_ID":       "app",
		"MHCAT_COMMAND_SYNC_SCOPE":           "global",
		"MHCAT_STAGING_MODE":                 "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_TICKETS": "true",
	}))
	if err == nil {
		t.Fatal("expected include tickets with global scope to fail")
	}
}

func TestCommandSyncIncludeTicketsStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                "token",
		"MHCAT_DISCORD_APPLICATION_ID":       "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":        "guild",
		"MHCAT_STAGING_MODE":                 "true",
		"MHCAT_STAGING_GUILD_ID":             "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TICKETS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeTickets {
		t.Fatal("expected include tickets to be enabled explicitly")
	}
}

func TestCommandSyncIncludePollsRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_DISCORD_APPLICATION_ID":     "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_POLLS": "true",
	}))
	if err == nil {
		t.Fatal("expected include polls without staging mode to fail")
	}
}

func TestCommandSyncIncludePollsRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_DISCORD_APPLICATION_ID":     "app",
		"MHCAT_COMMAND_SYNC_SCOPE":         "global",
		"MHCAT_STAGING_MODE":               "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_POLLS": "true",
	}))
	if err == nil {
		t.Fatal("expected include polls with global scope to fail")
	}
}

func TestCommandSyncIncludePollsStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_DISCORD_APPLICATION_ID":     "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":      "guild",
		"MHCAT_STAGING_MODE":               "true",
		"MHCAT_STAGING_GUILD_ID":           "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_POLLS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludePolls {
		t.Fatal("expected include polls to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomyQueryRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy query without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomyQueryRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                 "global",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy query with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomyQueryStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_STAGING_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomyQuery {
		t.Fatal("expected include economy query to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomySignInRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy sign-in without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomySignInRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                  "global",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy sign-in with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomySignInStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                       "token",
		"MHCAT_DISCORD_APPLICATION_ID":              "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":               "guild",
		"MHCAT_STAGING_MODE":                        "true",
		"MHCAT_STAGING_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomySignIn {
		t.Fatal("expected include economy sign-in to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomySettingsRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy settings without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomySettingsRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy settings with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomySettingsStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomySettings {
		t.Fatal("expected include economy settings to be enabled explicitly")
	}
}

func TestCommandSyncIncludeWorkRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":             "token",
		"MHCAT_DISCORD_APPLICATION_ID":    "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WORK": "true",
	}))
	if err == nil {
		t.Fatal("expected include work without staging mode to fail")
	}
}

func TestCommandSyncIncludeWorkRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":             "token",
		"MHCAT_DISCORD_APPLICATION_ID":    "app",
		"MHCAT_COMMAND_SYNC_SCOPE":        "global",
		"MHCAT_STAGING_MODE":              "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WORK": "true",
	}))
	if err == nil {
		t.Fatal("expected include work with global scope to fail")
	}
}

func TestCommandSyncIncludeWorkStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":             "token",
		"MHCAT_DISCORD_APPLICATION_ID":    "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":     "guild",
		"MHCAT_STAGING_MODE":              "true",
		"MHCAT_STAGING_GUILD_ID":          "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WORK": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWork {
		t.Fatal("expected include work to be enabled explicitly")
	}
}

func TestCommandSyncIncludeWarningsRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include warnings without staging mode to fail")
	}
}

func TestCommandSyncIncludeWarningsRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_SCOPE":            "global",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include warnings with global scope to fail")
	}
}

func TestCommandSyncIncludeWarningsStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":         "guild",
		"MHCAT_STAGING_MODE":                  "true",
		"MHCAT_STAGING_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWarnings {
		t.Fatal("expected include warnings to be enabled explicitly")
	}
}

func TestCommandSyncIncludeWarningSettingsRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning settings without staging mode to fail")
	}
}

func TestCommandSyncIncludeWarningSettingsRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning settings with global scope to fail")
	}
}

func TestCommandSyncIncludeWarningSettingsStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWarningSettings {
		t.Fatal("expected include warning settings to be enabled explicitly")
	}
}

func TestCommandSyncIncludeWarningRemovalRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning removal without staging mode to fail")
	}
}

func TestCommandSyncIncludeWarningRemovalRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning removal with global scope to fail")
	}
}

func TestCommandSyncIncludeWarningRemovalStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWarningRemoval {
		t.Fatal("expected include warning removal to be enabled explicitly")
	}
}

func TestCommandSyncIncludeWarningIssueRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning issue without staging mode to fail")
	}
}

func TestCommandSyncIncludeWarningIssueRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                 "global",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE": "true",
	}))
	if err == nil {
		t.Fatal("expected include warning issue with global scope to fail")
	}
}

func TestCommandSyncIncludeWarningIssueStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_STAGING_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWarningIssue {
		t.Fatal("expected include warning issue to be enabled explicitly")
	}
}

func TestCommandSyncIncludeMessageCleanupRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP": "true",
	}))
	if err == nil {
		t.Fatal("expected include message cleanup without staging mode to fail")
	}
}

func TestCommandSyncIncludeMessageCleanupRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP": "true",
	}))
	if err == nil {
		t.Fatal("expected include message cleanup with global scope to fail")
	}
}

func TestCommandSyncIncludeMessageCleanupStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeMessageCleanup {
		t.Fatal("expected include message cleanup to be enabled explicitly")
	}
}

func TestCommandSyncIncludeDeleteDataRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA": "true",
	}))
	if err == nil {
		t.Fatal("expected include delete data without staging mode to fail")
	}
}

func TestCommandSyncIncludeDeleteDataRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_SCOPE":               "global",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA": "true",
	}))
	if err == nil {
		t.Fatal("expected include delete data with global scope to fail")
	}
}

func TestCommandSyncIncludeDeleteDataStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_STAGING_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeDeleteData {
		t.Fatal("expected include delete data to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomyCoinAdminRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy coin-admin without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomyCoinAdminRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                      "global",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy coin-admin with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomyCoinAdminStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_STAGING_GUILD_ID":                        "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomyCoinAdmin {
		t.Fatal("expected include economy coin-admin to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomyCoinRankRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy coin-rank without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomyCoinRankRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                     "global",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy coin-rank with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomyCoinRankStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_STAGING_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomyCoinRank {
		t.Fatal("expected include economy coin-rank to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomyRPSRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy RPS without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomyRPSRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_SCOPE":               "global",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy RPS with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomyRPSStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                    "token",
		"MHCAT_DISCORD_APPLICATION_ID":           "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":            "guild",
		"MHCAT_STAGING_MODE":                     "true",
		"MHCAT_STAGING_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomyRPS {
		t.Fatal("expected include economy RPS to be enabled explicitly")
	}
}

func TestCommandSyncIncludeEconomyProfileRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy profile without staging mode to fail")
	}
}

func TestCommandSyncIncludeEconomyProfileRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE": "true",
	}))
	if err == nil {
		t.Fatal("expected include economy profile with global scope to fail")
	}
}

func TestCommandSyncIncludeEconomyProfileStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeEconomyProfile {
		t.Fatal("expected include economy profile to be enabled explicitly")
	}
}

func TestCommandSyncIncludeTranslateRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_DISCORD_APPLICATION_ID":         "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":          "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include translate without staging mode to fail")
	}
}

func TestCommandSyncIncludeTranslateRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_DISCORD_APPLICATION_ID":         "app",
		"MHCAT_COMMAND_SYNC_SCOPE":             "global",
		"MHCAT_STAGING_MODE":                   "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE": "true",
	}))
	if err == nil {
		t.Fatal("expected include translate with global scope to fail")
	}
}

func TestCommandSyncIncludeTranslateStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "token",
		"MHCAT_DISCORD_APPLICATION_ID":         "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":          "guild",
		"MHCAT_STAGING_MODE":                   "true",
		"MHCAT_STAGING_GUILD_ID":               "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeTranslate {
		t.Fatal("expected include translate to be enabled explicitly")
	}
}

func TestCommandSyncIncludeBalanceQueryRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include balance query without staging mode to fail")
	}
}

func TestCommandSyncIncludeBalanceQueryRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                 "global",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY": "true",
	}))
	if err == nil {
		t.Fatal("expected include balance query with global scope to fail")
	}
}

func TestCommandSyncIncludeBalanceQueryStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                      "token",
		"MHCAT_DISCORD_APPLICATION_ID":             "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":              "guild",
		"MHCAT_STAGING_MODE":                       "true",
		"MHCAT_STAGING_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeBalanceQuery {
		t.Fatal("expected include balance query to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAutoChatConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include autochat config without staging mode to fail")
	}
}

func TestCommandSyncIncludeAutoChatConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                   "global",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include autochat config with global scope to fail")
	}
}

func TestCommandSyncIncludeAutoChatConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                        "token",
		"MHCAT_DISCORD_APPLICATION_ID":               "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                "guild",
		"MHCAT_STAGING_MODE":                         "true",
		"MHCAT_STAGING_GUILD_ID":                     "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAutoChatConfig {
		t.Fatal("expected include autochat config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAutoNotificationConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include auto-notification config without staging mode to fail")
	}
}

func TestCommandSyncIncludeAutoNotificationConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                            "global",
		"MHCAT_STAGING_MODE":                                  "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include auto-notification config with global scope to fail")
	}
}

func TestCommandSyncIncludeAutoNotificationConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                                 "token",
		"MHCAT_DISCORD_APPLICATION_ID":                        "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                         "guild",
		"MHCAT_STAGING_MODE":                                  "true",
		"MHCAT_STAGING_GUILD_ID":                              "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAutoNotificationConfig {
		t.Fatal("expected include auto-notification config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAntiScamConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include anti-scam config without staging mode to fail")
	}
}

func TestCommandSyncIncludeAntiScamConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include anti-scam config with global scope to fail")
	}
}

func TestCommandSyncIncludeAntiScamConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAntiScamConfig {
		t.Fatal("expected include anti-scam config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAntiScamReportRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT": "true",
	}))
	if err == nil {
		t.Fatal("expected include anti-scam report without staging mode to fail")
	}
}

func TestCommandSyncIncludeAntiScamReportRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT": "true",
	}))
	if err == nil {
		t.Fatal("expected include anti-scam report with global scope to fail")
	}
}

func TestCommandSyncIncludeAntiScamReportStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAntiScamReport {
		t.Fatal("expected include anti-scam report to be enabled explicitly")
	}
}

func TestCommandSyncGlobalRejectsGuildID(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":          "token",
		"MHCAT_DISCORD_APPLICATION_ID": "app",
		"MHCAT_COMMAND_SYNC_SCOPE":     "global",
		"MHCAT_COMMAND_SYNC_GUILD_ID":  "guild",
	}))
	if err == nil {
		t.Fatal("expected global scope with guild id to fail")
	}
}

func TestCommandSyncDeleteRequiresApplyMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":             "token",
		"MHCAT_DISCORD_APPLICATION_ID":    "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":     "guild",
		"MHCAT_COMMAND_SYNC_ALLOW_DELETE": "true",
	}))
	if err == nil {
		t.Fatal("expected allow-delete dry-run validation error")
	}
}

func TestCommandSyncLegacyTokenAlias(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"TOKEN":                        "legacy-token",
		"MHCAT_DISCORD_APPLICATION_ID": "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":  "guild",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if cfg.DiscordToken != "legacy-token" {
		t.Fatalf("expected legacy token, got %q", cfg.DiscordToken)
	}
}

func TestCommandSyncIncludeWelcomeMessageConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                               "token",
		"MHCAT_DISCORD_APPLICATION_ID":                      "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include welcome-message config without staging mode to fail")
	}
}

func TestCommandSyncIncludeWelcomeMessageConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                               "token",
		"MHCAT_DISCORD_APPLICATION_ID":                      "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                          "global",
		"MHCAT_STAGING_MODE":                                "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include welcome-message config with global scope to fail")
	}
}

func TestCommandSyncIncludeWelcomeMessageConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                               "token",
		"MHCAT_DISCORD_APPLICATION_ID":                      "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                       "guild",
		"MHCAT_STAGING_MODE":                                "true",
		"MHCAT_STAGING_GUILD_ID":                            "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeWelcomeMessageConfig {
		t.Fatal("expected include welcome-message config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeVerificationConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                    "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include verification config without staging mode to fail")
	}
}

func TestCommandSyncIncludeVerificationConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                       "global",
		"MHCAT_STAGING_MODE":                             "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include verification config with global scope to fail")
	}
}

func TestCommandSyncIncludeVerificationConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                            "token",
		"MHCAT_DISCORD_APPLICATION_ID":                   "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                    "guild",
		"MHCAT_STAGING_MODE":                             "true",
		"MHCAT_STAGING_GUILD_ID":                         "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeVerificationConfig {
		t.Fatal("expected include verification config to be enabled explicitly")
	}
}

func TestCommandSyncIncludeVerificationFlowRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW": "true",
	}))
	if err == nil {
		t.Fatal("expected include verification flow without staging mode to fail")
	}
}

func TestCommandSyncIncludeVerificationFlowRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                     "global",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_STAGING_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW": "true",
	}))
	if err == nil {
		t.Fatal("expected include verification flow with global scope to fail")
	}
}

func TestCommandSyncIncludeVerificationFlowStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                          "token",
		"MHCAT_DISCORD_APPLICATION_ID":                 "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                  "guild",
		"MHCAT_STAGING_MODE":                           "true",
		"MHCAT_STAGING_GUILD_ID":                       "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeVerificationFlow {
		t.Fatal("expected include verification flow to be enabled explicitly")
	}
}

func TestCommandSyncIncludeAccountAgeConfigRequiresStagingMode(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include account-age config without staging mode to fail")
	}
}

func TestCommandSyncIncludeAccountAgeConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                      "global",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include account-age config with global scope to fail")
	}
}

func TestCommandSyncIncludeAccountAgeConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                           "token",
		"MHCAT_DISCORD_APPLICATION_ID":                  "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                   "guild",
		"MHCAT_STAGING_MODE":                            "true",
		"MHCAT_STAGING_GUILD_ID":                        "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG": "true",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if !cfg.IncludeAccountAgeConfig {
		t.Fatal("expected include account-age config to be enabled explicitly")
	}
}
