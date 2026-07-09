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
	if a.cfg.DiscordEnableGateway {
		removeHandler = discordSession.RegisterInteractionHandler(dispatcher.Handler())
		if gatewayEvents, ok := discordSession.(GatewayEventSession); ok {
			eventDispatcher, err := a.eventFactory(a.cfg, a.logger, discordSession, mongoClient)
			if err != nil {
				removeHandler()
				_ = mongoClient.Disconnect(context.Background())
				_ = discordSession.Close()
				return fmt.Errorf("create gateway event dispatcher: %w", err)
			}
			removeEventHandler = gatewayEvents.RegisterGatewayEventHandlers(eventDispatcher, discordadapter.GatewayEventOptions{
				Messages:         a.cfg.DiscordGuildMessagesIntent,
				MessageReactions: a.cfg.DiscordMessageReactionsIntent,
				GuildMembers:     a.cfg.DiscordGuildMembersIntent,
				VoiceStates:      a.cfg.DiscordVoiceStateIntent,
			})
		}
		if err := openWithTimeout(ctx, discordSession, a.cfg.DiscordGatewayConnectTimeout); err != nil {
			removeHandler()
			removeEventHandler()
			_ = mongoClient.Disconnect(context.Background())
			return fmt.Errorf("open discord gateway: %w", err)
		}
	}

	a.mu.Lock()
	a.mongo = mongoClient
	a.discord = discordSession
	a.removeInteractionHandler = removeHandler
	a.removeGatewayEventHandler = removeEventHandler
	a.mu.Unlock()

	a.logger.Info("mhcat skeleton initialized", "gateway_enabled", a.cfg.DiscordEnableGateway)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.shutdownOnce.Do(func() {
		a.mu.Lock()
		mongoClient := a.mongo
		discordSession := a.discord
		a.mu.Unlock()

		var errs []error
		if discordSession != nil {
			if a.removeInteractionHandler != nil {
				a.removeInteractionHandler()
			}
			if a.removeGatewayEventHandler != nil {
				a.removeGatewayEventHandler()
			}
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
	opts := RuntimeOptions{
		Config:                        cfg,
		Logger:                        logger,
		Session:                       session,
		UsageTracker:                  usageadapter.NoopTracker{},
		WorkFeatureEnabled:            cfg.FeatureWorkEnabled,
		WarningsFeatureEnabled:        cfg.FeatureWarningsEnabled,
		TranslateFeatureEnabled:       cfg.FeatureTranslateEnabled,
		LotteryDisabledCommandEnabled: cfg.FeatureLotteryDisabledCommandEnabled,
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
	if cfg.FeatureEconomyQueryEnabled || cfg.FeatureEconomySignInEnabled {
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
		opts.AutoNotificationRepository = autoNotificationRepo
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
	if cfg.FeatureGachaPrizeListEnabled {
		gachaRepo, err := gachaPrizePoolRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.GachaPrizePoolRepository = gachaRepo
	}
	if cfg.FeatureBirthdayConfigEnabled {
		birthdayRepo, err := birthdayConfigRepositoryFromMongo(mongoClient)
		if err != nil {
			return nil, err
		}
		opts.BirthdayConfigRepository = birthdayRepo
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
	if cfg.FeatureXPProfileDisabledEnabled {
		opts.XPProfileDisabledEnabled = true
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

func autoNotificationScheduleRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.AutoNotificationScheduleRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("auto-notification config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("auto-notification config feature database: %w", err)
	}
	repo, err := mongorepositories.NewAutoNotificationScheduleRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("auto-notification config feature repository: %w", err)
	}
	return repo, nil
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
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("anti-scam report feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("anti-scam report feature database: %w", err)
	}
	repo, err := mongorepositories.NewScamURLCatalogRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("anti-scam report feature repository: %w", err)
	}
	return repo, nil
}

func gachaPrizePoolRepositoryFromMongo(mongoClient MongoClient) (*mongorepositories.GachaRepository, error) {
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("gacha prize-list feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("gacha prize-list feature database: %w", err)
	}
	repo, err := mongorepositories.NewGachaRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("gacha prize-list feature repository: %w", err)
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
	concrete, ok := mongoClient.(*mongoadapter.Client)
	if !ok {
		return nil, fmt.Errorf("welcome-message config feature requires default mongo client")
	}
	database, err := concrete.Database()
	if err != nil {
		return nil, fmt.Errorf("welcome-message config feature database: %w", err)
	}
	repo, err := mongorepositories.NewLeaveMessageConfigRepositoryFromDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("welcome-message config feature repository: %w", err)
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
