package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPreflightMissingEnvFails(t *testing.T) {
	code, stdout, stderr := runPreflight(t, nil, nil)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "discord-token status=fail") {
		t.Fatalf("expected missing token failure, stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "staging preflight failed") {
		t.Fatalf("expected failure message, stderr=%q", stderr)
	}
}

func TestPreflightAllRequiredEnvPasses(t *testing.T) {
	code, stdout, stderr := runPreflight(t, nil, validEnv())
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if strings.Contains(stdout, "status=fail") {
		t.Fatalf("unexpected failure in stdout=%q", stdout)
	}
}

func TestPreflightRejectsGlobalScope(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_SCOPE"] = "global"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "command-sync-scope status=fail") {
		t.Fatalf("expected scope failure, stdout=%q", stdout)
	}
}

func TestPreflightRejectsDeleteBulkAndMessageContent(t *testing.T) {
	for key, want := range map[string]string{
		"MHCAT_COMMAND_SYNC_ALLOW_DELETE":         "allow-delete status=fail",
		"MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE": "allow-bulk-overwrite status=fail",
		"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT":    "message-content-intent status=fail",
	} {
		env := validEnv()
		env[key] = "true"
		code, stdout, _ := runPreflight(t, nil, env)
		if code == 0 {
			t.Fatalf("expected %s to fail", key)
		}
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in stdout=%q", want, stdout)
		}
	}
}

func TestPreflightAcceptsMessageContentOnlyForAnnouncementRelay(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED"] = "true"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GUILD_MESSAGES_INTENT"] = "true"
	env["MHCAT_DISCORD_MESSAGE_CONTENT_INTENT"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "message-content-intent status=pass") || !strings.Contains(stdout, "announcement-relay-runtime-readiness status=pass") {
		t.Fatalf("expected relay message content pass checks, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAnnouncementRelayWithoutRequiredGatewayFlags(t *testing.T) {
	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":         "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MESSAGES_INTENT":  "MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true",
		"MHCAT_DISCORD_MESSAGE_CONTENT_INTENT": "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true",
	} {
		env := validEnv()
		env["MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED"] = "true"
		env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
		env["MHCAT_DISCORD_GUILD_MESSAGES_INTENT"] = "true"
		env["MHCAT_DISCORD_MESSAGE_CONTENT_INTENT"] = "true"
		env[key] = "false"
		code, stdout, _ := runPreflight(t, nil, env)
		if code == 0 {
			t.Fatalf("expected %s to fail", key)
		}
		if !strings.Contains(stdout, "announcement-relay-runtime-readiness status=fail") || !strings.Contains(stdout, want) {
			t.Fatalf("expected relay readiness failure %q, stdout=%q", want, stdout)
		}
	}
}

func TestPreflightRejectsApplicationPinMismatch(t *testing.T) {
	env := validEnv()
	env["MHCAT_STAGING_ALLOWED_APPLICATION_ID"] = "different-app"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "application-id-pin status=fail") {
		t.Fatalf("expected application pin failure, stdout=%q", stdout)
	}
}

func TestPreflightRejectsTicketCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TICKETS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "ticket-runtime-pairing status=fail") {
		t.Fatalf("expected ticket pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsTicketCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TICKETS"] = "true"
	env["MHCAT_FEATURE_TICKETS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "ticket-command-sync status=pass") || !strings.Contains(stdout, "ticket-runtime-pairing status=pass") {
		t.Fatalf("expected ticket pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenTicketRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_TICKETS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "ticket-runtime-pairing status=warn") {
		t.Fatalf("expected ticket runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsPollCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_POLLS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "poll-runtime-pairing status=fail") {
		t.Fatalf("expected poll pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsPollCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_POLLS"] = "true"
	env["MHCAT_FEATURE_POLLS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "poll-command-sync status=pass") || !strings.Contains(stdout, "poll-runtime-pairing status=pass") {
		t.Fatalf("expected poll pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenPollRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_POLLS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "poll-runtime-pairing status=warn") {
		t.Fatalf("expected poll runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsEconomyQueryCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "economy-query-runtime-pairing status=fail") {
		t.Fatalf("expected economy query pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsEconomyQueryCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY"] = "true"
	env["MHCAT_FEATURE_ECONOMY_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-query-command-sync status=pass") || !strings.Contains(stdout, "economy-query-runtime-pairing status=pass") {
		t.Fatalf("expected economy query pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenEconomyQueryRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ECONOMY_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-query-runtime-pairing status=warn") {
		t.Fatalf("expected economy query runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsEconomySignInCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "economy-signin-runtime-pairing status=fail") {
		t.Fatalf("expected economy sign-in pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsEconomySignInCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN"] = "true"
	env["MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-signin-command-sync status=pass") || !strings.Contains(stdout, "economy-signin-runtime-pairing status=pass") {
		t.Fatalf("expected economy sign-in pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenEconomySignInRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-signin-runtime-pairing status=warn") {
		t.Fatalf("expected economy sign-in runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsEconomySettingsCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "economy-settings-runtime-pairing status=fail") {
		t.Fatalf("expected economy settings pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsEconomySettingsCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS"] = "true"
	env["MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-settings-command-sync status=pass") || !strings.Contains(stdout, "economy-settings-runtime-pairing status=pass") {
		t.Fatalf("expected economy settings pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenEconomySettingsRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "economy-settings-runtime-pairing status=warn") {
		t.Fatalf("expected economy settings runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWorkCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WORK"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "work-runtime-pairing status=fail") {
		t.Fatalf("expected work pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWorkCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WORK"] = "true"
	env["MHCAT_FEATURE_WORK_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "work-command-sync status=pass") || !strings.Contains(stdout, "work-runtime-pairing status=pass") {
		t.Fatalf("expected work pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWorkRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WORK_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "work-runtime-pairing status=warn") {
		t.Fatalf("expected work runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWarningsCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "warnings-runtime-pairing status=fail") {
		t.Fatalf("expected warnings pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWarningsCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS"] = "true"
	env["MHCAT_FEATURE_WARNINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warnings-command-sync status=pass") || !strings.Contains(stdout, "warnings-runtime-pairing status=pass") {
		t.Fatalf("expected warnings pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWarningsRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WARNINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warnings-runtime-pairing status=warn") {
		t.Fatalf("expected warnings runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWarningSettingsCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "warning-settings-runtime-pairing status=fail") {
		t.Fatalf("expected warning settings pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWarningSettingsCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS"] = "true"
	env["MHCAT_FEATURE_WARNING_SETTINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-settings-command-sync status=pass") || !strings.Contains(stdout, "warning-settings-runtime-pairing status=pass") {
		t.Fatalf("expected warning settings pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWarningSettingsRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WARNING_SETTINGS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-settings-runtime-pairing status=warn") {
		t.Fatalf("expected warning settings runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWarningRemovalCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "warning-removal-runtime-pairing status=fail") {
		t.Fatalf("expected warning removal pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWarningRemovalCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL"] = "true"
	env["MHCAT_FEATURE_WARNING_REMOVAL_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-removal-command-sync status=pass") || !strings.Contains(stdout, "warning-removal-runtime-pairing status=pass") {
		t.Fatalf("expected warning removal pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWarningRemovalRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WARNING_REMOVAL_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-removal-runtime-pairing status=warn") {
		t.Fatalf("expected warning removal runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWarningIssueCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "warning-issue-runtime-pairing status=fail") {
		t.Fatalf("expected warning issue pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWarningIssueCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE"] = "true"
	env["MHCAT_FEATURE_WARNING_ISSUE_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-issue-command-sync status=pass") || !strings.Contains(stdout, "warning-issue-runtime-pairing status=pass") {
		t.Fatalf("expected warning issue pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWarningIssueRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WARNING_ISSUE_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "warning-issue-runtime-pairing status=warn") {
		t.Fatalf("expected warning issue runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsTranslateCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "translate-runtime-pairing status=fail") {
		t.Fatalf("expected translate pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsTranslateCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE"] = "true"
	env["MHCAT_FEATURE_TRANSLATE_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "translate-command-sync status=pass") || !strings.Contains(stdout, "translate-runtime-pairing status=pass") {
		t.Fatalf("expected translate pass checks, stdout=%q", stdout)
	}
}

func TestPreflightRejectsBalanceQueryCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "balance-query-runtime-pairing status=fail") {
		t.Fatalf("expected balance query pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsBalanceQueryCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY"] = "true"
	env["MHCAT_FEATURE_BALANCE_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "balance-query-command-sync status=pass") || !strings.Contains(stdout, "balance-query-runtime-pairing status=pass") {
		t.Fatalf("expected balance query pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenBalanceQueryRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_BALANCE_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "balance-query-runtime-pairing status=warn") {
		t.Fatalf("expected balance query runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsRedeemCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_REDEEM"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "redeem-runtime-pairing status=fail") {
		t.Fatalf("expected redeem pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsRedeemCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_REDEEM"] = "true"
	env["MHCAT_FEATURE_REDEEM_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "redeem-command-sync status=pass") || !strings.Contains(stdout, "redeem-runtime-pairing status=pass") {
		t.Fatalf("expected redeem pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenRedeemRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_REDEEM_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "redeem-runtime-pairing status=warn") {
		t.Fatalf("expected redeem runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAutoChatConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "autochat-config-runtime-pairing status=fail") {
		t.Fatalf("expected autochat config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAutoChatConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG"] = "true"
	env["MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "autochat-config-command-sync status=pass") || !strings.Contains(stdout, "autochat-config-runtime-pairing status=pass") {
		t.Fatalf("expected autochat config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAutoChatConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "autochat-config-runtime-pairing status=warn") {
		t.Fatalf("expected autochat config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAutoNotificationConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "auto-notification-config-runtime-pairing status=fail") {
		t.Fatalf("expected auto-notification config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAutoNotificationConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG"] = "true"
	env["MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "auto-notification-config-command-sync status=pass") || !strings.Contains(stdout, "auto-notification-config-runtime-pairing status=pass") {
		t.Fatalf("expected auto-notification config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAutoNotificationConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "auto-notification-config-runtime-pairing status=warn") {
		t.Fatalf("expected auto-notification config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAntiScamConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "anti-scam-config-runtime-pairing status=fail") {
		t.Fatalf("expected anti-scam config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAntiScamConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG"] = "true"
	env["MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "anti-scam-config-command-sync status=pass") || !strings.Contains(stdout, "anti-scam-config-runtime-pairing status=pass") {
		t.Fatalf("expected anti-scam config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAntiScamConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "anti-scam-config-runtime-pairing status=warn") {
		t.Fatalf("expected anti-scam config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAntiScamReportCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "anti-scam-report-runtime-pairing status=fail") {
		t.Fatalf("expected anti-scam report pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAntiScamReportCommandSyncWithRuntimeFlagAndWebhook(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT"] = "true"
	env["MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED"] = "true"
	env["MHCAT_REPORT_WEBHOOK_URL"] = "https://example.test/webhook"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "anti-scam-report-command-sync status=pass") || !strings.Contains(stdout, "anti-scam-report-runtime-pairing status=pass") {
		t.Fatalf("expected anti-scam report pass checks, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAntiScamReportRuntimeWithoutWebhook(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "anti-scam-report-runtime-pairing status=fail") || !strings.Contains(stdout, "MHCAT_REPORT_WEBHOOK_URL or REPORT_WEBHOOK") {
		t.Fatalf("expected anti-scam report webhook failure, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAntiScamReportRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED"] = "true"
	env["REPORT_WEBHOOK"] = "https://example.test/webhook"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "anti-scam-report-runtime-pairing status=warn") {
		t.Fatalf("expected anti-scam report runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsLoggingConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "logging-config-runtime-pairing status=fail") {
		t.Fatalf("expected logging config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsLoggingConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG"] = "true"
	env["MHCAT_FEATURE_LOGGING_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "logging-config-command-sync status=pass") || !strings.Contains(stdout, "logging-config-runtime-pairing status=pass") {
		t.Fatalf("expected logging config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightRejectsGachaPrizeListCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "gacha-prize-list-runtime-pairing status=fail") {
		t.Fatalf("expected gacha prize-list pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsGachaPrizeListCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST"] = "true"
	env["MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "gacha-prize-list-command-sync status=pass") || !strings.Contains(stdout, "gacha-prize-list-runtime-pairing status=pass") {
		t.Fatalf("expected gacha prize-list pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenGachaPrizeListRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "gacha-prize-list-runtime-pairing status=warn") {
		t.Fatalf("expected gacha prize-list runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsLotteryDisabledCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "lottery-disabled-runtime-pairing status=fail") {
		t.Fatalf("expected lottery disabled command pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsLotteryDisabledCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND"] = "true"
	env["MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "lottery-disabled-command-sync status=pass") || !strings.Contains(stdout, "lottery-disabled-runtime-pairing status=pass") {
		t.Fatalf("expected lottery disabled command pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenLotteryDisabledRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "lottery-disabled-runtime-pairing status=warn") {
		t.Fatalf("expected lottery disabled runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsStatsQueryCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "stats-query-runtime-pairing status=fail") {
		t.Fatalf("expected stats query pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsStatsQueryCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY"] = "true"
	env["MHCAT_FEATURE_STATS_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "stats-query-command-sync status=pass") || !strings.Contains(stdout, "stats-query-runtime-pairing status=pass") {
		t.Fatalf("expected stats query pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenStatsQueryRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_STATS_QUERY_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "stats-query-runtime-pairing status=warn") {
		t.Fatalf("expected stats query runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsBirthdayConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "birthday-config-runtime-pairing status=fail") {
		t.Fatalf("expected birthday config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsBirthdayConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG"] = "true"
	env["MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "birthday-config-command-sync status=pass") || !strings.Contains(stdout, "birthday-config-runtime-pairing status=pass") {
		t.Fatalf("expected birthday config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenBirthdayConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "birthday-config-runtime-pairing status=warn") {
		t.Fatalf("expected birthday config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAnnouncementConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "announcement-config-runtime-pairing status=fail") {
		t.Fatalf("expected announcement config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAnnouncementConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG"] = "true"
	env["MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "announcement-config-command-sync status=pass") || !strings.Contains(stdout, "announcement-config-runtime-pairing status=pass") {
		t.Fatalf("expected announcement config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAnnouncementConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "announcement-config-runtime-pairing status=warn") {
		t.Fatalf("expected announcement config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAnnouncementSendCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "announcement-send-runtime-pairing status=fail") {
		t.Fatalf("expected announcement send pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAnnouncementSendCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND"] = "true"
	env["MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "announcement-send-command-sync status=pass") || !strings.Contains(stdout, "announcement-send-runtime-pairing status=pass") {
		t.Fatalf("expected announcement send pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenAnnouncementSendRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "announcement-send-runtime-pairing status=warn") {
		t.Fatalf("expected announcement send runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsTextXPConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "text-xp-config-runtime-pairing status=fail") {
		t.Fatalf("expected text XP config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsTextXPConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG"] = "true"
	env["MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "text-xp-config-command-sync status=pass") || !strings.Contains(stdout, "text-xp-config-runtime-pairing status=pass") {
		t.Fatalf("expected text XP config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenTextXPConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "text-xp-config-runtime-pairing status=warn") {
		t.Fatalf("expected text XP config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsVoiceXPConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "voice-xp-config-runtime-pairing status=fail") {
		t.Fatalf("expected voice XP config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsVoiceXPConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG"] = "true"
	env["MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "voice-xp-config-command-sync status=pass") || !strings.Contains(stdout, "voice-xp-config-runtime-pairing status=pass") {
		t.Fatalf("expected voice XP config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenVoiceXPConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "voice-xp-config-runtime-pairing status=warn") {
		t.Fatalf("expected voice XP config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsXPProfileDisabledCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "xp-profile-disabled-runtime-pairing status=fail") {
		t.Fatalf("expected XP profile disabled pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsXPProfileDisabledCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS"] = "true"
	env["MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "xp-profile-disabled-command-sync status=pass") || !strings.Contains(stdout, "xp-profile-disabled-runtime-pairing status=pass") {
		t.Fatalf("expected XP profile disabled pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenXPProfileDisabledRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "xp-profile-disabled-runtime-pairing status=warn") {
		t.Fatalf("expected XP profile disabled runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsVoiceRoomConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "voice-room-config-runtime-pairing status=fail") {
		t.Fatalf("expected voice-room config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsVoiceRoomConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG"] = "true"
	env["MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "voice-room-config-command-sync status=pass") || !strings.Contains(stdout, "voice-room-config-runtime-pairing status=pass") {
		t.Fatalf("expected voice-room config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenVoiceRoomConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "voice-room-config-runtime-pairing status=warn") {
		t.Fatalf("expected voice-room config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsJoinRoleConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "join-role-config-runtime-pairing status=fail") {
		t.Fatalf("expected join-role config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsJoinRoleConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG"] = "true"
	env["MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "join-role-config-command-sync status=pass") || !strings.Contains(stdout, "join-role-config-runtime-pairing status=pass") {
		t.Fatalf("expected join-role config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenJoinRoleConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "join-role-config-runtime-pairing status=warn") {
		t.Fatalf("expected join-role config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsJoinRoleAssignmentWithGatewayAndGuildMembers(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED"] = "true"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "join-role-assignment-runtime-readiness status=pass") {
		t.Fatalf("expected assignment readiness pass, stdout=%q", stdout)
	}
}

func TestPreflightRejectsJoinRoleAssignmentWithoutGatewayOrGuildMembers(t *testing.T) {
	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":       "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT": "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true",
	} {
		env := validEnv()
		env["MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED"] = "true"
		env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
		env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
		env[key] = "false"
		code, stdout, _ := runPreflight(t, nil, env)
		if code == 0 {
			t.Fatalf("expected %s to fail", key)
		}
		if !strings.Contains(stdout, "join-role-assignment-runtime-readiness status=fail") || !strings.Contains(stdout, want) {
			t.Fatalf("expected assignment failure %q, stdout=%q", want, stdout)
		}
	}
}

func TestPreflightAcceptsLeaveMessageDeliveryWithGatewayAndGuildMembers(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED"] = "true"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "leave-message-delivery-runtime-readiness status=pass") {
		t.Fatalf("expected leave delivery readiness pass, stdout=%q", stdout)
	}
}

func TestPreflightRejectsLeaveMessageDeliveryWithoutGatewayOrGuildMembers(t *testing.T) {
	for key, want := range map[string]string{
		"MHCAT_DISCORD_ENABLE_GATEWAY":       "MHCAT_DISCORD_ENABLE_GATEWAY=true",
		"MHCAT_DISCORD_GUILD_MEMBERS_INTENT": "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true",
	} {
		env := validEnv()
		env["MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED"] = "true"
		env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
		env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
		env[key] = "false"
		code, stdout, _ := runPreflight(t, nil, env)
		if code == 0 {
			t.Fatalf("expected %s to fail", key)
		}
		if !strings.Contains(stdout, "leave-message-delivery-runtime-readiness status=fail") || !strings.Contains(stdout, want) {
			t.Fatalf("expected leave delivery failure %q, stdout=%q", want, stdout)
		}
	}
}

func TestPreflightRejectsWelcomeMessageConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "welcome-message-config-runtime-pairing status=fail") {
		t.Fatalf("expected welcome-message config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWelcomeMessageConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG"] = "true"
	env["MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "welcome-message-config-command-sync status=pass") || !strings.Contains(stdout, "welcome-message-config-runtime-pairing status=pass") {
		t.Fatalf("expected welcome-message config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenWelcomeMessageConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "welcome-message-config-runtime-pairing status=warn") {
		t.Fatalf("expected welcome-message config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsVerificationConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "verification-config-runtime-pairing status=fail") {
		t.Fatalf("expected verification config pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsVerificationConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG"] = "true"
	env["MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "verification-config-command-sync status=pass") || !strings.Contains(stdout, "verification-config-runtime-pairing status=pass") {
		t.Fatalf("expected verification config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenVerificationConfigRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "verification-config-runtime-pairing status=warn") {
		t.Fatalf("expected verification config runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsVerificationFlowCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "verification-flow-runtime-pairing status=fail") {
		t.Fatalf("expected verification flow pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsVerificationFlowCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW"] = "true"
	env["MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "verification-flow-command-sync status=pass") || !strings.Contains(stdout, "verification-flow-runtime-pairing status=pass") {
		t.Fatalf("expected verification flow pass checks, stdout=%q", stdout)
	}
}

func TestPreflightWarnsWhenVerificationFlowRuntimeEnabledWithoutCommandSync(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected warning-only exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "verification-flow-runtime-pairing status=warn") {
		t.Fatalf("expected verification flow runtime warning, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAccountAgeConfigCommandSyncWithoutRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "account-age-config-runtime-pairing status=fail") {
		t.Fatalf("expected account-age pairing failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAccountAgeConfigCommandSyncWithRuntimeFlag(t *testing.T) {
	env := validEnv()
	env["MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG"] = "true"
	env["MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "account-age-config-command-sync status=pass") || !strings.Contains(stdout, "account-age-config-runtime-pairing status=pass") {
		t.Fatalf("expected account-age config pass checks, stdout=%q", stdout)
	}
}

func TestPreflightRejectsAccountAgePolicyWithoutGatewayIntent(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "account-age-policy-runtime-readiness status=fail") {
		t.Fatalf("expected account-age policy readiness failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsAccountAgePolicyWithGatewayIntent(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED"] = "true"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "account-age-policy-runtime-readiness status=pass") {
		t.Fatalf("expected account-age policy readiness pass, stdout=%q", stdout)
	}
}

func TestPreflightRejectsWelcomeMessageDeliveryWithoutGatewayIntent(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED"] = "true"
	code, stdout, _ := runPreflight(t, nil, env)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stdout, "welcome-message-delivery-runtime-readiness status=fail") {
		t.Fatalf("expected welcome-message delivery readiness failure, stdout=%q", stdout)
	}
}

func TestPreflightAcceptsWelcomeMessageDeliveryWithGatewayIntent(t *testing.T) {
	env := validEnv()
	env["MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED"] = "true"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GUILD_MEMBERS_INTENT"] = "true"
	code, stdout, stderr := runPreflight(t, nil, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stderr=%q stdout=%q", stderr, stdout)
	}
	if !strings.Contains(stdout, "welcome-message-delivery-runtime-readiness status=pass") {
		t.Fatalf("expected welcome-message delivery readiness pass, stdout=%q", stdout)
	}
}

func TestPreflightJSONDeterministicAndRedacted(t *testing.T) {
	env := validEnv()
	token := strings.Repeat("a", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 32)
	mongo := "mongodb" + "+srv://" + "user:" + "redaction-test-password" + "@cluster.example/mhcat"
	env["MHCAT_DISCORD_TOKEN"] = token
	env["MHCAT_MONGODB_URI"] = mongo
	code, first, _ := runPreflight(t, []string{"--format", "json"}, env)
	if code != 0 {
		t.Fatalf("expected exit 0, stdout=%q", first)
	}
	_, second, _ := runPreflight(t, []string{"--format", "json"}, env)
	if first != second {
		t.Fatalf("json output is not deterministic:\n%s\n---\n%s", first, second)
	}
	if strings.Contains(first, token) || strings.Contains(first, "super-secret-password") {
		t.Fatalf("secret leaked in json output: %s", first)
	}
	var parsed report
	if err := json.Unmarshal([]byte(first), &parsed); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if len(parsed.Checks) == 0 {
		t.Fatal("expected checks")
	}
}

func TestPreflightInvalidFormatFails(t *testing.T) {
	code, _, stderr := runPreflight(t, []string{"--format", "xml"}, validEnv())
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "format must be text or json") {
		t.Fatalf("stderr=%q", stderr)
	}
}

func runPreflight(t *testing.T, args []string, env map[string]string) (int, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(args, func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func validEnv() map[string]string {
	return map[string]string{
		"MHCAT_DISCORD_TOKEN":          "test-token",
		"MHCAT_DISCORD_APPLICATION_ID": "app-1",
		"MHCAT_STAGING_GUILD_ID":       "guild-1",
		"MHCAT_MONGODB_URI":            "mongodb://localhost:27017/mhcat-staging",
		"MHCAT_MONGODB_DATABASE":       "mhcat_staging",
	}
}
