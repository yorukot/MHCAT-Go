package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	discordgoadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureannouncements "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/announcements"
	featureautochat "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/autochat"
	featurebirthday "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/birthday"
	featureeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/economy"
	featuregacha "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/gacha"
	featurelogging "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/logging"
	featurelottery "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/lottery"
	featuremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/moderation"
	featureonboarding "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/onboarding"
	featurepoll "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/poll"
	featuresafety "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/safety"
	featurestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/stats"
	featureticket "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/ticket"
	featurework "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/work"
	featurexp "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
)

type clientFactory func(config.CommandSyncConfig) (commands.SyncClient, error)
type registryLoader func(config.CommandSyncConfig, commands.Scope) commands.Registry

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr, newDiscordClient, defaultCommandRegistry))
}

func run(
	ctx context.Context,
	args []string,
	lookup config.LookupFunc,
	stdout io.Writer,
	stderr io.Writer,
	newClient clientFactory,
	loadRegistry registryLoader,
) int {
	cfg, err := config.LoadCommandSyncRawWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "command sync config error: %v\n", err)
		return 1
	}
	if err := applyFlags(args, &cfg, stderr); err != nil {
		fmt.Fprintf(stderr, "command sync flag error: %v\n", err)
		return 1
	}
	if err := config.ValidateCommandSync(cfg); err != nil {
		fmt.Fprintf(stderr, "command sync config error: %v\n", err)
		return 1
	}

	logger := observability.NewLogger(observability.LoggerOptions{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Writer: stderr,
	})
	for _, warning := range cfg.AliasWarnings {
		fields := warning.RedactedFields()
		logger.WarnContext(ctx, warning.Message(), cfgAliasAttrs(fields)...)
	}

	scope := commands.Scope{Kind: cfg.Scope, GuildID: cfg.GuildID}
	registry := loadRegistry(cfg, scope)
	if cfg.Staging.Mode || !cfg.DryRun {
		if err := commands.ValidateStagingSync(registry, commands.StagingSyncOptions{
			Scope:              scope,
			ExpectedCommands:   expectedStagingCommands(cfg),
			AllowDelete:        cfg.AllowDelete,
			AllowBulkOverwrite: cfg.AllowBulkOverwrite,
		}); err != nil {
			fmt.Fprintf(stderr, "command sync staging safety error: %v\n", err)
			return 1
		}
	}
	client, err := newClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "command sync client error: %v\n", err)
		return 1
	}
	plan, err := commands.PlanSync(ctx, client, registry, commands.SyncOptions{
		Scope:              scope,
		DryRun:             cfg.DryRun,
		AllowDelete:        cfg.AllowDelete,
		AllowBulkOverwrite: cfg.AllowBulkOverwrite,
	})
	if err != nil {
		fmt.Fprintf(stderr, "command sync plan error: %v\n", err)
		return 1
	}
	if err := commands.FormatPlan(stdout, plan, cfg.Format); err != nil {
		fmt.Fprintf(stderr, "command sync output error: %v\n", err)
		return 1
	}
	if cfg.DryRun {
		fmt.Fprintln(stderr, "command sync dry-run: no Discord command writes performed")
		return 0
	}
	result, err := commands.ExecutePlan(ctx, client, registry, plan, commands.SyncOptions{
		Scope:              scope,
		DryRun:             cfg.DryRun,
		AllowDelete:        cfg.AllowDelete,
		AllowBulkOverwrite: cfg.AllowBulkOverwrite,
	})
	if err != nil {
		fmt.Fprintf(stderr, "command sync apply error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "command sync apply complete: writes=%d\n", result.Writes)
	return 0
}

func applyFlags(args []string, cfg *config.CommandSyncConfig, stderr io.Writer) error {
	flags := flag.NewFlagSet("mhcat-command-sync", flag.ContinueOnError)
	flags.SetOutput(stderr)
	scope := flags.String("scope", cfg.Scope, "command sync scope: guild or global")
	guildID := flags.String("guild-id", cfg.GuildID, "guild ID for guild-scoped command sync")
	dryRun := flags.Bool("dry-run", true, "print the sync plan without writing Discord commands")
	apply := flags.Bool("apply", false, "apply create/update operations")
	allowDelete := flags.Bool("allow-delete", false, "allow deleting owned remote commands; requires --apply")
	allowBulkOverwrite := flags.Bool("allow-bulk-overwrite", false, "allow bulk overwrite path; requires --apply")
	strict := flags.Bool("strict", cfg.Strict, "fail on strict registry validation errors")
	format := flags.String("format", cfg.Format, "output format: text or json")
	if err := flags.Parse(args); err != nil {
		return err
	}

	dryRunSet := flagWasSet(args, "dry-run")
	if *apply && dryRunSet && *dryRun {
		return fmt.Errorf("--apply and --dry-run=true cannot be used together")
	}
	if !*apply && dryRunSet && !*dryRun {
		return fmt.Errorf("apply mode requires explicit --apply")
	}

	cfg.Scope = strings.TrimSpace(*scope)
	cfg.GuildID = strings.TrimSpace(*guildID)
	cfg.Strict = *strict
	cfg.Format = strings.TrimSpace(*format)
	cfg.DryRun = true
	if *apply {
		cfg.DryRun = false
	}
	if dryRunSet && *dryRun {
		cfg.DryRun = true
	}
	cfg.AllowDelete = *allowDelete
	cfg.AllowBulkOverwrite = *allowBulkOverwrite
	return nil
}

func flagWasSet(args []string, name string) bool {
	prefix := "--" + name
	for _, arg := range args {
		if arg == prefix || strings.HasPrefix(arg, prefix+"=") {
			return true
		}
	}
	return false
}

func newDiscordClient(cfg config.CommandSyncConfig) (commands.SyncClient, error) {
	return discordgoadapter.NewCommandSyncClient(cfg.DiscordToken, cfg.ApplicationID)
}

func defaultCommandRegistry(cfg config.CommandSyncConfig, scope commands.Scope) commands.Registry {
	definitions := commands.BuiltinDefinitions()
	if cfg.IncludeTickets {
		definitions = append(definitions, featureticket.Definitions()...)
	}
	if cfg.IncludePolls {
		definitions = append(definitions, featurepoll.Definitions()...)
	}
	if cfg.IncludeEconomyQuery {
		definitions = append(definitions, featureeconomy.Definitions()...)
	}
	if cfg.IncludeEconomySignIn {
		definitions = append(definitions, featureeconomy.SignInDefinitions()...)
	}
	if cfg.IncludeEconomySettings {
		definitions = append(definitions, featureeconomy.SettingsDefinitions()...)
	}
	if cfg.IncludeWork {
		definitions = append(definitions, featurework.Definitions()...)
	}
	if cfg.IncludeWarnings {
		definitions = append(definitions, featuremoderation.Definitions()...)
	}
	if cfg.IncludeTranslate {
		definitions = append(definitions, commands.TranslateDefinition())
	}
	if cfg.IncludeAutoChatConfig {
		definitions = append(definitions, featureautochat.Definitions()...)
	}
	if cfg.IncludeAntiScamConfig {
		definitions = append(definitions, featuresafety.Definitions()...)
	}
	if cfg.IncludeLoggingConfig {
		definitions = append(definitions, featurelogging.Definitions()...)
	}
	if cfg.IncludeGachaPrizeList {
		definitions = append(definitions, featuregacha.Definitions()...)
	}
	if cfg.IncludeLotteryDisabledCommand {
		definitions = append(definitions, featurelottery.Definitions()...)
	}
	if cfg.IncludeStatsQuery {
		definitions = append(definitions, featurestats.Definitions()...)
	}
	if cfg.IncludeBirthdayConfig {
		definitions = append(definitions, featurebirthday.Definitions()...)
	}
	if cfg.IncludeAnnouncementConfig {
		definitions = append(definitions, featureannouncements.ConfigDefinitions()...)
	}
	if cfg.IncludeAnnouncementSend {
		definitions = append(definitions, featureannouncements.SendDefinitions()...)
	}
	if cfg.IncludeTextXPConfig {
		definitions = append(definitions, featurexp.TextDefinitions()...)
	}
	if cfg.IncludeVoiceXPConfig {
		definitions = append(definitions, featurexp.VoiceDefinitions()...)
	}
	if cfg.IncludeJoinRoleConfig {
		definitions = append(definitions, featureonboarding.JoinRoleDefinitions()...)
	}
	if cfg.IncludeWelcomeMessageConfig {
		definitions = append(definitions, featureonboarding.MessageDefinitions()...)
	}
	if cfg.IncludeVerificationConfig {
		definitions = append(definitions, featureonboarding.VerificationDefinitions()...)
	}
	if cfg.IncludeVerificationFlow {
		definitions = append(definitions, featureonboarding.VerificationFlowDefinitions()...)
	}
	if cfg.IncludeAccountAgeConfig {
		definitions = append(definitions, featureonboarding.AccountAgeDefinitions()...)
	}
	return commands.NewRegistry(scope, definitions)
}

func expectedStagingCommands(cfg config.CommandSyncConfig) []string {
	expected := append([]string(nil), cfg.Staging.ExpectedCommands...)
	if cfg.IncludeTickets {
		expected = append(expected, "私人頻道設置", "私人頻道刪除")
	}
	if cfg.IncludePolls {
		expected = append(expected, "投票創建")
	}
	if cfg.IncludeEconomyQuery {
		expected = append(expected, "代幣查詢")
	}
	if cfg.IncludeEconomySignIn {
		expected = append(expected, "簽到")
	}
	if cfg.IncludeEconomySettings {
		expected = append(expected, featureeconomy.EconomySettingsCommandName)
	}
	if cfg.IncludeWork {
		expected = append(expected, "打工系統")
	}
	if cfg.IncludeWarnings {
		expected = append(expected, "警告紀錄")
	}
	if cfg.IncludeTranslate {
		expected = append(expected, "翻譯")
	}
	if cfg.IncludeAutoChatConfig {
		expected = append(expected, featureautochat.AutoChatSetCommandName, featureautochat.AutoChatDeleteCommandName)
	}
	if cfg.IncludeAntiScamConfig {
		expected = append(expected, featuresafety.AntiScamCommandName)
	}
	if cfg.IncludeLoggingConfig {
		expected = append(expected, featurelogging.LoggingConfigCommandName)
	}
	if cfg.IncludeGachaPrizeList {
		expected = append(expected, featuregacha.GachaPrizeListCommandName)
	}
	if cfg.IncludeLotteryDisabledCommand {
		expected = append(expected, featurelottery.LotteryCreateCommandName)
	}
	if cfg.IncludeStatsQuery {
		expected = append(expected, featurestats.StatsQueryCommandName)
	}
	if cfg.IncludeBirthdayConfig {
		expected = append(expected, featurebirthday.BirthdayCommandName)
	}
	if cfg.IncludeAnnouncementConfig {
		expected = append(expected, featureannouncements.ConfigCommandName)
	}
	if cfg.IncludeAnnouncementSend {
		expected = append(expected, featureannouncements.SendCommandName)
	}
	if cfg.IncludeTextXPConfig {
		expected = append(expected, featurexp.TextXPSetCommandName, featurexp.TextXPDeleteCommandName)
	}
	if cfg.IncludeVoiceXPConfig {
		expected = append(expected, featurexp.VoiceXPSetCommandName, featurexp.VoiceXPDeleteCommandName)
	}
	if cfg.IncludeJoinRoleConfig {
		expected = append(expected, featureonboarding.JoinRoleSetCommandName, featureonboarding.JoinRoleDeleteCommandName)
	}
	if cfg.IncludeWelcomeMessageConfig {
		expected = append(expected, featureonboarding.JoinMessageSetCommandName, featureonboarding.LeaveMessageSetCommandName)
	}
	if cfg.IncludeVerificationConfig {
		expected = append(expected, featureonboarding.VerificationSetCommandName)
	}
	if cfg.IncludeVerificationFlow {
		expected = append(expected, featureonboarding.VerificationCommandName)
	}
	if cfg.IncludeAccountAgeConfig {
		expected = append(expected, featureonboarding.AccountAgeCommandName)
	}
	return expected
}

func cfgAliasAttrs(fields map[string]string) []any {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return attrs
}
