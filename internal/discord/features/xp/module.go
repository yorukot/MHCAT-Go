package xp

import (
	"log/slog"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service  coreservice.TextConfigService
	messages ports.DiscordMessagePort
	usage    ports.UsageTracker
}

type VoiceModule struct {
	service  coreservice.VoiceConfigService
	messages ports.DiscordMessagePort
	usage    ports.UsageTracker
}

type DisabledProfileModule struct {
	usage ports.UsageTracker
}

type RewardRoleModule struct {
	textService  coreservice.TextRewardRoleService
	voiceService coreservice.VoiceRewardRoleService
	roleCache    ports.DiscordCachedRoleReader
	usage        ports.UsageTracker
	color        func() int
}

type AdminModule struct {
	service coreservice.AdminService
	usage   ports.UsageTracker
}

type ResetModule struct {
	service       coreservice.ResetService
	guilds        ports.DiscordInfoProvider
	messages      ports.DiscordMessagePort
	usage         ports.UsageTracker
	clock         ports.Clock
	confirmations *resetConfirmationStore
}

type RankModule struct {
	service coreservice.RankService
	guilds  ports.DiscordInfoProvider
	usage   ports.UsageTracker
}

type TextEventModule struct {
	service     coreservice.TextAccrualService
	configs     ports.TextXPConfigReader
	messages    ports.DiscordMessagePort
	channels    ports.DiscordChannelPort
	direct      ports.DiscordDirectMessagePort
	rewardRoles coreservice.TextRewardRoleService
	coinRewards coreservice.TextCoinRewardService
}

type VoiceEventModule struct {
	service     coreservice.VoiceSessionService
	accrual     coreservice.VoiceAccrualService
	configs     ports.VoiceXPConfigReader
	messages    ports.DiscordMessagePort
	channels    ports.DiscordChannelPort
	direct      ports.DiscordDirectMessagePort
	guilds      ports.DiscordInfoProvider
	rewardRoles coreservice.VoiceRewardRoleService
	coinRewards coreservice.VoiceCoinRewardService
	worker      *VoiceXPWorker
	voiceStates ports.DiscordVoiceStateReader
	coordinator *voiceXPGuildCoordinator
	logger      *slog.Logger
}

func NewModule(repo ports.TextXPConfigRepository, messages ports.DiscordMessagePort, usage ports.UsageTracker) Module {
	return Module{
		service:  coreservice.TextConfigService{Repository: repo},
		messages: messages,
		usage:    usage,
	}
}

func NewVoiceModule(repo ports.VoiceXPConfigRepository, messages ports.DiscordMessagePort, usage ports.UsageTracker) VoiceModule {
	return VoiceModule{
		service:  coreservice.VoiceConfigService{Repository: repo},
		messages: messages,
		usage:    usage,
	}
}

func NewDisabledProfileModule(usage ports.UsageTracker) DisabledProfileModule {
	return DisabledProfileModule{usage: usage}
}

func NewRewardRoleModule(textRepo ports.TextXPRewardRoleRepository, voiceRepo ports.VoiceXPRewardRoleRepository, roles ports.DiscordRoleInspector, usage ports.UsageTracker) RewardRoleModule {
	module := RewardRoleModule{
		textService:  coreservice.TextRewardRoleService{Repository: textRepo, RoleInspector: roles},
		voiceService: coreservice.VoiceRewardRoleService{Repository: voiceRepo, RoleInspector: roles},
		usage:        usage,
		color:        randomXPColor,
	}
	if roleCache, ok := roles.(ports.DiscordCachedRoleReader); ok {
		module.roleCache = roleCache
	}
	return module
}

func NewAdminModule(repo ports.XPAdminRepository, usage ports.UsageTracker) AdminModule {
	return AdminModule{
		service: coreservice.AdminService{Repository: repo},
		usage:   usage,
	}
}

func NewResetModule(repo ports.XPResetRepository, guilds ports.DiscordInfoProvider, messages ports.DiscordMessagePort, usage ports.UsageTracker, clock ports.Clock) ResetModule {
	confirmations := defaultResetConfirmationStore
	if clock == nil {
		clock = ports.SystemClock{}
	} else {
		confirmations = newResetConfirmationStore(clock, time.Minute)
	}
	return ResetModule{
		service:       coreservice.ResetService{Repository: repo},
		guilds:        guilds,
		messages:      messages,
		usage:         usage,
		clock:         clock,
		confirmations: confirmations,
	}
}

func NewRankModule(repo ports.XPRankRepository, guilds ports.DiscordInfoProvider, usage ports.UsageTracker) RankModule {
	return RankModule{
		service: coreservice.RankService{Repository: repo},
		guilds:  guilds,
		usage:   usage,
	}
}

func NewTextEventModule(repo ports.TextXPAccrualRepository, configs ports.TextXPConfigReader, messages ports.DiscordMessagePort) TextEventModule {
	return TextEventModule{
		service:  coreservice.TextAccrualService{Repository: repo},
		configs:  configs,
		messages: messages,
	}
}

func (m TextEventModule) WithRewardRoles(repo ports.TextXPRewardRoleRepository, roles ports.DiscordRolePort) TextEventModule {
	m.rewardRoles = coreservice.TextRewardRoleService{Repository: repo, RolePort: roles}
	return m
}

func (m TextEventModule) WithAnnouncementFallbacks(channels ports.DiscordChannelPort, direct ports.DiscordDirectMessagePort) TextEventModule {
	m.channels = channels
	m.direct = direct
	return m
}

func (m TextEventModule) WithCoinRewards(repo ports.TextXPCoinRewardRepository) TextEventModule {
	m.coinRewards = coreservice.TextCoinRewardService{Repository: repo}
	return m
}

func NewVoiceEventModule(repo ports.VoiceXPSessionRepository) VoiceEventModule {
	module := VoiceEventModule{
		service:     coreservice.VoiceSessionService{Repository: repo},
		coordinator: newVoiceXPGuildCoordinator(),
		logger:      slog.Default(),
	}
	if accrualRepo, ok := repo.(ports.VoiceXPAccrualRepository); ok {
		module.accrual = coreservice.VoiceAccrualService{Repository: accrualRepo}
	}
	return module
}

func (m VoiceEventModule) WithVoiceStateReader(reader ports.DiscordVoiceStateReader) VoiceEventModule {
	m.voiceStates = reader
	return m
}

func (m VoiceEventModule) WithAccrual(repo ports.VoiceXPAccrualRepository, configs ports.VoiceXPConfigReader, messages ports.DiscordMessagePort) VoiceEventModule {
	m.accrual = coreservice.VoiceAccrualService{Repository: repo}
	m.configs = configs
	m.messages = messages
	return m
}

func (m VoiceEventModule) WithAnnouncementFallbacks(channels ports.DiscordChannelPort, direct ports.DiscordDirectMessagePort, guilds ports.DiscordInfoProvider) VoiceEventModule {
	m.channels = channels
	m.direct = direct
	m.guilds = guilds
	return m
}

func (m VoiceEventModule) WithRewardRoles(repo ports.VoiceXPRewardRoleRepository, roles ports.DiscordRolePort) VoiceEventModule {
	m.rewardRoles = coreservice.VoiceRewardRoleService{Repository: repo, RolePort: roles}
	return m
}

func (m VoiceEventModule) WithCoinRewards(repo ports.VoiceXPCoinRewardRepository) VoiceEventModule {
	m.coinRewards = coreservice.VoiceCoinRewardService{Repository: repo}
	return m
}

func (m Module) Name() string {
	return "text-xp-config"
}

func (m VoiceModule) Name() string {
	return "voice-xp-config"
}

func (m DisabledProfileModule) Name() string {
	return "xp-profile-disabled"
}

func (m RewardRoleModule) Name() string {
	return "xp-role-config"
}

func (m AdminModule) Name() string {
	return "xp-admin"
}

func (m ResetModule) Name() string {
	return "xp-reset"
}

func (m RankModule) Name() string {
	return "xp-rank"
}

func (m TextEventModule) Name() string {
	return "text-xp-accrual"
}

func (m VoiceEventModule) Name() string {
	return "voice-xp-sessions"
}

func (m Module) Commands() []commands.Definition {
	return TextDefinitions()
}

func (m VoiceModule) Commands() []commands.Definition {
	return VoiceDefinitions()
}

func (m DisabledProfileModule) Commands() []commands.Definition {
	return DisabledProfileDefinitions()
}

func (m RewardRoleModule) Commands() []commands.Definition {
	return RewardRoleDefinitions()
}

func (m AdminModule) Commands() []commands.Definition {
	return AdminDefinitions()
}

func (m ResetModule) Commands() []commands.Definition {
	return ResetDefinitions()
}

func (m RankModule) Commands() []commands.Definition {
	return RankDefinitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(TextXPSetCommandName, m.SetHandler()); err != nil {
		return err
	}
	return router.RegisterSlash(TextXPDeleteCommandName, m.DeleteHandler())
}

func (m VoiceModule) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(VoiceXPSetCommandName, m.SetHandler()); err != nil {
		return err
	}
	return router.RegisterSlash(VoiceXPDeleteCommandName, m.DeleteHandler())
}

func (m DisabledProfileModule) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(TextXPProfileCommandName, m.TextHandler()); err != nil {
		return err
	}
	return router.RegisterSlash(VoiceXPProfileCommandName, m.VoiceHandler())
}

func (m RewardRoleModule) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(TextXPRewardRoleCommandName, m.TextHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(VoiceXPRewardRoleCommandName, m.VoiceHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "xp", Action: "text_reward_page", Legacy: true}, m.TextPageHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "xp", Action: "voice_reward_page", Legacy: true}, m.VoicePageHandler())
}

func (m AdminModule) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(XPAdminCommandName, m.AdminHandler())
}

func (m ResetModule) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(XPResetCommandName, m.ResetHandler())
}

func (m RankModule) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(TextXPRankCommandName, m.TextHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(VoiceXPRankCommandName, m.VoiceHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "rank", Action: "text_page", Legacy: true}, m.TextPageHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "rank", Action: "voice_page", Legacy: true}, m.VoicePageHandler())
}

func (m ResetModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher != nil {
		dispatcher.Register(events.TypeMessageCreate, m.ConfirmationHandler())
	}
}

func (m TextEventModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || m.service.Repository == nil {
		return
	}
	dispatcher.Register(events.TypeMessageCreate, m.MessageCreateHandler())
}

func (m VoiceEventModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || m.service.Repository == nil {
		return
	}
	dispatcher.Register(events.TypeVoiceState, m.VoiceStateHandler())
	if m.voiceStates != nil {
		dispatcher.Register(events.TypeGuildAvailable, m.GuildAvailableHandler())
	}
}
