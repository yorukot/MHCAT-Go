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
	if cfg.IncludeWork {
		t.Fatal("include work must default false")
	}
	if cfg.IncludeLoggingConfig {
		t.Fatal("include logging config must default false")
	}
	if cfg.IncludeGachaPrizeList {
		t.Fatal("include gacha prize-list must default false")
	}
	if cfg.IncludeLotteryDisabledCommand {
		t.Fatal("include lottery disabled command must default false")
	}
	if cfg.IncludeStatsQuery {
		t.Fatal("include stats query must default false")
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
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include birthday config without staging mode to fail")
	}
}

func TestCommandSyncIncludeBirthdayConfigRequiresGuildScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_SCOPE":                    "global",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG": "true",
	}))
	if err == nil {
		t.Fatal("expected include birthday config with global scope to fail")
	}
}

func TestCommandSyncIncludeBirthdayConfigStagingGuildParses(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                         "token",
		"MHCAT_DISCORD_APPLICATION_ID":                "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":                 "guild",
		"MHCAT_STAGING_MODE":                          "true",
		"MHCAT_STAGING_GUILD_ID":                      "guild",
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
