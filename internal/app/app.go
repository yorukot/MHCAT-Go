package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	discordadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	externaladapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/external"
	mongoadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	mongorepositories "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/repositories"
	usageadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/usage"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

type MongoClient interface {
	Connect(context.Context) error
	Disconnect(context.Context) error
}

type DiscordSession interface {
	Open() error
	Close() error
	RegisterInteractionHandler(discordruntime.Handler) func()
	Ready() <-chan struct{}
}

type GatewayEventSession interface {
	RegisterGatewayEventHandlers(*discordevents.Dispatcher, discordadapter.GatewayEventOptions) func()
}

type MongoFactory func(config.Config) (MongoClient, error)
type DiscordFactory func(config.Config) (DiscordSession, error)
type RuntimeFactory func(config.Config, *slog.Logger, DiscordSession, MongoClient) (*discordruntime.Dispatcher, error)
type EventRuntimeFactory func(config.Config, *slog.Logger, DiscordSession, MongoClient) (*discordevents.Dispatcher, error)

type App struct {
	cfg            config.Config
	logger         *slog.Logger
	mongoFactory   MongoFactory
	discordFactory DiscordFactory
	runtimeFactory RuntimeFactory
	eventFactory   EventRuntimeFactory

	mu                        sync.Mutex
	mongo                     MongoClient
	discord                   DiscordSession
	removeInteractionHandler  func()
	removeGatewayEventHandler func()
	runtimeDispatcher         *discordruntime.Dispatcher
	eventDispatcher           *discordevents.Dispatcher

	shutdownOnce sync.Once
	shutdownErr  error
}

type Option func(*App)

func WithMongoFactory(factory MongoFactory) Option {
	return func(a *App) {
		a.mongoFactory = factory
	}
}

func WithDiscordFactory(factory DiscordFactory) Option {
	return func(a *App) {
		a.discordFactory = factory
	}
}

func WithRuntimeFactory(factory RuntimeFactory) Option {
	return func(a *App) {
		a.runtimeFactory = factory
	}
}

func WithEventRuntimeFactory(factory EventRuntimeFactory) Option {
	return func(a *App) {
		a.eventFactory = factory
	}
}

func New(cfg config.Config, logger *slog.Logger, opts ...Option) (*App, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	a := &App{
		cfg:            cfg,
		logger:         logger,
		mongoFactory:   defaultMongoFactory,
		discordFactory: defaultDiscordFactory,
		runtimeFactory: defaultRuntimeFactory,
		eventFactory:   defaultEventRuntimeFactory,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.Start(ctx); err != nil {
		return err
	}
	if !a.cfg.DiscordEnableGateway {
		return a.Shutdown(ctx)
	}
	if a.cfg.DiscordGatewaySmokeTest {
		smokeCtx, cancel := context.WithTimeout(ctx, GatewaySmokeTimeout(a.cfg))
		defer cancel()
		if err := a.WaitReady(smokeCtx); err != nil {
			_ = a.Shutdown(context.Background())
			return err
		}
		a.logger.Info("gateway smoke test ready")
		return a.Shutdown(context.Background())
	}
	<-ctx.Done()
	return a.Shutdown(context.Background())
}

func (a *App) Start(ctx context.Context) error {
	for _, warning := range a.cfg.AliasWarnings {
		a.logger.Warn(warning.Message(), slog.Any("values", warning.RedactedFields()))
	}

	mongoClient, err := a.mongoFactory(a.cfg)
	if err != nil {
		return fmt.Errorf("create mongo client: %w", err)
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}

	discordSession, err := a.discordFactory(a.cfg)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return fmt.Errorf("create discord session: %w", err)
	}

	dispatcher, err := a.runtimeFactory(a.cfg, a.logger, discordSession, mongoClient)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		_ = discordSession.Close()
		return fmt.Errorf("create runtime dispatcher: %w", err)
	}
	removeHandler := func() {}
	removeEventHandler := func() {}
	var eventDispatcher *discordevents.Dispatcher
	if a.cfg.DiscordEnableGateway {
		removeHandler = discordSession.RegisterInteractionHandler(dispatcher.Handler())
		if gatewayEvents, ok := discordSession.(GatewayEventSession); ok {
			eventDispatcher, err = a.eventFactory(a.cfg, a.logger, discordSession, mongoClient)
			if err != nil {
				removeHandler()
				_ = dispatcher.Shutdown(context.Background())
				_ = mongoClient.Disconnect(context.Background())
				_ = discordSession.Close()
				return fmt.Errorf("create gateway event dispatcher: %w", err)
			}
			removeEventHandler = gatewayEvents.RegisterGatewayEventHandlers(eventDispatcher, discordadapter.GatewayEventOptions{
				Messages:         a.cfg.DiscordGuildMessagesIntent,
				GuildChannels:    a.cfg.FeatureLoggingChannelEventsEnabled,
				MessageReactions: a.cfg.DiscordMessageReactionsIntent,
				GuildMembers:     a.cfg.DiscordGuildMembersIntent,
				VoiceStates:      a.cfg.DiscordVoiceStateIntent,
			})
		}
		if err := openWithTimeout(ctx, discordSession, a.cfg.DiscordGatewayConnectTimeout); err != nil {
			removeHandler()
			removeEventHandler()
			_ = dispatcher.Shutdown(context.Background())
			if eventDispatcher != nil {
				_ = eventDispatcher.Shutdown(context.Background())
			}
			_ = mongoClient.Disconnect(context.Background())
			return fmt.Errorf("open discord gateway: %w", err)
		}
	}

	a.mu.Lock()
	a.mongo = mongoClient
	a.discord = discordSession
	a.removeInteractionHandler = removeHandler
	a.removeGatewayEventHandler = removeEventHandler
	a.runtimeDispatcher = dispatcher
	a.eventDispatcher = eventDispatcher
	a.mu.Unlock()

	a.logger.Info("mhcat skeleton initialized", "gateway_enabled", a.cfg.DiscordEnableGateway)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.shutdownOnce.Do(func() {
		a.mu.Lock()
		mongoClient := a.mongo
		discordSession := a.discord
		runtimeDispatcher := a.runtimeDispatcher
		eventDispatcher := a.eventDispatcher
		a.mu.Unlock()

		var errs []error
		if discordSession != nil {
			if a.removeInteractionHandler != nil {
				a.removeInteractionHandler()
			}
			if a.removeGatewayEventHandler != nil {
				a.removeGatewayEventHandler()
			}
		}
		if runtimeDispatcher != nil {
			if err := runtimeDispatcher.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if eventDispatcher != nil {
			if err := eventDispatcher.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if discordSession != nil {
			if err := discordSession.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if mongoClient != nil {
			if err := mongoClient.Disconnect(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		a.shutdownErr = errors.Join(errs...)
	})
	return a.shutdownErr
}

func (a *App) WaitReady(ctx context.Context) error {
	a.mu.Lock()
	discordSession := a.discord
	a.mu.Unlock()
	if discordSession == nil {
		return discordruntime.ErrRuntimeNotConfigured
	}
	return discordruntime.WaitReady(ctx, discordSession)
}

func defaultMongoFactory(cfg config.Config) (MongoClient, error) {
	return mongoadapter.NewClient(mongoadapter.Options{
		URI:            cfg.MongoDBURI,
		Database:       cfg.MongoDBDatabase,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
}

func defaultDiscordFactory(cfg config.Config) (DiscordSession, error) {
	intents := discordadapter.BuildIntents(discordadapter.IntentOptions{
		GuildMembers:     cfg.DiscordGuildMembersIntent,
		GuildMessages:    cfg.DiscordGuildMessagesIntent,
		MessageReactions: cfg.DiscordMessageReactionsIntent,
		VoiceStates:      cfg.DiscordVoiceStateIntent,
		MessageContent:   cfg.DiscordMessageContentIntent,
	})
	return discordadapter.NewSession(cfg.DiscordToken, intents)
}

func openWithTimeout(ctx context.Context, session DiscordSession, timeout time.Duration) error {
	openCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- session.Open()
	}()
	select {
	case err := <-done:
		return err
	case <-openCtx.Done():
		return openCtx.Err()
	}
}

func defaultRuntimeFactory(cfg config.Config, logger *slog.Logger, session DiscordSession, mongoClient MongoClient) (*discordruntime.Dispatcher, error) {
	usageTracker := ports.UsageTracker(usageadapter.NoopTracker{})
	if cfg.FeatureUsageTrackingEnabled {
		tracker, err := usageTrackerFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		usageTracker = tracker
	}
	opts := RuntimeOptions{
		Config:                        cfg,
		Logger:                        logger,
		Session:                       session,
		UsageTracker:                  usageTracker,
		WorkFeatureEnabled:            cfg.FeatureWorkEnabled,
		WarningsFeatureEnabled:        cfg.FeatureWarningsEnabled,
		WarningSettingsFeatureEnabled: cfg.FeatureWarningSettingsEnabled,
		WarningRemovalFeatureEnabled:  cfg.FeatureWarningRemovalEnabled,
		WarningIssueFeatureEnabled:    cfg.FeatureWarningIssueEnabled,
		MessageCleanupFeatureEnabled:  cfg.FeatureMessageCleanupEnabled,
		DeleteDataFeatureEnabled:      cfg.FeatureDeleteDataEnabled,
		TranslateFeatureEnabled:       cfg.FeatureTranslateEnabled,
		LotteryDisabledCommandEnabled: cfg.FeatureLotteryDisabledCommandEnabled,
		LotteryComponentsEnabled:      cfg.FeatureLotteryComponentsEnabled,
		StatsQueryEnabled:             cfg.FeatureStatsQueryEnabled,
	}
	if cfg.FeatureTicketsEnabled {
		ticketRepo, err := ticketConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.TicketConfigRepository = ticketRepo
		opts.TicketChannelPort = sideEffects
		opts.TicketMessagePort = sideEffects
	}
	if cfg.FeaturePollsEnabled {
		pollRepo, err := pollRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.PollRepository = pollRepo
		opts.PollMessagePort = sideEffects
		opts.PollMemberCounter = sideEffects
	}
	if cfg.FeatureEconomyQueryEnabled || cfg.FeatureEconomySignInEnabled || cfg.FeatureEconomyCoinAdminEnabled || cfg.FeatureEconomyCoinRankEnabled || cfg.FeatureEconomyCoinResetEnabled || cfg.FeatureEconomyRPSEnabled || cfg.FeatureEconomyGameEnabled || cfg.FeatureEconomyShopEnabled {
		economyRepo, err := economyRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		if cfg.FeatureEconomyQueryEnabled || cfg.FeatureEconomySignInEnabled {
			opts.EconomyQueryRepository = economyRepo
		}
		if cfg.FeatureEconomySignInEnabled {
			opts.EconomySignInRepository = economyRepo
		}
		if cfg.FeatureEconomyCoinAdminEnabled {
			opts.EconomyCoinAdminRepository = economyRepo
		}
		if cfg.FeatureEconomyCoinRankEnabled {
			opts.EconomyCoinRankRepository = economyRepo
		}
		if cfg.FeatureEconomyCoinResetEnabled {
			sideEffects, err := messageSideEffectsFromSession(session, "economy coin-reset feature")
			if err != nil {
				return nil, err
			}
			opts.EconomyCoinResetRepository = economyRepo
			opts.EconomyCoinResetMessagePort = sideEffects
			opts.EconomyCoinResetGuildInfo = discordInfoProvider(session)
		}
		if cfg.FeatureEconomyRPSEnabled {
			opts.EconomyRPSRepository = economyRepo
		}
		if cfg.FeatureEconomyGameEnabled {
			concreteMongo, ok := mongoClient.(*mongoadapter.Client)
			if !ok {
				return nil, fmt.Errorf("economy game feature requires default mongo client")
			}
			transactions, err := mongoadapter.NewTransactionRunner(concreteMongo)
			if err != nil {
				return nil, fmt.Errorf("economy game transaction runner: %w", err)
			}
			if err := economyRepo.SetCoinGameTransactionRunner(transactions); err != nil {
				return nil, fmt.Errorf("configure economy game transactions: %w", err)
			}
			sideEffects, err := messageSideEffectsFromSession(session, "economy game feature")
			if err != nil {
				return nil, err
			}
			opts.EconomyGameRepository = economyRepo
			opts.EconomyGameMessagePort = sideEffects
		}
		if cfg.FeatureEconomyShopEnabled {
			sideEffects, err := messageSideEffectsFromSession(session, "economy shop feature")
			if err != nil {
				return nil, err
			}
			opts.EconomyShopRepository = economyRepo
			opts.EconomyShopDirectMessage = sideEffects
			opts.EconomyShopRolePort = sideEffects
			opts.EconomyShopRoleInspector = sideEffects
		}
	}
	if cfg.FeatureEconomyProfileEnabled {
		profileRepo, err := economyProfileRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.EconomyProfileRepository = profileRepo
	}
	if cfg.FeatureEconomySettingsEnabled {
		economyRepo, err := economyRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.EconomySettingsRepository = economyRepo
	}
	if cfg.FeatureWorkEnabled {
		workRepo, err := workInterfaceRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.WorkInterfaceRepository = workRepo
		opts.WorkStartRepository = workRepo
		opts.WorkAdminRepository = workRepo
	}
	if cfg.FeatureWarningsEnabled {
		warningRepo, err := warningHistoryRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		memberReader, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.WarningHistoryRepository = warningRepo
		opts.WarningMemberReader = memberReader
	}
	if cfg.FeatureWarningSettingsEnabled {
		warningSettingsRepo, err := warningSettingsRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.WarningSettingsRepository = warningSettingsRepo
	}
	if cfg.FeatureWarningRemovalEnabled {
		warningRepo, err := warningHistoryRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.WarningRemovalRepository = warningRepo
		opts.WarningRemovalDirectMessage = sideEffects
	}
	if cfg.FeatureWarningIssueEnabled {
		warningRepo, err := warningHistoryRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		if opts.WarningSettingsRepository == nil {
			warningSettingsRepo, err := warningSettingsRepositoryFromMongo(mongoClient)
			if err != nil {
				return nil, err
			}
			opts.WarningSettingsRepository = warningSettingsRepo
		}
		sideEffects, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.WarningIssueRepository = warningRepo
		opts.WarningIssueDirectMessage = sideEffects
		opts.WarningIssueMemberPort = sideEffects
		opts.WarningIssueHierarchy = sideEffects
		opts.WarningIssueMessagePort = sideEffects
	}
	if cfg.FeatureMessageCleanupEnabled {
		sideEffects, err := messageSideEffectsFromSession(session, "message cleanup feature")
		if err != nil {
			return nil, err
		}
		opts.MessageCleaner = sideEffects
	}
	if cfg.FeatureDeleteDataEnabled {
		deleteDataRepo, err := deleteDataRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.DeleteDataRepository = deleteDataRepo
	}
	if cfg.FeatureTranslateEnabled {
		translator := externaladapter.NewGoogleTranslateClient()
		opts.TranslateProvider = translator
	}
	if cfg.FeatureBalanceQueryEnabled {
		balanceRepo, err := balanceRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.BalanceRepository = balanceRepo
	}
	if cfg.FeatureRedeemEnabled {
		redeemRepo, err := redeemRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.RedeemRepository = redeemRepo
	}
	if cfg.FeatureAutoChatConfigEnabled {
		autoChatRepo, err := autoChatConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.AutoChatConfigRepository = autoChatRepo
	}
	if cfg.FeatureAutoNotificationConfigEnabled {
		autoNotificationRepo, err := autoNotificationScheduleRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := ticketSideEffectsFromSession(session)
		if err != nil {
			return nil, err
		}
		opts.AutoNotificationRepository = autoNotificationRepo
		opts.AutoNotificationMessagePort = sideEffects
	}
	if cfg.FeatureAntiScamConfigEnabled {
		antiScamRepo, err := antiScamConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.AntiScamConfigRepository = antiScamRepo
	}
	if cfg.FeatureAntiScamReportEnabled {
		scamURLRepo, err := scamURLCatalogRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.ScamURLCatalogRepository = scamURLRepo
		opts.ScamReportSender = externaladapter.NewDiscordWebhookReporter(cfg.ReportWebhookURL)
	}
	if cfg.FeatureLoggingConfigEnabled {
		loggingRepo, err := loggingConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.LoggingConfigRepository = loggingRepo
	}
	if cfg.FeatureGachaPrizeListEnabled || cfg.FeatureGachaDrawEnabled || cfg.FeatureGachaPrizeCreateEnabled || cfg.FeatureGachaPrizeEditEnabled || cfg.FeatureGachaPrizeDeleteEnabled {
		gachaRepo, err := gachaRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		if cfg.FeatureGachaPrizeListEnabled {
			opts.GachaPrizePoolRepository = gachaRepo
		}
		if cfg.FeatureGachaDrawEnabled {
			sideEffects, err := messageSideEffectsFromSession(session, "gacha draw feature")
			if err != nil {
				return nil, err
			}
			opts.GachaDrawRepository = gachaRepo
			opts.GachaDrawMessagePort = sideEffects
			opts.GachaDrawDirectMessagePort = sideEffects
		}
		if cfg.FeatureGachaPrizeCreateEnabled {
			opts.GachaPrizeCreateRepository = gachaRepo
		}
		if cfg.FeatureGachaPrizeEditEnabled {
			opts.GachaPrizeEditRepository = gachaRepo
		}
		if cfg.FeatureGachaPrizeDeleteEnabled {
			opts.GachaPrizeDeleteRepository = gachaRepo
		}
	}
	if cfg.FeatureLotteryComponentsEnabled {
		lotteryRepo, err := lotteryRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "lottery component feature")
		if err != nil {
			return nil, err
		}
		opts.LotteryRepository = lotteryRepo
		opts.LotteryDiscordInfo = discordInfoProvider(session)
		opts.LotteryMemberReader = sideEffects
		opts.LotteryMessagePort = sideEffects
	}
	if cfg.FeatureStatsDeleteEnabled {
		statsRepo, err := statsConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.StatsDeleteRepository = statsRepo
	}
	if cfg.FeatureStatsCreateEnabled {
		statsRepo, err := statsConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "stats create feature")
		if err != nil {
			return nil, err
		}
		opts.StatsCreateRepository = statsRepo
		opts.StatsCreateChannelPort = sideEffects
		opts.StatsCreateGuildStats = sideEffects
	}
	if cfg.FeatureStatsRoleCountEnabled {
		statsRepo, err := statsConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "stats role-count feature")
		if err != nil {
			return nil, err
		}
		opts.StatsRoleStatsRepository = statsRepo
		opts.StatsRoleConfigRepository = statsRepo
		opts.StatsRoleChannelPort = sideEffects
		opts.StatsRoleStatsReader = sideEffects
	}
	if cfg.FeatureBirthdayConfigEnabled {
		birthdayRepo, err := birthdayConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.BirthdayConfigRepository = birthdayRepo
		opts.BirthdayCachedUsers = discordCachedUserInfoProvider(session)
	}
	if cfg.FeatureAnnouncementConfigEnabled || cfg.FeatureAnnouncementSendEnabled {
		announcementRepo, err := announcementConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		if cfg.FeatureAnnouncementConfigEnabled {
			opts.AnnouncementConfigRepository = announcementRepo
		}
		if cfg.FeatureAnnouncementSendEnabled {
			sideEffects, err := messageSideEffectsFromSession(session, "announcement send feature")
			if err != nil {
				return nil, err
			}
			opts.AnnouncementSendRepository = announcementRepo
			opts.AnnouncementMessagePort = sideEffects
		}
	}
	if cfg.FeatureTextXPConfigEnabled {
		textXPRepo, err := textXPConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "text XP config feature")
		if err != nil {
			return nil, err
		}
		opts.TextXPConfigRepository = textXPRepo
		opts.TextXPMessagePort = sideEffects
	}
	if cfg.FeatureVoiceXPConfigEnabled {
		voiceXPRepo, err := voiceXPConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "voice XP config feature")
		if err != nil {
			return nil, err
		}
		opts.VoiceXPConfigRepository = voiceXPRepo
		opts.VoiceXPMessagePort = sideEffects
	}
	if cfg.FeatureXPRoleConfigEnabled {
		textRoleRepo, err := textXPRewardRoleRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		voiceRoleRepo, err := voiceXPRewardRoleRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "XP role config feature")
		if err != nil {
			return nil, err
		}
		opts.TextXPRewardRoleRepository = textRoleRepo
		opts.VoiceXPRewardRoleRepository = voiceRoleRepo
		opts.XPRewardRoleInspector = sideEffects
	}
	if cfg.FeatureXPProfileDisabledEnabled {
		opts.XPProfileDisabledEnabled = true
	}
	if cfg.FeatureXPAdminEnabled {
		xpAdminRepo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.XPAdminRepository = xpAdminRepo
	}
	if cfg.FeatureXPResetEnabled {
		xpResetRepo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "XP reset feature")
		if err != nil {
			return nil, err
		}
		opts.XPResetRepository = xpResetRepo
		opts.XPResetMessagePort = sideEffects
		opts.XPResetGuildInfo = discordInfoProvider(session)
	}
	if cfg.FeatureXPRankEnabled {
		xpRankRepo, err := xpAdminRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.XPRankRepository = xpRankRepo
	}
	if cfg.FeatureVoiceRoomConfigEnabled {
		voiceRoomRepo, err := voiceRoomConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.VoiceRoomConfigRepository = voiceRoomRepo
	}
	if cfg.FeatureVoiceRoomLockEnabled {
		voiceRoomLockRepo, err := voiceRoomLockRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.VoiceRoomLockRepository = voiceRoomLockRepo
	}
	if cfg.FeatureJoinRoleConfigEnabled {
		joinRoleRepo, err := joinRoleConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "join-role config feature")
		if err != nil {
			return nil, err
		}
		opts.JoinRoleConfigRepository = joinRoleRepo
		opts.JoinRoleInspector = sideEffects
	}
	if cfg.FeatureWelcomeMessageConfigEnabled {
		leaveMessageRepo, err := leaveMessageConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.LeaveMessageConfigRepository = leaveMessageRepo
	}
	if cfg.FeatureVerificationConfigEnabled {
		verificationRepo, err := verificationConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "verification config feature")
		if err != nil {
			return nil, err
		}
		opts.VerificationConfigRepository = verificationRepo
		opts.VerificationRoleInspector = sideEffects
	}
	if cfg.FeatureVerificationFlowEnabled {
		verificationRepo, err := verificationConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "verification flow feature")
		if err != nil {
			return nil, err
		}
		opts.VerificationFlowRepository = verificationRepo
		opts.VerificationRolePort = sideEffects
		opts.VerificationMemberPort = sideEffects
		opts.VerificationRoleInspector = sideEffects
		opts.VerificationGuildInfo = discordInfoProvider(session)
	}
	if cfg.FeatureAccountAgeConfigEnabled {
		accountAgeRepo, err := accountAgeConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.AccountAgeConfigRepository = accountAgeRepo
	}
	if roleSelectionOwnershipEnabled(cfg) {
		roleRepo, err := roleSelectionRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		sideEffects, err := messageSideEffectsFromSession(session, "role-selection feature")
		if err != nil {
			return nil, err
		}
		opts.RoleSelectionRepository = roleRepo
		opts.RoleSelectionRolePort = sideEffects
		opts.RoleSelectionRoleInspector = discordadapter.NewCachedRoleInspector(sideEffects)
		opts.RoleSelectionReactionPort = sideEffects
		opts.RoleSelectionMessagePort = sideEffects
		opts.RoleSelectionDirectMessage = sideEffects
	}
	return BuildRuntime(opts)
}

func ticketConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.TicketConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("ticket feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("ticket feature database: %w", err)
	}
	repo, err := mongorepositories.NewTicketConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("ticket feature repository: %w", err)
	}
	return repo, nil
}

func pollRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.PollRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("poll feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("poll feature database: %w", err)
	}
	repo, err := mongorepositories.NewPollRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("poll feature repository: %w", err)
	}
	return repo, nil
}

func economyRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.EconomyRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("economy query feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("economy query feature database: %w", err)
	}
	repo, err := mongorepositories.NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("economy query feature repository: %w", err)
	}
	return repo, nil
}

func roleSelectionRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.RoleSelectionRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("role-selection feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("role-selection feature database: %w", err)
	}
	repo, err := mongorepositories.NewRoleSelectionRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("role-selection feature repository: %w", err)
	}
	return repo, nil
}

func economyProfileRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.EconomyProfileRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("economy profile feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("economy profile feature database: %w", err)
	}
	repo, err := mongorepositories.NewEconomyProfileRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("economy profile feature repository: %w", err)
	}
	return repo, nil
}

func statsConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.StatsConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("stats delete feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("stats delete feature database: %w", err)
	}
	repo, err := mongorepositories.NewStatsConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("stats delete feature repository: %w", err)
	}
	return repo, nil
}

func workInterfaceRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.WorkInterfaceRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("work feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("work feature database: %w", err)
	}
	repo, err := mongorepositories.NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("work feature repository: %w", err)
	}
	return repo, nil
}

func warningHistoryRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.WarningHistoryRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("warnings feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("warnings feature database: %w", err)
	}
	repo, err := mongorepositories.NewWarningHistoryRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("warnings feature repository: %w", err)
	}
	return repo, nil
}

func warningSettingsRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.WarningSettingsRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("warning settings feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("warning settings feature database: %w", err)
	}
	repo, err := mongorepositories.NewWarningSettingsRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("warning settings feature repository: %w", err)
	}
	return repo, nil
}

func deleteDataRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.DeleteDataRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("delete data feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("delete data feature database: %w", err)
	}
	repo, err := mongorepositories.NewDeleteDataRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("delete data feature repository: %w", err)
	}
	return repo, nil
}

func loggingConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.LoggingConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("logging config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("logging config feature database: %w", err)
	}
	repo, err := mongorepositories.NewLoggingConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("logging config feature repository: %w", err)
	}
	return repo, nil
}

func autoChatConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AutoChatConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("autochat config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("autochat config feature database: %w", err)
	}
	repo, err := mongorepositories.NewAutoChatConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("autochat config feature repository: %w", err)
	}
	return repo, nil
}

func autoChatFallbackRepositoriesFromMongo(mongoClient MongoClient) (*mongorepositories.AutoChatConfigRepository, *mongorepositories.BalanceRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, nil, fmt.Errorf("autochat fallback feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, nil, fmt.Errorf("autochat fallback feature database: %w", err)
	}
	configRepo, err := mongorepositories.NewAutoChatConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, nil, fmt.Errorf("autochat fallback config repository: %w", err)
	}
	balanceRepo, err := mongorepositories.NewBalanceRepositoryFromDatabase(database)
	if err != nil {
		return nil, nil, fmt.Errorf("autochat fallback balance repository: %w", err)
	}
	return configRepo, balanceRepo, nil
}

func autoChatPaidRepositoriesFromMongo(mongoClient MongoClient) (*mongorepositories.AutoChatConfigRepository, *mongorepositories.BalanceRepository, *mongorepositories.AutoChatPaidRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, nil, nil, fmt.Errorf("paid autochat feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("paid autochat feature database: %w", err)
	}
	configRepo, err := mongorepositories.NewAutoChatConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("paid autochat config repository: %w", err)
	}
	balanceRepo, err := mongorepositories.NewBalanceRepositoryFromDatabase(database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("paid autochat balance repository: %w", err)
	}
	transactions, err := mongoadapter.NewTransactionRunner(concrete)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("paid autochat transaction runner: %w", err)
	}
	handoffRepo, err := mongorepositories.NewAutoChatPaidRepositoryFromDatabase(database, transactions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("paid autochat handoff repository: %w", err)
	}
	return configRepo, balanceRepo, handoffRepo, nil
}

func autoNotificationScheduleRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AutoNotificationScheduleRepository, error) {
	return autoNotificationScheduleRepositoryFromMongoForFeature(mongoClient, "auto-notification config feature")
}

func autoNotificationScheduleRepositoryFromMongoForFeature(mongoClient MongoClient, feature string) (*mongorepositories.AutoNotificationScheduleRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", feature)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", feature, err)
	}
	repo, err := mongorepositories.NewAutoNotificationScheduleRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s repository: %w", feature, err)
	}
	return repo, nil
}

func schedulerLeaseStoreFromMongo(mongoClient MongoClient, feature string) (*mongoadapter.SchedulerLeaseStore, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", feature)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", feature, err)
	}
	store, err := mongoadapter.NewSchedulerLeaseStoreFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s lease store: %w", feature, err)
	}
	return store, nil
}

func dailyResetRepositoryFromMongo(mongoClient MongoClient, feature string) (*mongorepositories.DailyResetRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", feature)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", feature, err)
	}
	repository, err := mongorepositories.NewDailyResetRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s repository: %w", feature, err)
	}
	return repository, nil
}

func workPayoutRepositoryFromMongo(mongoClient MongoClient, feature string) (*mongorepositories.WorkPayoutRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", feature)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", feature, err)
	}
	repository, err := mongorepositories.NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s repository: %w", feature, err)
	}
	return repository, nil
}

func balanceRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.BalanceRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("balance query feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("balance query feature database: %w", err)
	}
	repo, err := mongorepositories.NewBalanceRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("balance query feature repository: %w", err)
	}
	return repo, nil
}

func usageTrackerFromMongo(mongoClient MongoClient) (*mongorepositories.UsageTracker, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("usage tracking feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("usage tracking feature database: %w", err)
	}
	tracker, err := mongorepositories.NewUsageTrackerFromDatabase(database, mongorepositories.DefaultUsageTrackTimeout)
	if err != nil {
		return nil, fmt.Errorf("usage tracking feature adapter: %w", err)
	}
	return tracker, nil
}

func redeemRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.RedeemRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("redeem feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("redeem feature database: %w", err)
	}
	repo, err := mongorepositories.NewRedeemRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("redeem feature repository: %w", err)
	}
	return repo, nil
}

func antiScamConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AntiScamConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("anti-scam config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("anti-scam config feature database: %w", err)
	}
	repo, err := mongorepositories.NewAntiScamConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("anti-scam config feature repository: %w", err)
	}
	return repo, nil
}

func scamURLCatalogRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.ScamURLCatalogRepository, error) {
	return scamURLCatalogRepositoryFromMongoForFeature(mongoClient, "anti-scam report feature")
}

func scamURLCatalogRepositoryFromMongoForFeature(mongoClient MongoClient, featureLabel string) (*mongorepositories.ScamURLCatalogRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", featureLabel)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", featureLabel, err)
	}
	repo, err := mongorepositories.NewScamURLCatalogRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s repository: %w", featureLabel, err)
	}
	return repo, nil
}

func gachaRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.GachaRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("gacha feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("gacha feature database: %w", err)
	}
	repo, err := mongorepositories.NewGachaRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("gacha feature repository: %w", err)
	}
	return repo, nil
}

func lotteryRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.LotteryRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("lottery component feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("lottery component feature database: %w", err)
	}
	repo, err := mongorepositories.NewLotteryRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("lottery component feature repository: %w", err)
	}
	return repo, nil
}

func birthdayConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.BirthdayConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("birthday config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("birthday config feature database: %w", err)
	}
	repo, err := mongorepositories.NewBirthdayConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("birthday config feature repository: %w", err)
	}
	return repo, nil
}

func announcementConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AnnouncementConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("announcement config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("announcement config feature database: %w", err)
	}
	repo, err := mongorepositories.NewAnnouncementConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("announcement config feature repository: %w", err)
	}
	return repo, nil
}

func textXPConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.TextXPConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("text XP config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("text XP config feature database: %w", err)
	}
	repo, err := mongorepositories.NewTextXPConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("text XP config feature repository: %w", err)
	}
	return repo, nil
}

func voiceXPConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VoiceXPConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("voice XP config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("voice XP config feature database: %w", err)
	}
	repo, err := mongorepositories.NewVoiceXPConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("voice XP config feature repository: %w", err)
	}
	return repo, nil
}

func textXPRewardRoleRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.TextXPRewardRoleRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("XP role config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("XP role config feature database: %w", err)
	}
	repo, err := mongorepositories.NewTextXPRewardRoleRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("text XP reward role feature repository: %w", err)
	}
	return repo, nil
}

func voiceXPRewardRoleRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VoiceXPRewardRoleRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("XP role config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("XP role config feature database: %w", err)
	}
	repo, err := mongorepositories.NewVoiceXPRewardRoleRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("voice XP reward role feature repository: %w", err)
	}
	return repo, nil
}

func xpAdminRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.XPAdminRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("XP admin feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("XP admin feature database: %w", err)
	}
	repo, err := mongorepositories.NewXPAdminRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("XP admin feature repository: %w", err)
	}
	return repo, nil
}

func voiceRoomConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VoiceRoomConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("voice-room config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("voice-room config feature database: %w", err)
	}
	repo, err := mongorepositories.NewVoiceRoomConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("voice-room config feature repository: %w", err)
	}
	return repo, nil
}

func voiceRoomLockRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VoiceRoomLockRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("voice-room lock feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("voice-room lock feature database: %w", err)
	}
	repo, err := mongorepositories.NewVoiceRoomLockRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("voice-room lock feature repository: %w", err)
	}
	return repo, nil
}

func voiceRoomStateRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VoiceRoomStateRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("voice-room state feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("voice-room state feature database: %w", err)
	}
	repo, err := mongorepositories.NewVoiceRoomStateRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("voice-room state feature repository: %w", err)
	}
	return repo, nil
}

func joinRoleConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.JoinRoleConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("join-role config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("join-role config feature database: %w", err)
	}
	repo, err := mongorepositories.NewJoinRoleConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("join-role config feature repository: %w", err)
	}
	return repo, nil
}

func leaveMessageConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.LeaveMessageConfigRepository, error) {
	return leaveMessageConfigRepositoryFromMongoForFeature(mongoClient, "welcome-message config feature")
}

func leaveMessageConfigRepositoryFromMongoForFeature(mongoClient MongoClient, feature string) (*mongorepositories.LeaveMessageConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("%s requires default mongo client", feature)
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("%s database: %w", feature, err)
	}
	repo, err := mongorepositories.NewLeaveMessageConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("%s repository: %w", feature, err)
	}
	return repo, nil
}

func joinMessageConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.JoinMessageConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("welcome-message delivery feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("welcome-message delivery feature database: %w", err)
	}
	repo, err := mongorepositories.NewJoinMessageConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("welcome-message delivery feature repository: %w", err)
	}
	return repo, nil
}

func verificationConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.VerificationConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("verification config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("verification config feature database: %w", err)
	}
	repo, err := mongorepositories.NewVerificationConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("verification config feature repository: %w", err)
	}
	return repo, nil
}

func accountAgeConfigRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AccountAgeConfigRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("account-age feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("account-age feature database: %w", err)
	}
	repo, err := mongorepositories.NewAccountAgeConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("account-age feature repository: %w", err)
	}
	return repo, nil
}

func ticketSideEffectsFromSession(session DiscordSession) (discordadapter.SideEffectClient, error) {
	return messageSideEffectsFromSession(session, "ticket feature")
}

func messageSideEffectsFromSession(session DiscordSession, feature string) (discordadapter.SideEffectClient, error) {
	concrete, ok := session.(*discordadapter.Session)
	if !ok {
		return discordadapter.SideEffectClient{}, fmt.Errorf("%s requires default discord session", feature)
	}
	return discordadapter.SideEffectClient{Session: concrete}, nil
}
