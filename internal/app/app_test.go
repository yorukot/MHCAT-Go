package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	discordadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func TestNewMissingConfigReturnsError(t *testing.T) {
	if _, err := New(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil {
		t.Fatal("expected config error")
	}
}

func TestRunInitializesAndShutsDownWithoutGateway(t *testing.T) {
	cfg := validTestConfig()
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("run app: %v", err)
	}
	if mongo.connects != 1 || mongo.disconnects != 1 {
		t.Fatalf("unexpected mongo lifecycle: connects=%d disconnects=%d", mongo.connects, mongo.disconnects)
	}
	if discord.opens != 0 {
		t.Fatalf("gateway opened while disabled: %d", discord.opens)
	}
	if discord.closes != 1 {
		t.Fatalf("expected discord close to be idempotently called once, got %d", discord.closes)
	}
}

func TestTicketFeatureRequiresDefaultRuntimeAdapters(t *testing.T) {
	cfg := validTestConfig()
	cfg.FeatureTicketsEnabled = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := application.Start(context.Background()); err == nil {
		t.Fatal("expected ticket feature to reject fake runtime adapters")
	}
	if mongo.connects != 1 || mongo.disconnects != 1 {
		t.Fatalf("mongo should be cleaned up after runtime wiring failure: connects=%d disconnects=%d", mongo.connects, mongo.disconnects)
	}
	if discord.closes != 1 {
		t.Fatalf("discord session should be closed after runtime wiring failure, got %d", discord.closes)
	}
}

func TestAntiScamConfigFeatureRequiresDefaultRuntimeAdapters(t *testing.T) {
	cfg := validTestConfig()
	cfg.FeatureAntiScamConfigEnabled = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := application.Start(context.Background()); err == nil {
		t.Fatal("expected anti-scam config feature to reject fake runtime adapters")
	}
	if mongo.connects != 1 || mongo.disconnects != 1 {
		t.Fatalf("mongo should be cleaned up after runtime wiring failure: connects=%d disconnects=%d", mongo.connects, mongo.disconnects)
	}
	if discord.closes != 1 {
		t.Fatalf("discord session should be closed after runtime wiring failure, got %d", discord.closes)
	}
}

func TestWelcomeMessageDeliveryRequiresDefaultRuntimeAdapters(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	cfg.DiscordGuildMembersIntent = true
	cfg.FeatureWelcomeMessageDeliveryEnabled = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := application.Start(context.Background()); err == nil {
		t.Fatal("expected welcome-message delivery feature to reject fake runtime adapters")
	}
	if mongo.connects != 1 || mongo.disconnects != 1 {
		t.Fatalf("mongo should be cleaned up after event runtime wiring failure: connects=%d disconnects=%d", mongo.connects, mongo.disconnects)
	}
	if discord.closes != 1 {
		t.Fatalf("discord session should be closed after event runtime wiring failure, got %d", discord.closes)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	cfg := validTestConfig()
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := application.Start(context.Background()); err != nil {
		t.Fatalf("start app: %v", err)
	}
	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}
	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("second shutdown: %v", err)
	}
	if mongo.disconnects != 1 {
		t.Fatalf("expected one mongo disconnect, got %d", mongo.disconnects)
	}
	if discord.closes != 1 {
		t.Fatalf("expected one discord close, got %d", discord.closes)
	}
}

func TestGatewayWaitsForContext(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	time.Sleep(10 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("run app: %v", err)
	}
	if discord.opens != 1 {
		t.Fatalf("expected gateway open once, got %d", discord.opens)
	}
	if discord.registers != 1 {
		t.Fatalf("expected interaction handler registration once, got %d", discord.registers)
	}
	if discord.eventRegisters != 1 {
		t.Fatalf("expected gateway event registration once, got %d", discord.eventRegisters)
	}
	if discord.lastEventOptions.Messages || discord.lastEventOptions.MessageReactions || discord.lastEventOptions.GuildMembers || discord.lastEventOptions.VoiceStates {
		t.Fatalf("event options should be disabled by default: %#v", discord.lastEventOptions)
	}
	if discord.closes != 1 {
		t.Fatalf("expected gateway close once, got %d", discord.closes)
	}
}

func TestGatewayEventOptionsFollowConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	cfg.DiscordGuildMessagesIntent = true
	cfg.DiscordMessageReactionsIntent = true
	cfg.DiscordGuildMembersIntent = true
	cfg.DiscordVoiceStateIntent = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	time.Sleep(10 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run app: %v", err)
	}
	if discord.eventRegisters != 1 {
		t.Fatalf("expected one event registration, got %d", discord.eventRegisters)
	}
	if !discord.lastEventOptions.Messages || !discord.lastEventOptions.MessageReactions || !discord.lastEventOptions.GuildMembers || !discord.lastEventOptions.VoiceStates {
		t.Fatalf("event options not propagated: %#v", discord.lastEventOptions)
	}
}

func TestGatewayUsesConfiguredEventRuntimeFactory(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	mongo := &fakeMongo{}
	discord := &fakeDiscord{}
	wantDispatcher := discordevents.NewDispatcher(nil)
	wantDispatcher.Register(discordevents.TypeMessageCreate, func(context.Context, discordevents.Event) error { return nil })
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
		WithEventRuntimeFactory(func(config.Config, *slog.Logger, DiscordSession, MongoClient) (*discordevents.Dispatcher, error) {
			return wantDispatcher, nil
		}),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	time.Sleep(10 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run app: %v", err)
	}
	if discord.lastEventDispatcher != wantDispatcher {
		t.Fatalf("event dispatcher not registered from factory")
	}
	if !discord.lastEventDispatcher.HasHandlers(discordevents.TypeMessageCreate) {
		t.Fatal("expected message handler on event dispatcher")
	}
}

func TestGatewaySmokeWaitsReadyAndShutsDown(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	cfg.DiscordGatewaySmokeTest = true
	cfg.DiscordGatewaySmokeTimeout = time.Second
	cfg.Staging = config.StagingConfig{
		Mode:              true,
		AllowGatewaySmoke: true,
		SmokeTimeout:      time.Second,
	}
	mongo := &fakeMongo{}
	discord := &fakeDiscord{ready: make(chan struct{})}
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	close(discord.ready)
	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("run app: %v", err)
	}
	if discord.opens != 1 || discord.closes != 1 || discord.registers != 1 {
		t.Fatalf("gateway lifecycle opens=%d closes=%d registers=%d", discord.opens, discord.closes, discord.registers)
	}
}

func validTestConfig() config.Config {
	return config.Config{
		Env:                          config.DefaultEnv,
		LogLevel:                     config.DefaultLogLevel,
		LogFormat:                    config.DefaultLogFormat,
		DiscordToken:                 "test-token",
		MongoDBURI:                   "mongodb://localhost:27017/mhcat",
		MongoDBDatabase:              "mhcat",
		MongoConnectTimeout:          config.DefaultMongoConnectTimeout,
		MongoPingTimeout:             config.DefaultMongoPingTimeout,
		ShutdownTimeout:              config.DefaultShutdownTimeout,
		DiscordGatewayConnectTimeout: config.DefaultGatewayConnectTimeout,
		DiscordInteractionTimeout:    config.DefaultInteractionTimeout,
		DiscordGatewaySmokeTimeout:   config.DefaultGatewaySmokeTimeout,
	}
}

type fakeMongo struct {
	connects    int
	disconnects int
}

func (f *fakeMongo) Connect(context.Context) error {
	f.connects++
	return nil
}

func (f *fakeMongo) Disconnect(context.Context) error {
	f.disconnects++
	return nil
}

type fakeDiscord struct {
	opens               int
	closes              int
	registers           int
	eventRegisters      int
	lastEventOptions    discordadapter.GatewayEventOptions
	lastEventDispatcher *discordevents.Dispatcher
	ready               chan struct{}
}

func (f *fakeDiscord) Open() error {
	f.opens++
	return nil
}

func (f *fakeDiscord) Close() error {
	f.closes++
	return nil
}

func (f *fakeDiscord) RegisterInteractionHandler(discordruntime.Handler) func() {
	f.registers++
	return func() {}
}

func (f *fakeDiscord) RegisterGatewayEventHandlers(dispatcher *discordevents.Dispatcher, opts discordadapter.GatewayEventOptions) func() {
	f.eventRegisters++
	f.lastEventOptions = opts
	f.lastEventDispatcher = dispatcher
	return func() {}
}

func (f *fakeDiscord) Ready() <-chan struct{} {
	if f.ready == nil {
		f.ready = make(chan struct{})
	}
	return f.ready
}
