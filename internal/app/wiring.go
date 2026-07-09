package app

import (
	"log/slog"
	"time"

	discordadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
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
	featureonboarding "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/onboarding"
	featurepoll "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/poll"
	featuresafety "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/safety"
	featurestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/stats"
	featureticket "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/ticket"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
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
	WorkInterfaceRepository       ports.WorkInterfaceRepository
	WorkStartRepository           ports.WorkStartRepository
	WorkAdminRepository           ports.WorkAdminRepository
	WorkFeatureEnabled            bool
	WarningHistoryRepository      ports.WarningHistoryRepository
	WarningMemberReader           ports.DiscordGuildMemberReader
	WarningsFeatureEnabled        bool
	TranslateProvider             ports.Translator
	TranslateFeatureEnabled       bool
	BalanceRepository             ports.BalanceRepository
	AutoChatConfigRepository      ports.AutoChatConfigRepository
	AntiScamConfigRepository      ports.AntiScamConfigRepository
	ScamURLCatalogRepository      ports.ScamURLCatalog
	ScamReportSender              ports.ScamReportSender
	LoggingConfigRepository       ports.LoggingConfigRepository
	GachaPrizePoolRepository      ports.GachaPrizePoolRepository
	LotteryDisabledCommandEnabled bool
	StatsQueryEnabled             bool
	BirthdayConfigRepository      ports.BirthdayConfigRepository
	AnnouncementConfigRepository  ports.AnnouncementConfigRepository
	AnnouncementSendRepository    ports.AnnouncementChannelReader
	AnnouncementMessagePort       ports.DiscordMessagePort
	TextXPConfigRepository        ports.TextXPConfigRepository
	TextXPMessagePort             ports.DiscordMessagePort
	VoiceXPConfigRepository       ports.VoiceXPConfigRepository
	VoiceXPMessagePort            ports.DiscordMessagePort
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
	BotUserID                     string
	Clock                         ports.Clock
}

const translateInteractionTimeout = 10 * time.Second

func BuildRuntime(opts RuntimeOptions) (*discordruntime.Dispatcher, error) {
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
	if opts.WorkFeatureEnabled {
		definitions = append(definitions, featurework.Definitions()...)
	}
	if opts.WarningsFeatureEnabled {
		definitions = append(definitions, featuremoderation.Definitions()...)
	}
	if opts.TranslateFeatureEnabled && opts.TranslateProvider != nil {
		definitions = append(definitions, commands.TranslateDefinition())
	}
	if opts.BalanceRepository != nil {
		definitions = append(definitions, featurebalance.Definitions()...)
	}
	if opts.AutoChatConfigRepository != nil {
		definitions = append(definitions, featureautochat.Definitions()...)
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
		definitions = append(definitions, featuregacha.Definitions()...)
	}
	if opts.LotteryDisabledCommandEnabled {
		definitions = append(definitions, featurelottery.Definitions()...)
	}
	if opts.StatsQueryEnabled {
		definitions = append(definitions, featurestats.Definitions()...)
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
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, definitions)
	interactionTimeout := opts.Config.DiscordInteractionTimeout
	if opts.TranslateFeatureEnabled && interactionTimeout < translateInteractionTimeout {
		interactionTimeout = translateInteractionTimeout
	}
	router := interactions.NewRouter(
		interactions.Recover(),
		interactions.Timeout(interactionTimeout),
		interactions.Permission(interactions.AllowAllPermissions()),
		interactions.Usage(opts.UsageTracker),
		interactions.Logging(opts.Logger),
	)
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})

	concreteDiscord := discordInfoProvider(opts.Session)
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
	if opts.AutoChatConfigRepository != nil {
		autoChatModule := featureautochat.NewModule(opts.AutoChatConfigRepository, opts.UsageTracker)
		if err := autoChatModule.RegisterRoutes(router); err != nil {
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
	if opts.GachaPrizePoolRepository != nil {
		gachaModule := featuregacha.NewModule(opts.GachaPrizePoolRepository, concreteDiscord, opts.UsageTracker)
		if err := gachaModule.RegisterRoutes(router); err != nil {
			return nil, err
		}
	}
	if opts.LotteryDisabledCommandEnabled {
		lotteryModule := featurelottery.NewModule(opts.UsageTracker)
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
	if opts.BirthdayConfigRepository != nil {
		birthdayModule := featurebirthday.NewModuleWithClock(opts.BirthdayConfigRepository, opts.UsageTracker, clockOrSystem(opts.Clock))
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
	return discordruntime.NewDispatcher(router, opts.Logger)
}

func verificationFlowRuntimeEnabled(opts RuntimeOptions) bool {
	return opts.VerificationFlowRepository != nil && opts.VerificationRolePort != nil
}

func defaultEventRuntimeFactory(cfg config.Config, logger *slog.Logger, session DiscordSession, mongoClient MongoClient) (*discordevents.Dispatcher, error) {
	dispatcher := discordevents.NewDispatcher(logger)
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
		featureonboarding.NewJoinRoleAssignmentModule(repo, sideEffects).RegisterEventRoutes(dispatcher)
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

func clockOrSystem(clock ports.Clock) ports.Clock {
	if clock != nil {
		return clock
	}
	return ports.SystemClock{}
}
