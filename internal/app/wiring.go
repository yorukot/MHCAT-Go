package app

import (
	"context"
	"log/slog"
	"time"

	discordadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	corenotifications "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/notifications"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	corestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/stats"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
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
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	featurevoice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/voice"
	featurework "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/work"
	featurexp "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

type RuntimeOptions struct {
	Config                        config.Config
	Logger                        *slog.Logger
	Session                       DiscordSession
	UsageTracker                  ports.UsageTracker
	TicketConfigRepository        ports.TicketConfigRepository
	TicketChannelPort             ports.DiscordChannelPort
	TicketMessagePort             ports.DiscordMessagePort
	PollRepository                ports.PollRepository
	PollMessagePort               ports.DiscordMessagePort
	PollMemberCounter             ports.DiscordGuildMemberReader
	EconomyQueryRepository        ports.EconomyQueryRepository
	EconomySignInRepository       ports.EconomySignInRepository
	EconomySettingsRepository     ports.EconomySettingsRepository
	EconomyCoinAdminRepository    ports.EconomyCoinAdminRepository
	EconomyCoinRankRepository     ports.EconomyCoinRankRepository
	EconomyCoinResetRepository    ports.EconomyCoinResetRepository
	EconomyCoinResetMessagePort   ports.DiscordMessagePort
	EconomyCoinResetGuildInfo     ports.DiscordInfoProvider
	EconomyRPSRepository          ports.EconomyRockPaperScissorsRepository
	EconomyGameRepository         ports.EconomyCoinGameRepository
	EconomyGameMessagePort        ports.DiscordMessagePort
	EconomyShopRepository         ports.EconomyShopRepository
	EconomyShopDirectMessage      ports.DiscordDirectMessagePort
	EconomyShopRolePort           ports.DiscordRolePort
	EconomyShopRoleInspector      ports.DiscordRoleInspector
	EconomyProfileRepository      ports.EconomyProfileRepository
	WorkInterfaceRepository       ports.WorkInterfaceRepository
	WorkStartRepository           ports.WorkStartRepository
	WorkAdminRepository           ports.WorkAdminRepository
	WorkFeatureEnabled            bool
	WarningHistoryRepository      ports.WarningHistoryRepository
	WarningSettingsRepository     ports.WarningSettingsRepository
	WarningRemovalRepository      ports.WarningRemovalRepository
	WarningIssueRepository        ports.WarningIssueRepository
	WarningMemberReader           ports.DiscordGuildMemberReader
	WarningRemovalDirectMessage   ports.DiscordDirectMessagePort
	WarningIssueDirectMessage     ports.DiscordDirectMessagePort
	WarningIssueMemberPort        ports.DiscordMemberPort
	WarningIssueHierarchy         ports.DiscordMemberHierarchyInspector
	WarningIssueMessagePort       ports.DiscordMessagePort
	MessageCleaner                ports.DiscordMessageCleaner
	DeleteDataRepository          ports.DeleteDataRepository
	WarningsFeatureEnabled        bool
	WarningSettingsFeatureEnabled bool
	WarningRemovalFeatureEnabled  bool
	WarningIssueFeatureEnabled    bool
	MessageCleanupFeatureEnabled  bool
	DeleteDataFeatureEnabled      bool
	TranslateProvider             ports.Translator
	TranslateFeatureEnabled       bool
	BalanceRepository             ports.BalanceRepository
	RedeemRepository              ports.RedeemRepository
	AutoChatConfigRepository      ports.AutoChatConfigRepository
	AutoNotificationRepository    ports.AutoNotificationScheduleRepository
	AutoNotificationMessagePort   ports.DiscordMessagePort
	AntiScamConfigRepository      ports.AntiScamConfigRepository
	ScamURLCatalogRepository      ports.ScamURLCatalog
	ScamReportSender              ports.ScamReportSender
	LoggingConfigRepository       ports.LoggingConfigRepository
	GachaPrizePoolRepository      ports.GachaPrizePoolRepository
	GachaDrawRepository           ports.GachaDrawRepository
	GachaDrawMessagePort          ports.DiscordMessagePort
	GachaDrawDirectMessagePort    ports.DiscordDirectMessagePort
	GachaDrawWait                 func(context.Context, time.Duration) error
	GachaPrizeCreateRepository    ports.GachaPrizeCreateRepository
	GachaPrizeEditRepository      ports.GachaPrizeEditRepository
	GachaPrizeDeleteRepository    ports.GachaPrizeDeleteRepository
	LotteryDisabledCommandEnabled bool
	LotteryComponentsEnabled      bool
	LotteryRepository             ports.LotteryRepository
	LotteryDiscordInfo            ports.DiscordInfoProvider
	LotteryMemberReader           ports.DiscordGuildMemberReader
	LotteryMessagePort            ports.DiscordMessagePort
	StatsQueryEnabled             bool
	StatsCreateRepository         ports.StatsConfigRepository
	StatsCreateChannelPort        ports.DiscordChannelPort
	StatsCreateGuildStats         ports.DiscordGuildStatsReader
	StatsRoleStatsRepository      ports.StatsConfigRepository
	StatsRoleConfigRepository     ports.StatsRoleConfigRepository
	StatsRoleChannelPort          ports.DiscordChannelPort
	StatsRoleStatsReader          ports.DiscordRoleStatsReader
	StatsDeleteRepository         ports.StatsConfigRepository
	BirthdayConfigRepository      ports.BirthdayConfigRepository
	BirthdayCachedUsers           ports.DiscordCachedUserInfoProvider
	AnnouncementConfigRepository  ports.AnnouncementConfigRepository
	AnnouncementSendRepository    ports.AnnouncementChannelReader
	AnnouncementMessagePort       ports.DiscordMessagePort
	TextXPConfigRepository        ports.TextXPConfigRepository
	TextXPMessagePort             ports.DiscordMessagePort
	VoiceXPConfigRepository       ports.VoiceXPConfigRepository
	VoiceXPMessagePort            ports.DiscordMessagePort
	TextXPRewardRoleRepository    ports.TextXPRewardRoleRepository
	VoiceXPRewardRoleRepository   ports.VoiceXPRewardRoleRepository
	XPRewardRoleInspector         ports.DiscordRoleInspector
	XPProfileDisabledEnabled      bool
	XPAdminRepository             ports.XPAdminRepository
	XPResetRepository             ports.XPResetRepository
	XPResetMessagePort            ports.DiscordMessagePort
	XPResetGuildInfo              ports.DiscordInfoProvider
	XPRankRepository              ports.XPRankRepository
	VoiceRoomConfigRepository     ports.VoiceRoomConfigRepository
	VoiceRoomLockRepository       ports.VoiceRoomLockRepository
	JoinRoleConfigRepository      ports.JoinRoleConfigRepository
	JoinRoleInspector             ports.DiscordRoleInspector
	LeaveMessageConfigRepository  ports.LeaveMessageConfigRepository
	VerificationConfigRepository  ports.VerificationConfigRepository
	VerificationRoleInspector     ports.DiscordRoleInspector
	VerificationFlowRepository    ports.VerificationConfigReader
	VerificationRolePort          ports.DiscordRolePort
	VerificationMemberPort        ports.DiscordMemberPort
	VerificationGuildInfo         ports.DiscordInfoProvider
	AccountAgeConfigRepository    ports.AccountAgeConfigRepository
	AccountAgePolicyRepository    ports.AccountAgeConfigReader
	AccountAgeDirectMessagePort   ports.DiscordDirectMessagePort
	AccountAgeMemberPort          ports.DiscordMemberPort
	AccountAgeMessagePort         ports.DiscordMessagePort
	AccountAgeGuildInfo           ports.DiscordInfoProvider
	RoleSelectionRepository       ports.RoleSelectionRepository
	RoleSelectionRolePort         ports.DiscordRolePort
	RoleSelectionRoleInspector    ports.DiscordRoleInspector
	RoleSelectionReactionPort     ports.DiscordReactionPort
	RoleSelectionMessagePort      ports.DiscordMessagePort
	RoleSelectionDirectMessage    ports.DiscordDirectMessagePort
	BotUserID                     string
	Clock                         ports.Clock
}

const translateInteractionTimeout = 10 * time.Second

func BuildRuntime(opts RuntimeOptions) (*discordruntime.Dispatcher, error) {
	concreteDiscord := discordInfoProvider(opts.Session)
	xpResetGuilds := opts.XPResetGuildInfo
	if xpResetGuilds == nil {
		xpResetGuilds = concreteDiscord
	}
	coinResetGuilds := opts.EconomyCoinResetGuildInfo
	if coinResetGuilds == nil {
		coinResetGuilds = concreteDiscord
	}
	lotteryDiscord := opts.LotteryDiscordInfo
	if lotteryDiscord == nil {
		lotteryDiscord = concreteDiscord
	}
	definitions := commands.BuiltinDefinitions()
	if opts.TicketConfigRepository != nil {
		definitions = append(definitions, featureticket.Definitions()...)
	}
	if opts.PollRepository != nil {
		definitions = append(definitions, featurepoll.Definitions()...)
	}
	if opts.EconomyQueryRepository != nil {
		definitions = append(definitions, featureeconomy.Definitions()...)
	}
	if opts.EconomySignInRepository != nil {
		definitions = append(definitions, featureeconomy.SignInDefinitions()...)
	}
	if opts.EconomySettingsRepository != nil {
		definitions = append(definitions, featureeconomy.SettingsDefinitions()...)
	}
	if opts.EconomyCoinAdminRepository != nil {
		definitions = append(definitions, featureeconomy.CoinAdminDefinitions()...)
	}
	if opts.EconomyCoinRankRepository != nil {
		definitions = append(definitions, featureeconomy.CoinRankDefinitions()...)
	}
	if coinResetRuntimeEnabled(opts, coinResetGuilds) {
		definitions = append(definitions, featureeconomy.CoinResetDefinitions()...)
	}
	if opts.EconomyRPSRepository != nil {
		definitions = append(definitions, featureeconomy.RockPaperScissorsDefinitions()...)
	}
	if opts.EconomyGameRepository != nil {
		definitions = append(definitions, featureeconomy.CoinGameDefinitions()...)
	}
	if opts.EconomyShopRepository != nil {
		definitions = append(definitions, featureeconomy.ShopDefinitions()...)
	}
	if opts.EconomyProfileRepository != nil {
		definitions = append(definitions, featureeconomy.ProfileDefinitions()...)
	}
	if opts.WorkFeatureEnabled {
		definitions = append(definitions, featurework.Definitions()...)
	}
	if opts.WarningsFeatureEnabled {
		definitions = append(definitions, featuremoderation.Definitions()...)
	}
	if opts.WarningSettingsFeatureEnabled {
		definitions = append(definitions, featuremoderation.SettingsDefinitions()...)
	}
	if opts.WarningRemovalFeatureEnabled {
		definitions = append(definitions, featuremoderation.RemovalDefinitions()...)
	}
	if opts.WarningIssueFeatureEnabled {
		definitions = append(definitions, featuremoderation.IssueDefinitions()...)
	}
	if opts.MessageCleanupFeatureEnabled && opts.MessageCleaner != nil {
		definitions = append(definitions, featuremoderation.CleanupDefinitions()...)
	}
	if opts.DeleteDataFeatureEnabled && opts.DeleteDataRepository != nil {
		definitions = append(definitions, featuremoderation.DeleteDataDefinitions()...)
	}
	if opts.TranslateFeatureEnabled && opts.TranslateProvider != nil {
		definitions = append(definitions, commands.TranslateDefinition())
	}
	if opts.BalanceRepository != nil {
		definitions = append(definitions, featurebalance.Definitions()...)
	}
	if opts.RedeemRepository != nil {
		definitions = append(definitions, featureredeem.Definitions()...)
	}
	if opts.AutoChatConfigRepository != nil {
		definitions = append(definitions, featureautochat.Definitions()...)
	}
	if opts.AutoNotificationRepository != nil {
		definitions = append(definitions, featurenotifications.Definitions()...)
	}
	if opts.AntiScamConfigRepository != nil {
		definitions = append(definitions, featuresafety.ConfigDefinitions()...)
	}
	if opts.ScamURLCatalogRepository != nil && opts.ScamReportSender != nil {
		definitions = append(definitions, featuresafety.ReportDefinitions()...)
	}
	if opts.LoggingConfigRepository != nil {
		definitions = append(definitions, featurelogging.Definitions()...)
	}
	if opts.GachaPrizePoolRepository != nil {
		definitions = append(definitions, featuregacha.PrizeListDefinitions()...)
	}
	if opts.GachaDrawRepository != nil {
		definitions = append(definitions, featuregacha.DrawDefinitions()...)
	}
	if opts.GachaPrizeCreateRepository != nil {
		definitions = append(definitions, featuregacha.PrizeCreateDefinitions()...)
	}
	if opts.GachaPrizeEditRepository != nil {
		definitions = append(definitions, featuregacha.PrizeEditDefinitions()...)
	}
	if opts.GachaPrizeDeleteRepository != nil {
		definitions = append(definitions, featuregacha.PrizeDeleteDefinitions()...)
	}
	if opts.LotteryDisabledCommandEnabled {
		definitions = append(definitions, featurelottery.Definitions()...)
	}
	if opts.StatsQueryEnabled {
		definitions = append(definitions, featurestats.QueryDefinitions()...)
	}
	if opts.StatsCreateRepository != nil && opts.StatsCreateChannelPort != nil && opts.StatsCreateGuildStats != nil {
		definitions = append(definitions, featurestats.CreateDefinitions()...)
	}
	if opts.StatsRoleStatsRepository != nil && opts.StatsRoleConfigRepository != nil && opts.StatsRoleChannelPort != nil && opts.StatsRoleStatsReader != nil {
		definitions = append(definitions, featurestats.RoleDefinitions()...)
	}
	if opts.StatsDeleteRepository != nil {
		definitions = append(definitions, featurestats.DeleteDefinitions()...)
	}
	if opts.BirthdayConfigRepository != nil {
		definitions = append(definitions, featurebirthday.Definitions()...)
	}
	if opts.AnnouncementConfigRepository != nil {
		definitions = append(definitions, featureannouncements.ConfigDefinitions()...)
	}
	if opts.AnnouncementSendRepository != nil && opts.AnnouncementMessagePort != nil {
		definitions = append(definitions, featureannouncements.SendDefinitions()...)
	}
	if opts.TextXPConfigRepository != nil {
		definitions = append(definitions, featurexp.TextDefinitions()...)
	}
	if opts.VoiceXPConfigRepository != nil {
		definitions = append(definitions, featurexp.VoiceDefinitions()...)
	}
	if opts.TextXPRewardRoleRepository != nil && opts.VoiceXPRewardRoleRepository != nil {
		definitions = append(definitions, featurexp.RewardRoleDefinitions()...)
	}
	if opts.XPProfileDisabledEnabled {
		definitions = append(definitions, featurexp.DisabledProfileDefinitions()...)
	}
	if opts.XPAdminRepository != nil {
		definitions = append(definitions, featurexp.AdminDefinitions()...)
	}
	if xpResetRuntimeEnabled(opts, xpResetGuilds) {
		definitions = append(definitions, featurexp.ResetDefinitions()...)
	}
	if opts.XPRankRepository != nil {
		definitions = append(definitions, featurexp.RankDefinitions()...)
	}
	if opts.VoiceRoomConfigRepository != nil {
		definitions = append(definitions, featurevoice.Definitions()...)
	}
	if opts.VoiceRoomLockRepository != nil {
		definitions = append(definitions, featurevoice.LockDefinitions()...)
	}
	if opts.JoinRoleConfigRepository != nil {
		definitions = append(definitions, featureonboarding.JoinRoleDefinitions()...)
	}
	if opts.LeaveMessageConfigRepository != nil {
		definitions = append(definitions, featureonboarding.MessageDefinitions()...)
	}
	if opts.VerificationConfigRepository != nil {
		definitions = append(definitions, featureonboarding.VerificationDefinitions()...)
	}
	if verificationFlowRuntimeEnabled(opts) {
		definitions = append(definitions, featureonboarding.VerificationFlowDefinitions()...)
	}
	if opts.AccountAgeConfigRepository != nil {
		definitions = append(definitions, featureonboarding.AccountAgeDefinitions()...)
	}
	if opts.RoleSelectionRepository != nil {
		definitions = append(definitions, featureroles.Definitions()...)
	}
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, definitions)
	interactionTimeout := opts.Config.DiscordInteractionTimeout
	if opts.TranslateFeatureEnabled && interactionTimeout < translateInteractionTimeout {
		interactionTimeout = translateInteractionTimeout
	}
	router := interactions.NewRouter(
		interactions.Recover(),
		interactions.Timeout(interactionTimeout),
		interactions.Usage(opts.UsageTracker),
		interactions.Permission(interactions.AllowAllPermissions()),
		interactions.Logging(opts.Logger),
	)
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	runtimeShutdowns := []discordruntime.ShutdownFunc{}
	// Runtime usage belongs to the slash middleware; route-level tracking would double count.
	opts.UsageTracker = nil

	module := featureutility.NewModuleWithDiscordInfo(registry, botInfoProvider(opts.Session), concreteDiscord, clockOrSystem(opts.Clock), nil)
	if opts.TranslateFeatureEnabled && opts.TranslateProvider != nil {
		module = featureutility.NewModuleWithTranslator(registry, botInfoProvider(opts.Session), concreteDiscord, opts.TranslateProvider, clockOrSystem(opts.Clock), nil)
	}
	if err := module.RegisterRoutes(router); err != nil {
		return nil, err
	}
	if opts.TicketConfigRepository != nil {
		ticketModule := featureticket.NewModuleWithSideEffects(opts.TicketConfigRepository, opts.UsageTracker, opts.TicketChannelPort, opts.TicketMessagePort, opts.BotUserID)
		if err := ticketModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.PollRepository != nil {
		pollModule := featurepoll.NewModuleWithSideEffects(opts.PollRepository, opts.UsageTracker, opts.PollMessagePort, opts.PollMemberCounter, clockOrSystem(opts.Clock))
		if err := pollModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomySignInRepository != nil {
		var economyModule featureeconomy.Module
		if opts.EconomyQueryRepository != nil {
			economyModule = featureeconomy.NewModuleWithSignIn(opts.EconomyQueryRepository, opts.EconomySignInRepository, concreteDiscord, clockOrSystem(opts.Clock), opts.UsageTracker)
		} else {
			economyModule = featureeconomy.NewSignInOnlyModule(opts.EconomySignInRepository, concreteDiscord, clockOrSystem(opts.Clock), opts.UsageTracker)
		}
		if err := economyModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	} else if opts.EconomyQueryRepository != nil {
		economyModule := featureeconomy.NewModule(opts.EconomyQueryRepository, concreteDiscord, opts.UsageTracker)
		if err := economyModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomySettingsRepository != nil {
		settingsModule := featureeconomy.NewSettingsModule(opts.EconomySettingsRepository, concreteDiscord, opts.UsageTracker)
		if err := settingsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomyCoinAdminRepository != nil {
		coinAdminModule := featureeconomy.NewCoinAdminModule(opts.EconomyCoinAdminRepository, concreteDiscord, opts.UsageTracker)
		if err := coinAdminModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomyCoinRankRepository != nil {
		coinRankModule := featureeconomy.NewCoinRankModule(opts.EconomyCoinRankRepository, concreteDiscord, opts.UsageTracker)
		if err := coinRankModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if coinResetRuntimeEnabled(opts, coinResetGuilds) {
		coinResetModule := featureeconomy.NewCoinResetModule(opts.EconomyCoinResetRepository, coinResetGuilds, opts.EconomyCoinResetMessagePort, opts.UsageTracker, nil)
		if err := coinResetModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomyRPSRepository != nil {
		rpsModule := featureeconomy.NewRockPaperScissorsModule(opts.EconomyRPSRepository, concreteDiscord, opts.UsageTracker)
		if err := rpsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomyGameRepository != nil {
		gameModule := featureeconomy.NewCoinGameModuleWithMessages(
			opts.EconomyGameRepository,
			concreteDiscord,
			opts.EconomyGameMessagePort,
			opts.UsageTracker,
			clockOrSystem(opts.Clock),
		).WithLogger(opts.Logger)
		if err := gameModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
		runtimeShutdowns = append(runtimeShutdowns, gameModule.StopCoinGameLifecycle)
	}
	if opts.EconomyShopRepository != nil {
		shopModule := featureeconomy.NewShopModule(
			opts.EconomyShopRepository,
			concreteDiscord,
			opts.EconomyShopRoleInspector,
			opts.EconomyShopRolePort,
			opts.EconomyShopDirectMessage,
			opts.UsageTracker,
			clockOrSystem(opts.Clock),
		)
		if err := shopModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.EconomyProfileRepository != nil {
		profileModule := featureeconomy.NewProfileModule(opts.EconomyProfileRepository, concreteDiscord, clockOrSystem(opts.Clock), opts.UsageTracker)
		if err := profileModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.WorkFeatureEnabled {
		workModule := featurework.NewModule(opts.UsageTracker)
		if opts.WorkAdminRepository != nil {
			workModule = featurework.NewModuleWithAdminRepositoryAndDiscordInfo(opts.WorkAdminRepository, concreteDiscord, opts.UsageTracker, clockOrSystem(opts.Clock))
		} else if opts.WorkStartRepository != nil {
			workModule = featurework.NewModuleWithStartRepositoryAndDiscordInfo(opts.WorkStartRepository, concreteDiscord, opts.UsageTracker, clockOrSystem(opts.Clock))
		} else if opts.WorkInterfaceRepository != nil {
			workModule = featurework.NewModuleWithRepositoryAndDiscordInfo(opts.WorkInterfaceRepository, concreteDiscord, opts.UsageTracker, clockOrSystem(opts.Clock))
		}
		if err := workModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.WarningsFeatureEnabled && opts.WarningHistoryRepository != nil {
		warningsModule := featuremoderation.NewModule(opts.WarningHistoryRepository, opts.WarningMemberReader, concreteDiscord, opts.UsageTracker)
		if err := warningsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.WarningSettingsFeatureEnabled && opts.WarningSettingsRepository != nil {
		warningSettingsModule := featuremoderation.NewSettingsModule(opts.WarningSettingsRepository, opts.UsageTracker)
		if err := warningSettingsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.WarningRemovalFeatureEnabled && opts.WarningRemovalRepository != nil {
		warningRemovalModule := featuremoderation.NewRemovalModule(opts.WarningRemovalRepository, opts.WarningRemovalDirectMessage, concreteDiscord, opts.UsageTracker)
		if err := warningRemovalModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.WarningIssueFeatureEnabled && opts.WarningIssueRepository != nil {
		warningIssueModule := featuremoderation.NewIssueModule(opts.WarningIssueRepository, opts.WarningSettingsRepository, opts.WarningIssueDirectMessage, concreteDiscord, opts.WarningIssueHierarchy, opts.WarningIssueMemberPort, opts.WarningIssueMessagePort, clockOrSystem(opts.Clock), opts.UsageTracker)
		if err := warningIssueModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.MessageCleanupFeatureEnabled && opts.MessageCleaner != nil {
		cleanupModule := featuremoderation.NewCleanupModule(opts.MessageCleaner, opts.UsageTracker)
		if err := cleanupModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.DeleteDataFeatureEnabled && opts.DeleteDataRepository != nil {
		deleteDataModule := featuremoderation.NewDeleteDataModule(opts.DeleteDataRepository, opts.UsageTracker)
		if err := deleteDataModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.LoggingConfigRepository != nil {
		loggingModule := featurelogging.NewModule(opts.LoggingConfigRepository, opts.UsageTracker)
		if err := loggingModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.BalanceRepository != nil {
		balanceModule := featurebalance.NewModule(opts.BalanceRepository, opts.UsageTracker)
		if err := balanceModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.RedeemRepository != nil {
		redeemModule := featureredeem.NewModule(opts.RedeemRepository, clockOrSystem(opts.Clock), opts.UsageTracker)
		if err := redeemModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.AutoChatConfigRepository != nil {
		autoChatModule := featureautochat.NewModule(opts.AutoChatConfigRepository, opts.UsageTracker)
		if err := autoChatModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.AutoNotificationRepository != nil {
		notificationModule := featurenotifications.NewModuleWithMessagePort(opts.AutoNotificationRepository, concreteDiscord, opts.AutoNotificationMessagePort, opts.UsageTracker)
		if err := notificationModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.AntiScamConfigRepository != nil {
		antiScamModule := featuresafety.NewModule(opts.AntiScamConfigRepository, opts.UsageTracker)
		if err := antiScamModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.ScamURLCatalogRepository != nil && opts.ScamReportSender != nil {
		reportModule := featuresafety.NewReportModule(opts.ScamURLCatalogRepository, opts.ScamReportSender, opts.UsageTracker)
		if err := reportModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.GachaPrizePoolRepository != nil || opts.GachaDrawRepository != nil || opts.GachaPrizeCreateRepository != nil || opts.GachaPrizeEditRepository != nil || opts.GachaPrizeDeleteRepository != nil {
		gachaModule := featuregacha.NewModuleWithRepositories(
			opts.GachaPrizePoolRepository,
			opts.GachaDrawRepository,
			opts.GachaPrizeCreateRepository,
			opts.GachaPrizeEditRepository,
			opts.GachaPrizeDeleteRepository,
			concreteDiscord,
			opts.GachaDrawMessagePort,
			opts.GachaDrawDirectMessagePort,
			opts.UsageTracker,
		)
		gachaModule = gachaModule.WithDrawWait(opts.GachaDrawWait)
		if err := gachaModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.LotteryDisabledCommandEnabled || (opts.LotteryComponentsEnabled && opts.LotteryRepository != nil) {
		var lotteryModule featurelottery.Module
		switch {
		case opts.LotteryDisabledCommandEnabled && opts.LotteryComponentsEnabled && opts.LotteryRepository != nil:
			lotteryModule = featurelottery.NewModuleWithComponents(opts.LotteryRepository, lotteryDiscord, opts.LotteryMemberReader, opts.LotteryMessagePort, clockOrSystem(opts.Clock), opts.UsageTracker)
		case opts.LotteryComponentsEnabled && opts.LotteryRepository != nil:
			lotteryModule = featurelottery.NewComponentModule(opts.LotteryRepository, lotteryDiscord, opts.LotteryMemberReader, opts.LotteryMessagePort, clockOrSystem(opts.Clock))
		default:
			lotteryModule = featurelottery.NewModule(opts.UsageTracker)
		}
		if err := lotteryModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.StatsQueryEnabled {
		statsModule := featurestats.NewModule(opts.UsageTracker)
		if err := statsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.StatsCreateRepository != nil && opts.StatsCreateChannelPort != nil && opts.StatsCreateGuildStats != nil {
		statsModule := featurestats.NewCreateModule(opts.StatsCreateRepository, opts.StatsCreateChannelPort, opts.StatsCreateGuildStats, opts.UsageTracker, opts.BotUserID)
		if err := statsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.StatsRoleStatsRepository != nil && opts.StatsRoleConfigRepository != nil && opts.StatsRoleChannelPort != nil && opts.StatsRoleStatsReader != nil {
		statsModule := featurestats.NewRoleModule(opts.StatsRoleStatsRepository, opts.StatsRoleConfigRepository, opts.StatsRoleChannelPort, opts.StatsRoleStatsReader, opts.UsageTracker, opts.BotUserID)
		if err := statsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.StatsDeleteRepository != nil {
		statsModule := featurestats.NewDeleteModule(opts.StatsDeleteRepository, opts.UsageTracker)
		if err := statsModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.BirthdayConfigRepository != nil {
		birthdayModule := featurebirthday.NewModuleWithClockAndCachedUsers(opts.BirthdayConfigRepository, opts.BirthdayCachedUsers, opts.UsageTracker, clockOrSystem(opts.Clock))
		if err := birthdayModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.AnnouncementConfigRepository != nil || (opts.AnnouncementSendRepository != nil && opts.AnnouncementMessagePort != nil) {
		var announcementModule featureannouncements.Module
		if opts.AnnouncementConfigRepository != nil && opts.AnnouncementMessagePort != nil {
			announcementModule = featureannouncements.NewModuleWithSend(opts.AnnouncementConfigRepository, opts.AnnouncementMessagePort, opts.UsageTracker)
		} else if opts.AnnouncementConfigRepository != nil {
			announcementModule = featureannouncements.NewModule(opts.AnnouncementConfigRepository, opts.UsageTracker)
		} else {
			announcementModule = featureannouncements.NewSendModule(opts.AnnouncementSendRepository, opts.AnnouncementMessagePort, opts.UsageTracker)
		}
		if err := announcementModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.TextXPConfigRepository != nil {
		xpModule := featurexp.NewModule(opts.TextXPConfigRepository, opts.TextXPMessagePort, opts.UsageTracker)
		if err := xpModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.VoiceXPConfigRepository != nil {
		voiceXPModule := featurexp.NewVoiceModule(opts.VoiceXPConfigRepository, opts.VoiceXPMessagePort, opts.UsageTracker)
		if err := voiceXPModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.TextXPRewardRoleRepository != nil && opts.VoiceXPRewardRoleRepository != nil {
		rewardRoleModule := featurexp.NewRewardRoleModule(opts.TextXPRewardRoleRepository, opts.VoiceXPRewardRoleRepository, opts.XPRewardRoleInspector, opts.UsageTracker)
		if err := rewardRoleModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.XPProfileDisabledEnabled {
		xpProfileModule := featurexp.NewDisabledProfileModule(opts.UsageTracker)
		if err := xpProfileModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.XPAdminRepository != nil {
		xpAdminModule := featurexp.NewAdminModule(opts.XPAdminRepository, opts.UsageTracker)
		if err := xpAdminModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if xpResetRuntimeEnabled(opts, xpResetGuilds) {
		xpResetModule := featurexp.NewResetModule(opts.XPResetRepository, xpResetGuilds, opts.XPResetMessagePort, opts.UsageTracker, nil)
		if err := xpResetModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.XPRankRepository != nil {
		xpRankModule := featurexp.NewRankModule(opts.XPRankRepository, concreteDiscord, opts.UsageTracker)
		if err := xpRankModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.VoiceRoomConfigRepository != nil {
		voiceModule := featurevoice.NewModule(opts.VoiceRoomConfigRepository, opts.UsageTracker)
		if err := voiceModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.VoiceRoomLockRepository != nil {
		voiceLockModule := featurevoice.NewLockModule(opts.VoiceRoomLockRepository, opts.UsageTracker)
		if err := voiceLockModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.JoinRoleConfigRepository != nil {
		joinRoleModule := featureonboarding.NewModule(opts.JoinRoleConfigRepository, opts.JoinRoleInspector, opts.UsageTracker)
		if err := joinRoleModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.LeaveMessageConfigRepository != nil {
		messageModule := featureonboarding.NewMessageModule(opts.LeaveMessageConfigRepository, opts.UsageTracker)
		if err := messageModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.VerificationConfigRepository != nil {
		verificationModule := featureonboarding.NewVerificationModule(opts.VerificationConfigRepository, opts.VerificationRoleInspector, opts.UsageTracker)
		if err := verificationModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if verificationFlowRuntimeEnabled(opts) {
		verificationModule := featureonboarding.NewVerificationFlowModule(
			opts.VerificationFlowRepository,
			opts.VerificationRolePort,
			opts.VerificationMemberPort,
			opts.VerificationRoleInspector,
			opts.VerificationGuildInfo,
			opts.UsageTracker,
		)
		if err := verificationModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.AccountAgeConfigRepository != nil {
		accountAgeModule := featureonboarding.NewAccountAgeModule(opts.AccountAgeConfigRepository, opts.UsageTracker)
		if err := accountAgeModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.RoleSelectionRepository != nil {
		roleModule := featureroles.NewModule(
			opts.RoleSelectionRepository,
			opts.RoleSelectionRolePort,
			opts.RoleSelectionRoleInspector,
			opts.RoleSelectionReactionPort,
			opts.RoleSelectionMessagePort,
			opts.RoleSelectionDirectMessage,
		)
		if err := roleModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	dispatcher, err := discordruntime.NewDispatcher(router, opts.Logger)
	if err != nil {
		return nil, err
	}
	for _, shutdown := range runtimeShutdowns {
		dispatcher.RegisterShutdown(shutdown)
	}
	return dispatcher, nil
}

func verificationFlowRuntimeEnabled(opts RuntimeOptions) bool {
	return opts.VerificationFlowRepository != nil && opts.VerificationRolePort != nil
}

func xpResetRuntimeEnabled(opts RuntimeOptions, guilds ports.DiscordInfoProvider) bool {
	return opts.XPResetRepository != nil && opts.XPResetMessagePort != nil && guilds != nil
}

func coinResetRuntimeEnabled(opts RuntimeOptions, guilds ports.DiscordInfoProvider) bool {
	return opts.EconomyCoinResetRepository != nil && opts.EconomyCoinResetMessagePort != nil && guilds != nil
}

func defaultEventRuntimeFactory(cfg config.Config, logger *slog.Logger, session DiscordSession, mongoClient MongoClient) (_ *discordevents.Dispatcher, returnErr error) {
	dispatcher := discordevents.NewDispatcher(logger)
	defer func() {
		if returnErr == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = dispatcher.Shutdown(cleanupCtx)
	}()
	if cfg.FeatureAnnouncementRelayEnabled {
		repo, err := announcementConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "announcement relay feature")
		if err != nil {
			return nil, err
		}
		featureannouncements.NewRelayModule(repo, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureAntiScamMessageDeleteEnabled {
		configRepo, err := antiScamConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		catalogRepo, err := scamURLCatalogRepositoryFromMongoForFeature(mongoClient, "anti-scam message delete feature")
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "anti-scam message delete feature")
		if err != nil {
			return nil, err
		}
		featuresafety.NewMessageDeleteModule(configRepo, catalogRepo, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureLoggingMessageEventsEnabled {
		repo, err := loggingConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "logging message events feature")
		if err != nil {
			return nil, err
		}
		featurelogging.NewMessageEventModule(repo, sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureLoggingChannelEventsEnabled {
		repo, err := loggingConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "logging channel events feature")
		if err != nil {
			return nil, err
		}
		featurelogging.NewChannelEventModule(repo, sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureLoggingVoiceEventsEnabled {
		repo, err := loggingConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "logging voice events feature")
		if err != nil {
			return nil, err
		}
		featurelogging.NewVoiceEventModule(repo, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureAccountAgePolicyEnabled {
		repo, err := accountAgeConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "account-age policy feature")
		if err != nil {
			return nil, err
		}
		featureonboarding.NewAccountAgePolicyModule(
			repo,
			sideEffects,
			sideEffects,
			sideEffects,
			discordInfoProvider(session),
			clockOrSystem(nil),
		).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureWelcomeMessageDeliveryEnabled {
		repo, err := joinMessageConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "welcome-message delivery feature")
		if err != nil {
			return nil, err
		}
		featureonboarding.NewWelcomeMessageDeliveryModule(
			repo,
			sideEffects,
			coreservice.SpecialWelcomeConfig{
				GuildID:          cfg.LegacyWelcomeSpecialGuildID,
				BotID:            cfg.LegacyWelcomeSpecialBotID,
				ChannelID:        cfg.LegacyWelcomeSpecialChannelID,
				ChatChannelID:    cfg.LegacyWelcomeSpecialChatChannelID,
				HelpChannelID:    cfg.LegacyWelcomeSpecialHelpChannelID,
				BugChannelID:     cfg.LegacyWelcomeSpecialBugChannelID,
				SupportChannelID: cfg.LegacyWelcomeSpecialSupportChannelID,
			},
		).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureJoinRoleAssignmentEnabled {
		repo, err := joinRoleConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "join-role assignment feature")
		if err != nil {
			return nil, err
		}
		featureonboarding.NewJoinRoleAssignmentModule(
			repo,
			sideEffects,
			sideEffects,
			discordInfoProvider(session),
			sideEffects,
		).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureLeaveMessageDeliveryEnabled {
		repo, err := leaveMessageConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "leave-message delivery feature")
		if err != nil {
			return nil, err
		}
		featureonboarding.NewLeaveMessageDeliveryModule(repo, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureXPResetEnabled {
		repo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "XP reset feature")
		if err != nil {
			return nil, err
		}
		featurexp.NewResetModule(repo, discordInfoProvider(session), sideEffects, nil, nil).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureEconomyCoinResetEnabled {
		repo, err := economyRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "economy coin-reset feature")
		if err != nil {
			return nil, err
		}
		featureeconomy.NewCoinResetModule(repo, discordInfoProvider(session), sideEffects, nil, nil).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureTextXPAccrualEnabled {
		repo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		configRepo, err := textXPConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "text XP accrual feature")
		if err != nil {
			return nil, err
		}
		rewardRoleRepo, err := textXPRewardRoleRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		economyRepo, err := economyRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		featurexp.NewTextEventModule(repo, configRepo, sideEffects).
			WithAnnouncementFallbacks(sideEffects, sideEffects).
			WithRewardRoles(rewardRoleRepo, sideEffects).
			WithCoinRewards(economyRepo).
			RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureRoleSelectionEnabled {
		repo, err := roleSelectionRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "role-selection feature")
		if err != nil {
			return nil, err
		}
		featureroles.NewModule(repo, sideEffects, discordadapter.NewCachedRoleInspector(sideEffects), sideEffects, sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureVoiceRoomLockEnabled {
		repo, err := voiceRoomLockRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "voice-room lock event feature")
		if err != nil {
			return nil, err
		}
		featurevoice.NewLockEventModule(repo, sideEffects, sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureVoiceRoomConfigEnabled {
		configRepo, err := voiceRoomConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		stateRepo, err := voiceRoomStateRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		lockRepo, err := voiceRoomLockRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "voice-room dynamic event feature")
		if err != nil {
			return nil, err
		}
		featurevoice.NewRoomEventModule(configRepo, stateRepo, lockRepo, sideEffects, sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureVoiceXPSessionsEnabled {
		repo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		configRepo, err := voiceXPConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "voice XP sessions feature")
		if err != nil {
			return nil, err
		}
		rewardRoleRepo, err := voiceXPRewardRoleRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		economyRepo, err := economyRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		module := featurexp.NewVoiceEventModule(repo).
			WithAccrual(repo, configRepo, sideEffects).
			WithAnnouncementFallbacks(sideEffects, sideEffects, discordInfoProvider(session)).
			WithRewardRoles(rewardRoleRepo, sideEffects).
			WithCoinRewards(economyRepo).
			WithRuntimeWorker(featurexp.LegacyVoiceXPInterval, logger)
		started, err := module.StartJoinedSessions(context.Background())
		if err != nil {
			_ = module.StopRuntimeWorker(context.Background())
			return nil, err
		}
		if logger != nil && started > 0 {
			logger.Info("reconciled voice xp sessions", "workers", started)
		}
		module.RegisterEventRoutes(dispatcher)
		dispatcher.RegisterShutdown(module.StopRuntimeWorker)
	}
	if cfg.FeatureStatsRenameWorkerEnabled {
		repo, err := statsConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "stats rename worker feature")
		if err != nil {
			return nil, err
		}
		worker := corestats.NewRenameWorker(corestats.RenameService{
			Repository: repo,
			Channels:   sideEffects,
			GuildStats: sideEffects,
			RoleStats:  sideEffects,
		}, corestats.LegacyStatsRenameInterval, logger)
		worker.Start(context.Background())
		dispatcher.RegisterShutdown(worker.Stop)
		if logger != nil {
			logger.Info("stats rename worker started", "interval", corestats.LegacyStatsRenameInterval.String())
		}
	}
	// Keep delayed chat replies behind all other MessageCreate handlers.
	if cfg.FeatureAutoChatPaidHandoffEnabled {
		configRepo, balanceRepo, handoffRepo, err := autoChatPaidRepositoriesFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "paid autochat feature")
		if err != nil {
			return nil, err
		}
		module, err := featureautochat.NewPaidRuntimeModule(configRepo, balanceRepo, handoffRepo, sideEffects, sideEffects, clockOrSystem(nil))
		if err != nil {
			return nil, err
		}
		module.RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureAutoChatFallbackEnabled {
		configRepo, balanceRepo, err := autoChatFallbackRepositoriesFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "autochat fallback feature")
		if err != nil {
			return nil, err
		}
		module, err := featureautochat.NewRuntimeModule(configRepo, balanceRepo, sideEffects, sideEffects)
		if err != nil {
			return nil, err
		}
		module.RegisterEventRoutes(dispatcher)
	}
	if cfg.FeatureAutoNotificationDelivery {
		const feature = "auto-notification delivery feature"
		repo, err := autoNotificationScheduleRepositoryFromMongoForFeature(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		leases, err := schedulerLeaseStoreFromMongo(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, feature)
		if err != nil {
			return nil, err
		}
		worker, err := corenotifications.NewDeliveryWorker(
			corenotifications.DeliveryService{Repository: repo, Messages: sideEffects},
			leases,
			cfg.SchedulerLeaseOwner,
			cfg.SchedulerLeaseTTL,
			cfg.SchedulerLeaseTimeout,
			logger,
		)
		if err != nil {
			return nil, err
		}
		worker.Start(context.Background())
		dispatcher.RegisterShutdown(worker.Stop)
		if logger != nil {
			logger.Info("auto-notification delivery worker started", "lease", corenotifications.AutoNotificationDeliveryLeaseName, "timezone", corenotifications.AutoNotificationDeliveryLocationName)
		}
	}
	if cfg.FeatureDailyResetSchedulerEnabled {
		const feature = "daily reset scheduler feature"
		repo, err := dailyResetRepositoryFromMongo(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		leases, err := schedulerLeaseStoreFromMongo(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		worker, err := coreeconomy.NewDailyResetWorker(
			repo,
			leases,
			cfg.SchedulerLeaseOwner,
			cfg.SchedulerLeaseTTL,
			cfg.SchedulerLeaseTimeout,
			cfg.JobsDailyResetTimeout,
			logger,
		)
		if err != nil {
			return nil, err
		}
		worker.Start(context.Background())
		dispatcher.RegisterShutdown(worker.Stop)
		if logger != nil {
			logger.Info("daily reset scheduler started", "lease", coreeconomy.DailyResetSchedulerLeaseName, "cron", coreeconomy.DailyResetSchedulerCronSpec, "timezone", coreeconomy.DailyResetSchedulerLocationName)
		}
	}
	if cfg.FeatureWorkPayoutSchedulerEnabled {
		const feature = "work payout scheduler feature"
		repo, err := workPayoutRepositoryFromMongo(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		leases, err := schedulerLeaseStoreFromMongo(mongoClient, feature)
		if err != nil {
			return nil, err
		}
		worker, err := coreeconomy.NewWorkPayoutWorker(
			repo,
			leases,
			cfg.JobsWorkPayoutLeaseName,
			cfg.SchedulerLeaseOwner,
			cfg.SchedulerLeaseTTL,
			cfg.SchedulerLeaseTimeout,
			cfg.JobsWorkPayoutTimeout,
			logger,
		)
		if err != nil {
			return nil, err
		}
		worker.Start(context.Background())
		dispatcher.RegisterShutdown(worker.Stop)
		if logger != nil {
			logger.Info("work payout scheduler started", "lease", cfg.JobsWorkPayoutLeaseName, "cron", coreeconomy.WorkPayoutSchedulerCronSpec, "timezone", coreeconomy.WorkPayoutSchedulerLocationName)
		}
	}
	return dispatcher, nil
}

func botInfoProvider(session DiscordSession) ports.BotInfoProvider {
	if concrete, ok := session.(*discordadapter.Session); ok {
		return discordadapter.BotInfoProvider{
			Session:    concrete,
			Name:       "MHCAT",
			StartedAt:  time.Now(),
			ShardCount: 1,
		}
	}
	return nil
}

func discordInfoProvider(session DiscordSession) ports.DiscordInfoProvider {
	if concrete, ok := session.(*discordadapter.Session); ok {
		return discordadapter.DiscordInfoProvider{Session: concrete}
	}
	return nil
}

func discordCachedUserInfoProvider(session DiscordSession) ports.DiscordCachedUserInfoProvider {
	if concrete, ok := session.(*discordadapter.Session); ok {
		return discordadapter.DiscordInfoProvider{Session: concrete}
	}
	return nil
}

func clockOrSystem(clock ports.Clock) ports.Clock {
	if clock != nil {
		return clock
	}
	return ports.SystemClock{}
}
