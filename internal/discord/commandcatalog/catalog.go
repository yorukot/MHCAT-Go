package commandcatalog

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureannouncements "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/announcements"
	featureautochat "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/autochat"
	featurebalance "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/balance"
	featurebirthday "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/birthday"
	featureeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/economy"
	featuregacha "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/gacha"
	featurelogging "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/logging"
	featurelottery "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/lottery"
	featuremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/moderation"
	featurenotifications "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/notifications"
	featureonboarding "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/onboarding"
	featurepoll "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/poll"
	featureredeem "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/redeem"
	featureroles "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/roles"
	featuresafety "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/safety"
	featurestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/stats"
	featureticket "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/ticket"
	featurevoice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/voice"
	featurework "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/work"
	featurexp "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/xp"
)

// AllDefinitions returns the complete legacy-compatible command inventory,
// independent of runtime feature gates.
func AllDefinitions() []commands.Definition {
	definitions := commands.BuiltinDefinitions()
	definitions = append(definitions, commands.TranslateDefinition())
	definitions = append(definitions, featureticket.Definitions()...)
	definitions = append(definitions, featurepoll.Definitions()...)
	definitions = append(definitions, featureeconomy.Definitions()...)
	definitions = append(definitions, featureeconomy.SignInDefinitions()...)
	definitions = append(definitions, featureeconomy.SettingsDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinAdminDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinRankDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinResetDefinitions()...)
	definitions = append(definitions, featureeconomy.RockPaperScissorsDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinGameDefinitions()...)
	definitions = append(definitions, featureeconomy.ShopDefinitions()...)
	definitions = append(definitions, featureeconomy.ProfileDefinitions()...)
	definitions = append(definitions, featurework.Definitions()...)
	definitions = append(definitions, featuremoderation.Definitions()...)
	definitions = append(definitions, featuremoderation.SettingsDefinitions()...)
	definitions = append(definitions, featuremoderation.RemovalDefinitions()...)
	definitions = append(definitions, featuremoderation.IssueDefinitions()...)
	definitions = append(definitions, featuremoderation.CleanupDefinitions()...)
	definitions = append(definitions, featuremoderation.DeleteDataDefinitions()...)
	definitions = append(definitions, featurebalance.Definitions()...)
	definitions = append(definitions, featureredeem.Definitions()...)
	definitions = append(definitions, featureautochat.Definitions()...)
	definitions = append(definitions, featurenotifications.Definitions()...)
	definitions = append(definitions, featuresafety.Definitions()...)
	definitions = append(definitions, featurelogging.Definitions()...)
	definitions = append(definitions, featuregacha.AllDefinitions()...)
	definitions = append(definitions, featurelottery.Definitions()...)
	definitions = append(definitions, featurestats.Definitions()...)
	definitions = append(definitions, featurebirthday.Definitions()...)
	definitions = append(definitions, featureannouncements.ConfigDefinitions()...)
	definitions = append(definitions, featureannouncements.SendDefinitions()...)
	definitions = append(definitions, featurexp.TextDefinitions()...)
	definitions = append(definitions, featurexp.VoiceDefinitions()...)
	definitions = append(definitions, featurexp.RewardRoleDefinitions()...)
	definitions = append(definitions, featurexp.DisabledProfileDefinitions()...)
	definitions = append(definitions, featurexp.AdminDefinitions()...)
	definitions = append(definitions, featurexp.ResetDefinitions()...)
	definitions = append(definitions, featurexp.RankDefinitions()...)
	definitions = append(definitions, featurevoice.Definitions()...)
	definitions = append(definitions, featurevoice.LockDefinitions()...)
	definitions = append(definitions, featureonboarding.JoinRoleDefinitions()...)
	definitions = append(definitions, featureonboarding.MessageDefinitions()...)
	definitions = append(definitions, featureonboarding.VerificationDefinitions()...)
	definitions = append(definitions, featureonboarding.VerificationFlowDefinitions()...)
	definitions = append(definitions, featureonboarding.AccountAgeDefinitions()...)
	definitions = append(definitions, featureroles.Definitions()...)
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "all-commands"}, definitions)
	return registry.Commands
}
