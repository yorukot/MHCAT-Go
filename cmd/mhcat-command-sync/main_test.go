package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestCommandSyncMissingApplicationIDFails(t *testing.T) {
	env := baseCommandSyncEnv()
	delete(env, "MHCAT_DISCORD_APPLICATION_ID")
	exitCode, stdout, stderr, _ := runWithFake(t, nil, env, nil)
	if exitCode == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "MHCAT_DISCORD_APPLICATION_ID") {
		t.Fatalf("expected missing application id error, stderr=%q stdout=%q", stderr, stdout)
	}
}

func TestCommandSyncGuildScopeMissingGuildIDFails(t *testing.T) {
	env := baseCommandSyncEnv()
	delete(env, "MHCAT_COMMAND_SYNC_GUILD_ID")
	exitCode, _, stderr, _ := runWithFake(t, nil, env, nil)
	if exitCode == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "MHCAT_COMMAND_SYNC_GUILD_ID") {
		t.Fatalf("expected missing guild id error, stderr=%q", stderr)
	}
}

func TestCommandSyncDryRunPerformsNoWrites(t *testing.T) {
	registry := commands.NewRegistry(scope(), []commands.Definition{definition("ping", "Ping")})
	exitCode, _, stderr, client := runWithFake(t, nil, baseCommandSyncEnv(), func(config.CommandSyncConfig, commands.Scope) commands.Registry {
		return registry
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if len(client.Created) != 0 || len(client.Updated) != 0 || len(client.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", client)
	}
	if !strings.Contains(stderr, "dry-run") {
		t.Fatalf("expected dry-run notice, stderr=%q", stderr)
	}
}

func TestCommandSyncApplyRequiresStagingMode(t *testing.T) {
	fake := &fakediscord.CommandSyncClient{}
	exitCode, _, stderr, _ := runWithFakeClient(t, []string{"--apply"}, baseCommandSyncEnv(), fake, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected apply without staging to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe apply performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncStagingApplyCreatesAndUpdatesManagedCommands(t *testing.T) {
	fake := &fakediscord.CommandSyncClient{
		Remote: []commands.RemoteCommand{{ID: "remote-ping", Definition: definition("ping", "Old"), Owned: true}},
	}
	exitCode, _, stderr, client := runWithFakeClient(t, []string{"--apply"}, stagingCommandSyncEnv(), fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if client != fake {
		t.Fatal("expected injected fake client")
	}
	if len(fake.Created) != 2 {
		t.Fatalf("created = %#v", fake.Created)
	}
	if len(fake.Updated) != 1 || fake.Updated[0].RemoteID != "remote-ping" {
		t.Fatalf("updated = %#v", fake.Updated)
	}
}

func TestCommandSyncApplyDoesNotDeleteWithoutAllowDelete(t *testing.T) {
	fake := &fakediscord.CommandSyncClient{
		Remote: []commands.RemoteCommand{{ID: "remote-old", Definition: definition("old", "Old"), Owned: true}},
	}
	exitCode, _, stderr, _ := runWithFakeClient(t, []string{"--apply"}, stagingCommandSyncEnv(), fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	if len(fake.Deleted) != 0 {
		t.Fatalf("delete happened without allow-delete: %#v", fake.Deleted)
	}
}

func TestCommandSyncStagingApplyRejectsDeleteFlag(t *testing.T) {
	fake := &fakediscord.CommandSyncClient{}
	exitCode, _, stderr, _ := runWithFakeClient(t, []string{"--apply", "--allow-delete"}, stagingCommandSyncEnv(), fake, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected delete flag to fail in staging apply")
	}
	if len(fake.Deleted) != 0 {
		t.Fatalf("delete happened: %#v", fake.Deleted)
	}
	if !strings.Contains(stderr, "deletion") {
		t.Fatalf("expected deletion rejection, stderr=%q", stderr)
	}
}

func TestCommandSyncDoesNotPrintRawToken(t *testing.T) {
	env := baseCommandSyncEnv()
	primaryToken := strings.Repeat("a", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 32)
	aliasToken := strings.Repeat("d", 24) + "." + strings.Repeat("e", 6) + "." + strings.Repeat("f", 32)
	env["MHCAT_DISCORD_TOKEN"] = primaryToken
	env["TOKEN"] = aliasToken
	exitCode, stdout, stderr, _ := runWithFake(t, nil, env, nil)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, stderr=%q", stderr)
	}
	combined := stdout + stderr
	for _, raw := range []string{env["MHCAT_DISCORD_TOKEN"], env["TOKEN"]} {
		if strings.Contains(combined, raw) {
			t.Fatalf("raw token appeared in output: %q", combined)
		}
	}
}

func TestCommandSyncIncludeTicketsRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TICKETS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include tickets without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include tickets performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeTicketsStagingDryRunIncludesTicketDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TICKETS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected ticket dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"私人頻道設置", "私人頻道刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("ticket command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludePollsRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_POLLS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include polls without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include polls performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludePollsStagingDryRunIncludesPollDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_POLLS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected poll dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "投票創建") {
		t.Fatalf("poll command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeLoggingConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include logging config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include logging config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeLoggingConfigStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected logging config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "set-log-channel") {
		t.Fatalf("logging config command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeAntiScamReportStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected anti-scam report dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "詐騙網址回報") {
		t.Fatalf("anti-scam report command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeGachaPrizeListRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include gacha prize-list without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include gacha prize-list performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeGachaPrizeListStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected gacha prize-list dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "扭蛋獎池查詢") {
		t.Fatalf("gacha prize-list command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeLotteryDisabledCommandRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include lottery disabled command without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include lottery disabled command performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeLotteryDisabledCommandStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected lottery disabled command dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "抽獎設置") {
		t.Fatalf("lottery disabled command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeStatsQueryRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include stats query without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include stats query performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeStatsQueryStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected stats query dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "統計系統查詢") {
		t.Fatalf("stats query command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeAnnouncementConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include announcement config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include announcement config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAnnouncementConfigStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected announcement config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "公告頻道設置") {
		t.Fatalf("announcement config command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeAnnouncementSendRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include announcement send without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include announcement send performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAnnouncementSendStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected announcement send dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "公告發送") {
		t.Fatalf("announcement send command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeTextXPConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include text XP config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include text XP config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeTextXPConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected text XP config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"聊天經驗設定", "聊天經驗刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("text XP config command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeVoiceXPConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include voice XP config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include voice XP config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeVoiceXPConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected voice XP config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"語音經驗設定", "語音經驗刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("voice XP config command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeXPProfileDisabledRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include XP profile disabled commands without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include XP profile disabled commands performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeXPProfileDisabledStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected XP profile disabled dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"聊天經驗", "語音經驗"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("XP profile disabled command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeVoiceRoomConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include voice-room config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include voice-room config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeVoiceRoomConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected voice-room config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"語音包廂設置", "語音包廂刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("voice-room config command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeJoinRoleConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include join-role config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include join-role config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeJoinRoleConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected join-role config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"加入身份組設置", "加入身份組刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("join-role config command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeWelcomeMessageConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include welcome-message config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include welcome-message config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWelcomeMessageConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected welcome-message config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"加入訊息設置", "退出訊息設置"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("welcome-message config command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeVerificationConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include verification config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include verification config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeVerificationConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected verification config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "驗證設置") {
		t.Fatalf("verification config command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeVerificationFlowRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include verification flow without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include verification flow performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeVerificationFlowStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected verification flow dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "驗證") {
		t.Fatalf("verification command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeEconomyQueryRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include economy query without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include economy query performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeEconomyQueryStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected economy query dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "代幣查詢") {
		t.Fatalf("economy query command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeEconomySignInRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include economy sign-in without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include economy sign-in performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeEconomySignInStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected economy sign-in dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"簽到", "簽到列表"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("economy sign-in command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeEconomySettingsRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include economy settings without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include economy settings performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeEconomySettingsStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected economy settings dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "coin-related-settings") {
		t.Fatalf("economy settings command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeEconomyCoinAdminRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include economy coin-admin without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include economy coin-admin performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeEconomyCoinAdminStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected economy coin-admin dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "代幣增加") {
		t.Fatalf("economy coin-admin command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeEconomyCoinRankRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include economy coin-rank without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include economy coin-rank performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeEconomyCoinRankStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected economy coin-rank dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "代幣排行榜") {
		t.Fatalf("economy coin-rank command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeWorkRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WORK"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include work without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include work performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWorkStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WORK"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected work dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "打工系統") {
		t.Fatalf("work command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeWarningsRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include warnings without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include warnings performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWarningsStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected warnings dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "警告紀錄") {
		t.Fatalf("warnings command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeWarningSettingsRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include warning settings without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include warning settings performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWarningSettingsStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected warning settings dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "警告設定") {
		t.Fatalf("warning settings command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeWarningRemovalRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include warning removal without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include warning removal performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWarningRemovalStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected warning removal dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "警告清除") || !strings.Contains(stdout, "警告全部清除") {
		t.Fatalf("warning removal commands missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeWarningIssueRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include warning issue without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include warning issue performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeWarningIssueStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected warning issue dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "name=警告 ") {
		t.Fatalf("warning issue command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeMessageCleanupRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include message cleanup without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include message cleanup performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeMessageCleanupStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected message cleanup dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "刪除訊息") {
		t.Fatalf("message cleanup command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeDeleteDataRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include delete data without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include delete data performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeDeleteDataStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected delete data dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "刪除資料") {
		t.Fatalf("delete data command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeTranslateRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include translate without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include translate performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeTranslateStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected translate dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "翻譯") {
		t.Fatalf("translate command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeBalanceQueryRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include balance query without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include balance query performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeBalanceQueryStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected balance query dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "查看餘額") {
		t.Fatalf("balance query command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeRedeemRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_REDEEM"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include redeem without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include redeem performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeRedeemStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_REDEEM"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected redeem dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "兌換") {
		t.Fatalf("redeem command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeAutoChatConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include autochat config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include autochat config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAutoChatConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected autochat config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	for _, name := range []string{"自動聊天頻道", "自動聊天頻道刪除"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("autochat command %q missing from dry-run output: %q", name, stdout)
		}
	}
}

func TestCommandSyncIncludeAutoNotificationConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include auto-notification config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include auto-notification config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAutoNotificationConfigStagingDryRunIncludesDefinitions(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected auto-notification config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "自動通知列表") || !strings.Contains(stdout, "自動通知刪除") {
		t.Fatalf("expected auto-notification definitions in stdout=%q", stdout)
	}
}

func TestCommandSyncIncludeAntiScamConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include anti-scam config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include anti-scam config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAntiScamConfigStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected anti-scam config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "防詐騙網址") {
		t.Fatalf("anti-scam config command missing from dry-run output: %q", stdout)
	}
}

func TestCommandSyncIncludeAccountAgeConfigRequiresStagingMode(t *testing.T) {
	env := baseCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG"] = "true"
	exitCode, _, stderr, fake := runWithFakeClient(t, nil, env, &fakediscord.CommandSyncClient{}, defaultCommandRegistry)
	if exitCode == 0 {
		t.Fatal("expected include account-age config without staging mode to fail")
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("unsafe include account-age config performed writes: %#v", fake)
	}
	if !strings.Contains(stderr, "MHCAT_STAGING_MODE") {
		t.Fatalf("expected staging mode error, stderr=%q", stderr)
	}
}

func TestCommandSyncIncludeAccountAgeConfigStagingDryRunIncludesDefinition(t *testing.T) {
	env := stagingCommandSyncEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG"] = "true"
	fake := &fakediscord.CommandSyncClient{}
	exitCode, stdout, stderr, _ := runWithFakeClient(t, nil, env, fake, defaultCommandRegistry)
	if exitCode != 0 {
		t.Fatalf("expected account-age config dry-run to pass, stderr=%q stdout=%q", stderr, stdout)
	}
	if len(fake.Created) != 0 || len(fake.Updated) != 0 || len(fake.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", fake)
	}
	if !strings.Contains(stdout, "帳號需創建時數") {
		t.Fatalf("account-age command missing from dry-run output: %q", stdout)
	}
}

func runWithFake(t *testing.T, args []string, env map[string]string, loader registryLoader) (int, string, string, *fakediscord.CommandSyncClient) {
	t.Helper()
	return runWithFakeClient(t, args, env, &fakediscord.CommandSyncClient{}, loader)
}

func runWithFakeClient(t *testing.T, args []string, env map[string]string, client *fakediscord.CommandSyncClient, loader registryLoader) (int, string, string, *fakediscord.CommandSyncClient) {
	t.Helper()
	if loader == nil {
		loader = defaultCommandRegistry
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		context.Background(),
		args,
		mapLookup(env),
		&stdout,
		&stderr,
		func(config.CommandSyncConfig) (commands.SyncClient, error) {
			return client, nil
		},
		loader,
	)
	return exitCode, stdout.String(), stderr.String(), client
}

func baseCommandSyncEnv() map[string]string {
	return map[string]string{
		"MHCAT_DISCORD_TOKEN":          "test-token",
		"MHCAT_DISCORD_APPLICATION_ID": "app-1",
		"MHCAT_COMMAND_SYNC_SCOPE":     "guild",
		"MHCAT_COMMAND_SYNC_GUILD_ID":  "guild-1",
	}
}

func stagingCommandSyncEnv() map[string]string {
	env := baseCommandSyncEnv()
	env["MHCAT_STAGING_MODE"] = "true"
	env["MHCAT_STAGING_GUILD_ID"] = "guild-1"
	env["MHCAT_STAGING_ALLOWED_APPLICATION_ID"] = "app-1"
	env["MHCAT_STAGING_ALLOW_COMMAND_APPLY"] = "true"
	return env
}

func mapLookup(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func scope() commands.Scope {
	return commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
}

func definition(name, description string) commands.Definition {
	return commands.Definition{Type: commands.CommandTypeChatInput, Name: name, Description: description}
}
