package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

type lookupFunc func(string) (string, bool)

type status string

const (
	statusPass    status = "pass"
	statusWarn    status = "warn"
	statusFail    status = "fail"
	statusSkipped status = "skipped"
)

type checkResult struct {
	Name    string `json:"name"`
	Status  status `json:"status"`
	Message string `json:"message"`
}

type report struct {
	Checks []checkResult `json:"checks"`
}

func main() {
	os.Exit(run(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func run(args []string, lookup lookupFunc, stdout io.Writer, stderr io.Writer) int {
	format, err := parseFlags(args, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "staging preflight flag error: %v\n", err)
		return 1
	}
	results := buildReport(lookup)
	out := report{Checks: results}
	if err := formatReport(stdout, out, format); err != nil {
		fmt.Fprintf(stderr, "staging preflight output error: %v\n", err)
		return 1
	}
	if hasFailures(results) {
		fmt.Fprintln(stderr, "staging preflight failed: fix failed checks before running staging sync or gateway smoke")
		return 1
	}
	return 0
}

func parseFlags(args []string, stderr io.Writer) (string, error) {
	flags := flag.NewFlagSet("mhcat-staging-preflight", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "text", "output format: text or json")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	switch *format {
	case "text", "json":
		return *format, nil
	default:
		return "", fmt.Errorf("format must be text or json")
	}
}

func buildReport(lookup lookupFunc) []checkResult {
	checks := []checkResult{
		required(lookup, "discord-token", "MHCAT_DISCORD_TOKEN", true),
		required(lookup, "application-id", "MHCAT_DISCORD_APPLICATION_ID", false),
		required(lookup, "staging-guild-id", "MHCAT_STAGING_GUILD_ID", false),
		required(lookup, "mongo-uri", "MHCAT_MONGODB_URI", true),
		required(lookup, "mongo-database", "MHCAT_MONGODB_DATABASE", false),
		commandScope(lookup),
		boolMustNotBeTrue(lookup, "allow-delete", "MHCAT_COMMAND_SYNC_ALLOW_DELETE"),
		boolMustNotBeTrue(lookup, "allow-bulk-overwrite", "MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE"),
		messageContentIntent(lookup),
		applicationPin(lookup),
		ticketCommandSync(lookup),
		ticketRuntimePairing(lookup),
		pollCommandSync(lookup),
		pollRuntimePairing(lookup),
		economyQueryCommandSync(lookup),
		economyQueryRuntimePairing(lookup),
		economySignInCommandSync(lookup),
		economySignInRuntimePairing(lookup),
		economySettingsCommandSync(lookup),
		economySettingsRuntimePairing(lookup),
		workCommandSync(lookup),
		workRuntimePairing(lookup),
		warningsCommandSync(lookup),
		warningsRuntimePairing(lookup),
		translateCommandSync(lookup),
		translateRuntimePairing(lookup),
		balanceQueryCommandSync(lookup),
		balanceQueryRuntimePairing(lookup),
		autoChatConfigCommandSync(lookup),
		autoChatConfigRuntimePairing(lookup),
		autoNotificationConfigCommandSync(lookup),
		autoNotificationConfigRuntimePairing(lookup),
		antiScamConfigCommandSync(lookup),
		antiScamConfigRuntimePairing(lookup),
		antiScamReportCommandSync(lookup),
		antiScamReportRuntimePairing(lookup),
		loggingConfigCommandSync(lookup),
		loggingConfigRuntimePairing(lookup),
		gachaPrizeListCommandSync(lookup),
		gachaPrizeListRuntimePairing(lookup),
		lotteryDisabledCommandSync(lookup),
		lotteryDisabledRuntimePairing(lookup),
		statsQueryCommandSync(lookup),
		statsQueryRuntimePairing(lookup),
		birthdayConfigCommandSync(lookup),
		birthdayConfigRuntimePairing(lookup),
		announcementConfigCommandSync(lookup),
		announcementConfigRuntimePairing(lookup),
		announcementSendCommandSync(lookup),
		announcementSendRuntimePairing(lookup),
		announcementRelayRuntimeReadiness(lookup),
		textXPConfigCommandSync(lookup),
		textXPConfigRuntimePairing(lookup),
		voiceXPConfigCommandSync(lookup),
		voiceXPConfigRuntimePairing(lookup),
		joinRoleConfigCommandSync(lookup),
		joinRoleConfigRuntimePairing(lookup),
		joinRoleAssignmentRuntimeReadiness(lookup),
		leaveMessageDeliveryRuntimeReadiness(lookup),
		welcomeMessageDeliveryRuntimeReadiness(lookup),
		welcomeMessageConfigCommandSync(lookup),
		welcomeMessageConfigRuntimePairing(lookup),
		verificationConfigCommandSync(lookup),
		verificationConfigRuntimePairing(lookup),
		verificationFlowCommandSync(lookup),
		verificationFlowRuntimePairing(lookup),
		accountAgeConfigCommandSync(lookup),
		accountAgeConfigRuntimePairing(lookup),
		accountAgePolicyRuntimeReadiness(lookup),
	}
	sort.SliceStable(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})
	return checks
}

func required(lookup lookupFunc, name string, key string, secret bool) checkResult {
	value, ok := lookup(key)
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return checkResult{Name: name, Status: statusFail, Message: key + " is required"}
	}
	if secret {
		return checkResult{Name: name, Status: statusPass, Message: key + " is present as " + config.RedactValue(key, value)}
	}
	return checkResult{Name: name, Status: statusPass, Message: key + " is present"}
}

func commandScope(lookup lookupFunc) checkResult {
	value, ok := lookup("MHCAT_COMMAND_SYNC_SCOPE")
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return checkResult{Name: "command-sync-scope", Status: statusPass, Message: "MHCAT_COMMAND_SYNC_SCOPE is unset and defaults to guild"}
	}
	if value != "guild" {
		return checkResult{Name: "command-sync-scope", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_SCOPE must be guild for staging"}
	}
	return checkResult{Name: "command-sync-scope", Status: statusPass, Message: "MHCAT_COMMAND_SYNC_SCOPE is guild"}
}

func boolMustNotBeTrue(lookup lookupFunc, name string, key string) checkResult {
	value, ok := lookup(key)
	value = strings.TrimSpace(strings.ToLower(value))
	if !ok || value == "" || value == "false" || value == "0" {
		return checkResult{Name: name, Status: statusPass, Message: key + " is disabled"}
	}
	if value == "true" || value == "1" {
		return checkResult{Name: name, Status: statusFail, Message: key + " must be false for staging preflight"}
	}
	return checkResult{Name: name, Status: statusFail, Message: key + " must be a boolean"}
}

func messageContentIntent(lookup lookupFunc) checkResult {
	messageContent, err := boolValue(lookup, "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT")
	if err != nil {
		return checkResult{Name: "message-content-intent", Status: statusFail, Message: err.Error()}
	}
	relayEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED")
	if err != nil {
		return checkResult{Name: "message-content-intent", Status: statusFail, Message: err.Error()}
	}
	if !messageContent {
		return checkResult{Name: "message-content-intent", Status: statusPass, Message: "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT is disabled"}
	}
	if relayEnabled {
		return checkResult{Name: "message-content-intent", Status: statusPass, Message: "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT is enabled for announcement relay only"}
	}
	return checkResult{Name: "message-content-intent", Status: statusFail, Message: "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT must be false unless MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true"}
}

func ticketCommandSync(lookup lookupFunc) checkResult {
	includeTickets, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TICKETS")
	if err != nil {
		return checkResult{Name: "ticket-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeTickets {
		return checkResult{Name: "ticket-command-sync", Status: statusPass, Message: "ticket command sync include is enabled for staging review"}
	}
	return checkResult{Name: "ticket-command-sync", Status: statusSkipped, Message: "ticket command sync include is disabled"}
}

func ticketRuntimePairing(lookup lookupFunc) checkResult {
	includeTickets, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TICKETS")
	if err != nil {
		return checkResult{Name: "ticket-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	ticketsEnabled, err := boolValue(lookup, "MHCAT_FEATURE_TICKETS_ENABLED")
	if err != nil {
		return checkResult{Name: "ticket-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeTickets && !ticketsEnabled {
		return checkResult{Name: "ticket-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true requires MHCAT_FEATURE_TICKETS_ENABLED=true in the staging runtime"}
	}
	if includeTickets && ticketsEnabled {
		return checkResult{Name: "ticket-runtime-pairing", Status: statusPass, Message: "ticket command sync and runtime feature flag are paired"}
	}
	if ticketsEnabled {
		return checkResult{Name: "ticket-runtime-pairing", Status: statusWarn, Message: "ticket runtime is enabled but ticket command sync include is disabled"}
	}
	return checkResult{Name: "ticket-runtime-pairing", Status: statusSkipped, Message: "ticket runtime and command sync include are disabled"}
}

func pollCommandSync(lookup lookupFunc) checkResult {
	includePolls, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_POLLS")
	if err != nil {
		return checkResult{Name: "poll-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includePolls {
		return checkResult{Name: "poll-command-sync", Status: statusPass, Message: "poll command sync include is enabled for staging review"}
	}
	return checkResult{Name: "poll-command-sync", Status: statusSkipped, Message: "poll command sync include is disabled"}
}

func pollRuntimePairing(lookup lookupFunc) checkResult {
	includePolls, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_POLLS")
	if err != nil {
		return checkResult{Name: "poll-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	pollsEnabled, err := boolValue(lookup, "MHCAT_FEATURE_POLLS_ENABLED")
	if err != nil {
		return checkResult{Name: "poll-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includePolls && !pollsEnabled {
		return checkResult{Name: "poll-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true requires MHCAT_FEATURE_POLLS_ENABLED=true in the staging runtime"}
	}
	if includePolls && pollsEnabled {
		return checkResult{Name: "poll-runtime-pairing", Status: statusPass, Message: "poll command sync and runtime feature flag are paired"}
	}
	if pollsEnabled {
		return checkResult{Name: "poll-runtime-pairing", Status: statusWarn, Message: "poll runtime is enabled but poll command sync include is disabled"}
	}
	return checkResult{Name: "poll-runtime-pairing", Status: statusSkipped, Message: "poll runtime and command sync include are disabled"}
}

func economyQueryCommandSync(lookup lookupFunc) checkResult {
	includeEconomyQuery, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY")
	if err != nil {
		return checkResult{Name: "economy-query-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeEconomyQuery {
		return checkResult{Name: "economy-query-command-sync", Status: statusPass, Message: "economy query command sync include is enabled for staging review"}
	}
	return checkResult{Name: "economy-query-command-sync", Status: statusSkipped, Message: "economy query command sync include is disabled"}
}

func economyQueryRuntimePairing(lookup lookupFunc) checkResult {
	includeEconomyQuery, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY")
	if err != nil {
		return checkResult{Name: "economy-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	economyQueryEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ECONOMY_QUERY_ENABLED")
	if err != nil {
		return checkResult{Name: "economy-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeEconomyQuery && !economyQueryEnabled {
		return checkResult{Name: "economy-query-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true requires MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true in the staging runtime"}
	}
	if includeEconomyQuery && economyQueryEnabled {
		return checkResult{Name: "economy-query-runtime-pairing", Status: statusPass, Message: "economy query command sync and runtime feature flag are paired"}
	}
	if economyQueryEnabled {
		return checkResult{Name: "economy-query-runtime-pairing", Status: statusWarn, Message: "economy query runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "economy-query-runtime-pairing", Status: statusSkipped, Message: "economy query runtime and command sync include are disabled"}
}

func economySignInCommandSync(lookup lookupFunc) checkResult {
	includeEconomySignIn, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN")
	if err != nil {
		return checkResult{Name: "economy-signin-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeEconomySignIn {
		return checkResult{Name: "economy-signin-command-sync", Status: statusPass, Message: "economy sign-in commands sync include is enabled for staging review"}
	}
	return checkResult{Name: "economy-signin-command-sync", Status: statusSkipped, Message: "economy sign-in commands sync include is disabled"}
}

func economySignInRuntimePairing(lookup lookupFunc) checkResult {
	includeEconomySignIn, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN")
	if err != nil {
		return checkResult{Name: "economy-signin-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	economySignInEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED")
	if err != nil {
		return checkResult{Name: "economy-signin-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeEconomySignIn && !economySignInEnabled {
		return checkResult{Name: "economy-signin-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true requires MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true in the staging runtime"}
	}
	if includeEconomySignIn && economySignInEnabled {
		return checkResult{Name: "economy-signin-runtime-pairing", Status: statusPass, Message: "economy sign-in commands sync and runtime feature flag are paired"}
	}
	if economySignInEnabled {
		return checkResult{Name: "economy-signin-runtime-pairing", Status: statusWarn, Message: "economy sign-in runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "economy-signin-runtime-pairing", Status: statusSkipped, Message: "economy sign-in runtime and command sync include are disabled"}
}

func economySettingsCommandSync(lookup lookupFunc) checkResult {
	includeEconomySettings, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS")
	if err != nil {
		return checkResult{Name: "economy-settings-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeEconomySettings {
		return checkResult{Name: "economy-settings-command-sync", Status: statusPass, Message: "economy settings command sync include is enabled for staging review"}
	}
	return checkResult{Name: "economy-settings-command-sync", Status: statusSkipped, Message: "economy settings command sync include is disabled"}
}

func economySettingsRuntimePairing(lookup lookupFunc) checkResult {
	includeEconomySettings, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS")
	if err != nil {
		return checkResult{Name: "economy-settings-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	economySettingsEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED")
	if err != nil {
		return checkResult{Name: "economy-settings-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeEconomySettings && !economySettingsEnabled {
		return checkResult{Name: "economy-settings-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true requires MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true in the staging runtime"}
	}
	if includeEconomySettings && economySettingsEnabled {
		return checkResult{Name: "economy-settings-runtime-pairing", Status: statusPass, Message: "economy settings command sync and runtime feature flag are paired"}
	}
	if economySettingsEnabled {
		return checkResult{Name: "economy-settings-runtime-pairing", Status: statusWarn, Message: "economy settings runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "economy-settings-runtime-pairing", Status: statusSkipped, Message: "economy settings runtime and command sync include are disabled"}
}

func workCommandSync(lookup lookupFunc) checkResult {
	includeWork, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WORK")
	if err != nil {
		return checkResult{Name: "work-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeWork {
		return checkResult{Name: "work-command-sync", Status: statusPass, Message: "work command sync include is enabled for staging review"}
	}
	return checkResult{Name: "work-command-sync", Status: statusSkipped, Message: "work command sync include is disabled"}
}

func workRuntimePairing(lookup lookupFunc) checkResult {
	includeWork, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WORK")
	if err != nil {
		return checkResult{Name: "work-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	workEnabled, err := boolValue(lookup, "MHCAT_FEATURE_WORK_ENABLED")
	if err != nil {
		return checkResult{Name: "work-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeWork && !workEnabled {
		return checkResult{Name: "work-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_WORK=true requires MHCAT_FEATURE_WORK_ENABLED=true in the staging runtime"}
	}
	if includeWork && workEnabled {
		return checkResult{Name: "work-runtime-pairing", Status: statusPass, Message: "work command sync and runtime feature flag are paired"}
	}
	if workEnabled {
		return checkResult{Name: "work-runtime-pairing", Status: statusWarn, Message: "work runtime is enabled but work command sync include is disabled"}
	}
	return checkResult{Name: "work-runtime-pairing", Status: statusSkipped, Message: "work runtime and command sync include are disabled"}
}

func warningsCommandSync(lookup lookupFunc) checkResult {
	includeWarnings, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS")
	if err != nil {
		return checkResult{Name: "warnings-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeWarnings {
		return checkResult{Name: "warnings-command-sync", Status: statusPass, Message: "warnings command sync include is enabled for staging review"}
	}
	return checkResult{Name: "warnings-command-sync", Status: statusSkipped, Message: "warnings command sync include is disabled"}
}

func warningsRuntimePairing(lookup lookupFunc) checkResult {
	includeWarnings, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS")
	if err != nil {
		return checkResult{Name: "warnings-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	warningsEnabled, err := boolValue(lookup, "MHCAT_FEATURE_WARNINGS_ENABLED")
	if err != nil {
		return checkResult{Name: "warnings-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeWarnings && !warningsEnabled {
		return checkResult{Name: "warnings-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true requires MHCAT_FEATURE_WARNINGS_ENABLED=true in the staging runtime"}
	}
	if includeWarnings && warningsEnabled {
		return checkResult{Name: "warnings-runtime-pairing", Status: statusPass, Message: "warnings command sync and runtime feature flag are paired"}
	}
	if warningsEnabled {
		return checkResult{Name: "warnings-runtime-pairing", Status: statusWarn, Message: "warnings runtime is enabled but warnings command sync include is disabled"}
	}
	return checkResult{Name: "warnings-runtime-pairing", Status: statusSkipped, Message: "warnings runtime and command sync include are disabled"}
}

func translateCommandSync(lookup lookupFunc) checkResult {
	includeTranslate, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE")
	if err != nil {
		return checkResult{Name: "translate-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeTranslate {
		return checkResult{Name: "translate-command-sync", Status: statusPass, Message: "translate command sync include is enabled for staging review"}
	}
	return checkResult{Name: "translate-command-sync", Status: statusSkipped, Message: "translate command sync include is disabled"}
}

func translateRuntimePairing(lookup lookupFunc) checkResult {
	includeTranslate, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE")
	if err != nil {
		return checkResult{Name: "translate-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	translateEnabled, err := boolValue(lookup, "MHCAT_FEATURE_TRANSLATE_ENABLED")
	if err != nil {
		return checkResult{Name: "translate-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeTranslate && !translateEnabled {
		return checkResult{Name: "translate-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true requires MHCAT_FEATURE_TRANSLATE_ENABLED=true in the staging runtime"}
	}
	if includeTranslate && translateEnabled {
		return checkResult{Name: "translate-runtime-pairing", Status: statusPass, Message: "translate command sync and runtime feature flag are paired"}
	}
	if translateEnabled {
		return checkResult{Name: "translate-runtime-pairing", Status: statusWarn, Message: "translate runtime is enabled but translate command sync include is disabled"}
	}
	return checkResult{Name: "translate-runtime-pairing", Status: statusSkipped, Message: "translate runtime and command sync include are disabled"}
}

func balanceQueryCommandSync(lookup lookupFunc) checkResult {
	includeBalance, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY")
	if err != nil {
		return checkResult{Name: "balance-query-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeBalance {
		return checkResult{Name: "balance-query-command-sync", Status: statusPass, Message: "balance query command sync include is enabled for staging review"}
	}
	return checkResult{Name: "balance-query-command-sync", Status: statusSkipped, Message: "balance query command sync include is disabled"}
}

func balanceQueryRuntimePairing(lookup lookupFunc) checkResult {
	includeBalance, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY")
	if err != nil {
		return checkResult{Name: "balance-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	balanceEnabled, err := boolValue(lookup, "MHCAT_FEATURE_BALANCE_QUERY_ENABLED")
	if err != nil {
		return checkResult{Name: "balance-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeBalance && !balanceEnabled {
		return checkResult{Name: "balance-query-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true requires MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true in the staging runtime"}
	}
	if includeBalance && balanceEnabled {
		return checkResult{Name: "balance-query-runtime-pairing", Status: statusPass, Message: "balance query command sync and runtime feature flag are paired"}
	}
	if balanceEnabled {
		return checkResult{Name: "balance-query-runtime-pairing", Status: statusWarn, Message: "balance query runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "balance-query-runtime-pairing", Status: statusSkipped, Message: "balance query runtime and command sync include are disabled"}
}

func autoChatConfigCommandSync(lookup lookupFunc) checkResult {
	includeAutoChat, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG")
	if err != nil {
		return checkResult{Name: "autochat-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAutoChat {
		return checkResult{Name: "autochat-config-command-sync", Status: statusPass, Message: "autochat config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "autochat-config-command-sync", Status: statusSkipped, Message: "autochat config command sync include is disabled"}
}

func autoChatConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeAutoChat, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG")
	if err != nil {
		return checkResult{Name: "autochat-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	autoChatEnabled, err := boolValue(lookup, "MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "autochat-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAutoChat && !autoChatEnabled {
		return checkResult{Name: "autochat-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true requires MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeAutoChat && autoChatEnabled {
		return checkResult{Name: "autochat-config-runtime-pairing", Status: statusPass, Message: "autochat config command sync and runtime feature flag are paired"}
	}
	if autoChatEnabled {
		return checkResult{Name: "autochat-config-runtime-pairing", Status: statusWarn, Message: "autochat config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "autochat-config-runtime-pairing", Status: statusSkipped, Message: "autochat config runtime and command sync include are disabled"}
}

func autoNotificationConfigCommandSync(lookup lookupFunc) checkResult {
	includeAutoNotification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG")
	if err != nil {
		return checkResult{Name: "auto-notification-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAutoNotification {
		return checkResult{Name: "auto-notification-config-command-sync", Status: statusPass, Message: "auto-notification config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "auto-notification-config-command-sync", Status: statusSkipped, Message: "auto-notification config command sync include is disabled"}
}

func autoNotificationConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeAutoNotification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG")
	if err != nil {
		return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	autoNotificationEnabled, err := boolValue(lookup, "MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAutoNotification && !autoNotificationEnabled {
		return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true requires MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeAutoNotification && autoNotificationEnabled {
		return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusPass, Message: "auto-notification config command sync and runtime feature flag are paired"}
	}
	if autoNotificationEnabled {
		return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusWarn, Message: "auto-notification config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "auto-notification-config-runtime-pairing", Status: statusSkipped, Message: "auto-notification config runtime and command sync include are disabled"}
}

func antiScamConfigCommandSync(lookup lookupFunc) checkResult {
	includeAntiScam, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG")
	if err != nil {
		return checkResult{Name: "anti-scam-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAntiScam {
		return checkResult{Name: "anti-scam-config-command-sync", Status: statusPass, Message: "anti-scam config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "anti-scam-config-command-sync", Status: statusSkipped, Message: "anti-scam config command sync include is disabled"}
}

func antiScamConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeAntiScam, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG")
	if err != nil {
		return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	antiScamEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAntiScam && !antiScamEnabled {
		return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true requires MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeAntiScam && antiScamEnabled {
		return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusPass, Message: "anti-scam config command sync and runtime feature flag are paired"}
	}
	if antiScamEnabled {
		return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusWarn, Message: "anti-scam config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "anti-scam-config-runtime-pairing", Status: statusSkipped, Message: "anti-scam config runtime and command sync include are disabled"}
}

func antiScamReportCommandSync(lookup lookupFunc) checkResult {
	includeAntiScamReport, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT")
	if err != nil {
		return checkResult{Name: "anti-scam-report-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAntiScamReport {
		return checkResult{Name: "anti-scam-report-command-sync", Status: statusPass, Message: "anti-scam report command sync include is enabled for staging review"}
	}
	return checkResult{Name: "anti-scam-report-command-sync", Status: statusSkipped, Message: "anti-scam report command sync include is disabled"}
}

func antiScamReportRuntimePairing(lookup lookupFunc) checkResult {
	includeAntiScamReport, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT")
	if err != nil {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	antiScamReportEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED")
	if err != nil {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	webhookURL, _ := lookup("MHCAT_REPORT_WEBHOOK_URL")
	legacyWebhookURL, _ := lookup("REPORT_WEBHOOK")
	hasWebhook := strings.TrimSpace(webhookURL) != "" || strings.TrimSpace(legacyWebhookURL) != ""
	if includeAntiScamReport && !antiScamReportEnabled {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT=true requires MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true in the staging runtime"}
	}
	if antiScamReportEnabled && !hasWebhook {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusFail, Message: "MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true requires MHCAT_REPORT_WEBHOOK_URL or REPORT_WEBHOOK"}
	}
	if includeAntiScamReport && antiScamReportEnabled {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusPass, Message: "anti-scam report command sync, runtime feature flag, and webhook URL are paired"}
	}
	if antiScamReportEnabled {
		return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusWarn, Message: "anti-scam report runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "anti-scam-report-runtime-pairing", Status: statusSkipped, Message: "anti-scam report runtime and command sync include are disabled"}
}

func loggingConfigCommandSync(lookup lookupFunc) checkResult {
	includeLogging, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG")
	if err != nil {
		return checkResult{Name: "logging-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeLogging {
		return checkResult{Name: "logging-config-command-sync", Status: statusPass, Message: "logging config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "logging-config-command-sync", Status: statusSkipped, Message: "logging config command sync include is disabled"}
}

func loggingConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeLogging, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG")
	if err != nil {
		return checkResult{Name: "logging-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	loggingEnabled, err := boolValue(lookup, "MHCAT_FEATURE_LOGGING_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "logging-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeLogging && !loggingEnabled {
		return checkResult{Name: "logging-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true requires MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeLogging && loggingEnabled {
		return checkResult{Name: "logging-config-runtime-pairing", Status: statusPass, Message: "logging config command sync and runtime feature flag are paired"}
	}
	if loggingEnabled {
		return checkResult{Name: "logging-config-runtime-pairing", Status: statusWarn, Message: "logging config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "logging-config-runtime-pairing", Status: statusSkipped, Message: "logging config runtime and command sync include are disabled"}
}

func gachaPrizeListCommandSync(lookup lookupFunc) checkResult {
	includeGacha, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST")
	if err != nil {
		return checkResult{Name: "gacha-prize-list-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeGacha {
		return checkResult{Name: "gacha-prize-list-command-sync", Status: statusPass, Message: "gacha prize-list command sync include is enabled for staging review"}
	}
	return checkResult{Name: "gacha-prize-list-command-sync", Status: statusSkipped, Message: "gacha prize-list command sync include is disabled"}
}

func gachaPrizeListRuntimePairing(lookup lookupFunc) checkResult {
	includeGacha, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST")
	if err != nil {
		return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	gachaEnabled, err := boolValue(lookup, "MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED")
	if err != nil {
		return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeGacha && !gachaEnabled {
		return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true requires MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true in the staging runtime"}
	}
	if includeGacha && gachaEnabled {
		return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusPass, Message: "gacha prize-list command sync and runtime feature flag are paired"}
	}
	if gachaEnabled {
		return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusWarn, Message: "gacha prize-list runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "gacha-prize-list-runtime-pairing", Status: statusSkipped, Message: "gacha prize-list runtime and command sync include are disabled"}
}

func lotteryDisabledCommandSync(lookup lookupFunc) checkResult {
	includeLottery, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND")
	if err != nil {
		return checkResult{Name: "lottery-disabled-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeLottery {
		return checkResult{Name: "lottery-disabled-command-sync", Status: statusPass, Message: "lottery disabled command sync include is enabled for staging review"}
	}
	return checkResult{Name: "lottery-disabled-command-sync", Status: statusSkipped, Message: "lottery disabled command sync include is disabled"}
}

func lotteryDisabledRuntimePairing(lookup lookupFunc) checkResult {
	includeLottery, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND")
	if err != nil {
		return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	lotteryEnabled, err := boolValue(lookup, "MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED")
	if err != nil {
		return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeLottery && !lotteryEnabled {
		return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true requires MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true in the staging runtime"}
	}
	if includeLottery && lotteryEnabled {
		return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusPass, Message: "lottery disabled command sync and runtime feature flag are paired"}
	}
	if lotteryEnabled {
		return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusWarn, Message: "lottery disabled command runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "lottery-disabled-runtime-pairing", Status: statusSkipped, Message: "lottery disabled command runtime and command sync include are disabled"}
}

func statsQueryCommandSync(lookup lookupFunc) checkResult {
	includeStats, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY")
	if err != nil {
		return checkResult{Name: "stats-query-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeStats {
		return checkResult{Name: "stats-query-command-sync", Status: statusPass, Message: "stats query command sync include is enabled for staging review"}
	}
	return checkResult{Name: "stats-query-command-sync", Status: statusSkipped, Message: "stats query command sync include is disabled"}
}

func statsQueryRuntimePairing(lookup lookupFunc) checkResult {
	includeStats, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY")
	if err != nil {
		return checkResult{Name: "stats-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	statsEnabled, err := boolValue(lookup, "MHCAT_FEATURE_STATS_QUERY_ENABLED")
	if err != nil {
		return checkResult{Name: "stats-query-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeStats && !statsEnabled {
		return checkResult{Name: "stats-query-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true requires MHCAT_FEATURE_STATS_QUERY_ENABLED=true in the staging runtime"}
	}
	if includeStats && statsEnabled {
		return checkResult{Name: "stats-query-runtime-pairing", Status: statusPass, Message: "stats query command sync and runtime feature flag are paired"}
	}
	if statsEnabled {
		return checkResult{Name: "stats-query-runtime-pairing", Status: statusWarn, Message: "stats query runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "stats-query-runtime-pairing", Status: statusSkipped, Message: "stats query runtime and command sync include are disabled"}
}

func birthdayConfigCommandSync(lookup lookupFunc) checkResult {
	includeBirthday, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG")
	if err != nil {
		return checkResult{Name: "birthday-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeBirthday {
		return checkResult{Name: "birthday-config-command-sync", Status: statusPass, Message: "birthday config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "birthday-config-command-sync", Status: statusSkipped, Message: "birthday config command sync include is disabled"}
}

func birthdayConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeBirthday, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG")
	if err != nil {
		return checkResult{Name: "birthday-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	birthdayEnabled, err := boolValue(lookup, "MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "birthday-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeBirthday && !birthdayEnabled {
		return checkResult{Name: "birthday-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true requires MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeBirthday && birthdayEnabled {
		return checkResult{Name: "birthday-config-runtime-pairing", Status: statusPass, Message: "birthday config command sync and runtime feature flag are paired"}
	}
	if birthdayEnabled {
		return checkResult{Name: "birthday-config-runtime-pairing", Status: statusWarn, Message: "birthday config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "birthday-config-runtime-pairing", Status: statusSkipped, Message: "birthday config runtime and command sync include are disabled"}
}

func announcementConfigCommandSync(lookup lookupFunc) checkResult {
	includeAnnouncement, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG")
	if err != nil {
		return checkResult{Name: "announcement-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAnnouncement {
		return checkResult{Name: "announcement-config-command-sync", Status: statusPass, Message: "announcement config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "announcement-config-command-sync", Status: statusSkipped, Message: "announcement config command sync include is disabled"}
}

func announcementConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeAnnouncement, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG")
	if err != nil {
		return checkResult{Name: "announcement-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	announcementEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "announcement-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAnnouncement && !announcementEnabled {
		return checkResult{Name: "announcement-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true requires MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeAnnouncement && announcementEnabled {
		return checkResult{Name: "announcement-config-runtime-pairing", Status: statusPass, Message: "announcement config command sync and runtime feature flag are paired"}
	}
	if announcementEnabled {
		return checkResult{Name: "announcement-config-runtime-pairing", Status: statusWarn, Message: "announcement config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "announcement-config-runtime-pairing", Status: statusSkipped, Message: "announcement config runtime and command sync include are disabled"}
}

func announcementSendCommandSync(lookup lookupFunc) checkResult {
	includeAnnouncement, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND")
	if err != nil {
		return checkResult{Name: "announcement-send-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAnnouncement {
		return checkResult{Name: "announcement-send-command-sync", Status: statusPass, Message: "announcement send command sync include is enabled for staging review"}
	}
	return checkResult{Name: "announcement-send-command-sync", Status: statusSkipped, Message: "announcement send command sync include is disabled"}
}

func announcementSendRuntimePairing(lookup lookupFunc) checkResult {
	includeAnnouncement, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND")
	if err != nil {
		return checkResult{Name: "announcement-send-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	announcementEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED")
	if err != nil {
		return checkResult{Name: "announcement-send-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAnnouncement && !announcementEnabled {
		return checkResult{Name: "announcement-send-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true requires MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true in the staging runtime"}
	}
	if includeAnnouncement && announcementEnabled {
		return checkResult{Name: "announcement-send-runtime-pairing", Status: statusPass, Message: "announcement send command sync and runtime feature flag are paired"}
	}
	if announcementEnabled {
		return checkResult{Name: "announcement-send-runtime-pairing", Status: statusWarn, Message: "announcement send runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "announcement-send-runtime-pairing", Status: statusSkipped, Message: "announcement send runtime and command sync include are disabled"}
}

func announcementRelayRuntimeReadiness(lookup lookupFunc) checkResult {
	relayEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED")
	if err != nil {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	if !relayEnabled {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusSkipped, Message: "announcement relay runtime is disabled"}
	}
	gatewayEnabled, err := boolValue(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY")
	if err != nil {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	guildMessages, err := boolValue(lookup, "MHCAT_DISCORD_GUILD_MESSAGES_INTENT")
	if err != nil {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	messageContent, err := boolValue(lookup, "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT")
	if err != nil {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	var missing []string
	if !gatewayEnabled {
		missing = append(missing, "MHCAT_DISCORD_ENABLE_GATEWAY=true")
	}
	if !guildMessages {
		missing = append(missing, "MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true")
	}
	if !messageContent {
		missing = append(missing, "MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true")
	}
	if len(missing) > 0 {
		return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusFail, Message: "announcement relay requires " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "announcement-relay-runtime-readiness", Status: statusPass, Message: "announcement relay gateway and privileged intent gates are explicit"}
}

func textXPConfigCommandSync(lookup lookupFunc) checkResult {
	includeTextXP, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG")
	if err != nil {
		return checkResult{Name: "text-xp-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeTextXP {
		return checkResult{Name: "text-xp-config-command-sync", Status: statusPass, Message: "text XP config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "text-xp-config-command-sync", Status: statusSkipped, Message: "text XP config command sync include is disabled"}
}

func textXPConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeTextXP, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG")
	if err != nil {
		return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	textXPEnabled, err := boolValue(lookup, "MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeTextXP && !textXPEnabled {
		return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true requires MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeTextXP && textXPEnabled {
		return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusPass, Message: "text XP config command sync and runtime feature flag are paired"}
	}
	if textXPEnabled {
		return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusWarn, Message: "text XP config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "text-xp-config-runtime-pairing", Status: statusSkipped, Message: "text XP config runtime and command sync include are disabled"}
}

func voiceXPConfigCommandSync(lookup lookupFunc) checkResult {
	includeVoiceXP, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG")
	if err != nil {
		return checkResult{Name: "voice-xp-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeVoiceXP {
		return checkResult{Name: "voice-xp-config-command-sync", Status: statusPass, Message: "voice XP config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "voice-xp-config-command-sync", Status: statusSkipped, Message: "voice XP config command sync include is disabled"}
}

func voiceXPConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeVoiceXP, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG")
	if err != nil {
		return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	voiceXPEnabled, err := boolValue(lookup, "MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeVoiceXP && !voiceXPEnabled {
		return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true requires MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeVoiceXP && voiceXPEnabled {
		return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusPass, Message: "voice XP config command sync and runtime feature flag are paired"}
	}
	if voiceXPEnabled {
		return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusWarn, Message: "voice XP config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "voice-xp-config-runtime-pairing", Status: statusSkipped, Message: "voice XP config runtime and command sync include are disabled"}
}

func joinRoleConfigCommandSync(lookup lookupFunc) checkResult {
	includeJoinRole, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG")
	if err != nil {
		return checkResult{Name: "join-role-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeJoinRole {
		return checkResult{Name: "join-role-config-command-sync", Status: statusPass, Message: "join-role config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "join-role-config-command-sync", Status: statusSkipped, Message: "join-role config command sync include is disabled"}
}

func joinRoleConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeJoinRole, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG")
	if err != nil {
		return checkResult{Name: "join-role-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	joinRoleEnabled, err := boolValue(lookup, "MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "join-role-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeJoinRole && !joinRoleEnabled {
		return checkResult{Name: "join-role-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true requires MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeJoinRole && joinRoleEnabled {
		return checkResult{Name: "join-role-config-runtime-pairing", Status: statusPass, Message: "join-role config command sync and runtime feature flag are paired"}
	}
	if joinRoleEnabled {
		return checkResult{Name: "join-role-config-runtime-pairing", Status: statusWarn, Message: "join-role config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "join-role-config-runtime-pairing", Status: statusSkipped, Message: "join-role config runtime and command sync include are disabled"}
}

func joinRoleAssignmentRuntimeReadiness(lookup lookupFunc) checkResult {
	assignmentEnabled, err := boolValue(lookup, "MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED")
	if err != nil {
		return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	if !assignmentEnabled {
		return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusSkipped, Message: "join-role assignment runtime is disabled"}
	}
	gatewayEnabled, err := boolValue(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY")
	if err != nil {
		return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	guildMembers, err := boolValue(lookup, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT")
	if err != nil {
		return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	var missing []string
	if !gatewayEnabled {
		missing = append(missing, "MHCAT_DISCORD_ENABLE_GATEWAY=true")
	}
	if !guildMembers {
		missing = append(missing, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true")
	}
	if len(missing) > 0 {
		return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusFail, Message: "join-role assignment requires " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "join-role-assignment-runtime-readiness", Status: statusPass, Message: "join-role assignment gateway and Guild Members intent gates are explicit"}
}

func leaveMessageDeliveryRuntimeReadiness(lookup lookupFunc) checkResult {
	deliveryEnabled, err := boolValue(lookup, "MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED")
	if err != nil {
		return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	if !deliveryEnabled {
		return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusSkipped, Message: "leave-message delivery runtime is disabled"}
	}
	gatewayEnabled, err := boolValue(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY")
	if err != nil {
		return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	guildMembers, err := boolValue(lookup, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT")
	if err != nil {
		return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	var missing []string
	if !gatewayEnabled {
		missing = append(missing, "MHCAT_DISCORD_ENABLE_GATEWAY=true")
	}
	if !guildMembers {
		missing = append(missing, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true")
	}
	if len(missing) > 0 {
		return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusFail, Message: "leave-message delivery requires " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "leave-message-delivery-runtime-readiness", Status: statusPass, Message: "leave-message delivery gateway and Guild Members intent gates are explicit"}
}

func welcomeMessageDeliveryRuntimeReadiness(lookup lookupFunc) checkResult {
	deliveryEnabled, err := boolValue(lookup, "MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED")
	if err != nil {
		return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	if !deliveryEnabled {
		return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusSkipped, Message: "welcome-message delivery runtime is disabled"}
	}
	gatewayEnabled, err := boolValue(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY")
	if err != nil {
		return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	guildMembers, err := boolValue(lookup, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT")
	if err != nil {
		return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	var missing []string
	if !gatewayEnabled {
		missing = append(missing, "MHCAT_DISCORD_ENABLE_GATEWAY=true")
	}
	if !guildMembers {
		missing = append(missing, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true")
	}
	if len(missing) > 0 {
		return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusFail, Message: "welcome-message delivery requires " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "welcome-message-delivery-runtime-readiness", Status: statusPass, Message: "welcome-message delivery gateway and Guild Members intent gates are explicit"}
}

func welcomeMessageConfigCommandSync(lookup lookupFunc) checkResult {
	includeWelcome, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG")
	if err != nil {
		return checkResult{Name: "welcome-message-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeWelcome {
		return checkResult{Name: "welcome-message-config-command-sync", Status: statusPass, Message: "welcome-message config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "welcome-message-config-command-sync", Status: statusSkipped, Message: "welcome-message config command sync include is disabled"}
}

func welcomeMessageConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeWelcome, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG")
	if err != nil {
		return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	welcomeEnabled, err := boolValue(lookup, "MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeWelcome && !welcomeEnabled {
		return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true requires MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeWelcome && welcomeEnabled {
		return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusPass, Message: "welcome-message config command sync and runtime feature flag are paired"}
	}
	if welcomeEnabled {
		return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusWarn, Message: "welcome-message config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "welcome-message-config-runtime-pairing", Status: statusSkipped, Message: "welcome-message config runtime and command sync include are disabled"}
}

func verificationConfigCommandSync(lookup lookupFunc) checkResult {
	includeVerification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG")
	if err != nil {
		return checkResult{Name: "verification-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeVerification {
		return checkResult{Name: "verification-config-command-sync", Status: statusPass, Message: "verification config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "verification-config-command-sync", Status: statusSkipped, Message: "verification config command sync include is disabled"}
}

func verificationConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeVerification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG")
	if err != nil {
		return checkResult{Name: "verification-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	verificationEnabled, err := boolValue(lookup, "MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "verification-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeVerification && !verificationEnabled {
		return checkResult{Name: "verification-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true requires MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeVerification && verificationEnabled {
		return checkResult{Name: "verification-config-runtime-pairing", Status: statusPass, Message: "verification config command sync and runtime feature flag are paired"}
	}
	if verificationEnabled {
		return checkResult{Name: "verification-config-runtime-pairing", Status: statusWarn, Message: "verification config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "verification-config-runtime-pairing", Status: statusSkipped, Message: "verification config runtime and command sync include are disabled"}
}

func verificationFlowCommandSync(lookup lookupFunc) checkResult {
	includeVerification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW")
	if err != nil {
		return checkResult{Name: "verification-flow-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeVerification {
		return checkResult{Name: "verification-flow-command-sync", Status: statusPass, Message: "verification flow command sync include is enabled for staging review"}
	}
	return checkResult{Name: "verification-flow-command-sync", Status: statusSkipped, Message: "verification flow command sync include is disabled"}
}

func verificationFlowRuntimePairing(lookup lookupFunc) checkResult {
	includeVerification, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW")
	if err != nil {
		return checkResult{Name: "verification-flow-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	verificationEnabled, err := boolValue(lookup, "MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED")
	if err != nil {
		return checkResult{Name: "verification-flow-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeVerification && !verificationEnabled {
		return checkResult{Name: "verification-flow-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true requires MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true in the staging runtime"}
	}
	if includeVerification && verificationEnabled {
		return checkResult{Name: "verification-flow-runtime-pairing", Status: statusPass, Message: "verification flow command sync and runtime feature flag are paired"}
	}
	if verificationEnabled {
		return checkResult{Name: "verification-flow-runtime-pairing", Status: statusWarn, Message: "verification flow runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "verification-flow-runtime-pairing", Status: statusSkipped, Message: "verification flow runtime and command sync include are disabled"}
}

func accountAgeConfigCommandSync(lookup lookupFunc) checkResult {
	includeAccountAge, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG")
	if err != nil {
		return checkResult{Name: "account-age-config-command-sync", Status: statusFail, Message: err.Error()}
	}
	if includeAccountAge {
		return checkResult{Name: "account-age-config-command-sync", Status: statusPass, Message: "account-age config command sync include is enabled for staging review"}
	}
	return checkResult{Name: "account-age-config-command-sync", Status: statusSkipped, Message: "account-age config command sync include is disabled"}
}

func accountAgeConfigRuntimePairing(lookup lookupFunc) checkResult {
	includeAccountAge, err := boolValue(lookup, "MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG")
	if err != nil {
		return checkResult{Name: "account-age-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	accountAgeEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED")
	if err != nil {
		return checkResult{Name: "account-age-config-runtime-pairing", Status: statusFail, Message: err.Error()}
	}
	if includeAccountAge && !accountAgeEnabled {
		return checkResult{Name: "account-age-config-runtime-pairing", Status: statusFail, Message: "MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true requires MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true in the staging runtime"}
	}
	if includeAccountAge && accountAgeEnabled {
		return checkResult{Name: "account-age-config-runtime-pairing", Status: statusPass, Message: "account-age config command sync and runtime feature flag are paired"}
	}
	if accountAgeEnabled {
		return checkResult{Name: "account-age-config-runtime-pairing", Status: statusWarn, Message: "account-age config runtime is enabled but command sync include is disabled"}
	}
	return checkResult{Name: "account-age-config-runtime-pairing", Status: statusSkipped, Message: "account-age config runtime and command sync include are disabled"}
}

func accountAgePolicyRuntimeReadiness(lookup lookupFunc) checkResult {
	policyEnabled, err := boolValue(lookup, "MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED")
	if err != nil {
		return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	if !policyEnabled {
		return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusSkipped, Message: "account-age policy runtime is disabled"}
	}
	gatewayEnabled, err := boolValue(lookup, "MHCAT_DISCORD_ENABLE_GATEWAY")
	if err != nil {
		return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	guildMembers, err := boolValue(lookup, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT")
	if err != nil {
		return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusFail, Message: err.Error()}
	}
	var missing []string
	if !gatewayEnabled {
		missing = append(missing, "MHCAT_DISCORD_ENABLE_GATEWAY=true")
	}
	if !guildMembers {
		missing = append(missing, "MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true")
	}
	if len(missing) > 0 {
		return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusFail, Message: "account-age policy requires " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "account-age-policy-runtime-readiness", Status: statusPass, Message: "account-age policy gateway and Guild Members intent gates are explicit"}
}

func boolValue(lookup lookupFunc, key string) (bool, error) {
	value, ok := lookup(key)
	value = strings.TrimSpace(strings.ToLower(value))
	if !ok || value == "" || value == "false" || value == "0" {
		return false, nil
	}
	if value == "true" || value == "1" {
		return true, nil
	}
	return false, fmt.Errorf("%s must be a boolean", key)
}

func applicationPin(lookup lookupFunc) checkResult {
	pin, pinOK := lookup("MHCAT_STAGING_ALLOWED_APPLICATION_ID")
	app, appOK := lookup("MHCAT_DISCORD_APPLICATION_ID")
	pin = strings.TrimSpace(pin)
	app = strings.TrimSpace(app)
	if !pinOK || pin == "" {
		return checkResult{Name: "application-id-pin", Status: statusSkipped, Message: "MHCAT_STAGING_ALLOWED_APPLICATION_ID is not set"}
	}
	if !appOK || app == "" {
		return checkResult{Name: "application-id-pin", Status: statusFail, Message: "MHCAT_DISCORD_APPLICATION_ID is required to verify application pin"}
	}
	if pin != app {
		return checkResult{Name: "application-id-pin", Status: statusFail, Message: "MHCAT_DISCORD_APPLICATION_ID does not match MHCAT_STAGING_ALLOWED_APPLICATION_ID"}
	}
	return checkResult{Name: "application-id-pin", Status: statusPass, Message: "application id matches staging pin"}
}

func formatReport(w io.Writer, result report, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	for _, check := range result.Checks {
		if _, err := fmt.Fprintf(w, "%s status=%s message=%q\n", check.Name, check.Status, check.Message); err != nil {
			return err
		}
	}
	return nil
}

func hasFailures(checks []checkResult) bool {
	for _, check := range checks {
		if check.Status == statusFail {
			return true
		}
	}
	return false
}
