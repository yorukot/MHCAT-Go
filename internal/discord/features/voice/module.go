package voice

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/voice"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service coreservice.ConfigService
	usage   ports.UsageTracker
}

type LockModule struct {
	service coreservice.LockService
	usage   ports.UsageTracker
}

type LockEventModule struct {
	service  coreservice.LockService
	messages ports.DiscordMessagePort
	direct   ports.DiscordDirectMessagePort
	members  ports.DiscordMemberPort
}

func NewModule(repo ports.VoiceRoomConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.NewConfigService(repo),
		usage:   usage,
	}
}

func NewLockModule(repo ports.VoiceRoomLockRepository, usage ports.UsageTracker) LockModule {
	return LockModule{
		service: coreservice.NewLockService(repo),
		usage:   usage,
	}
}

func NewLockEventModule(repo ports.VoiceRoomLockRepository, messages ports.DiscordMessagePort, direct ports.DiscordDirectMessagePort, members ports.DiscordMemberPort) LockEventModule {
	return LockEventModule{
		service:  coreservice.NewLockService(repo),
		messages: messages,
		direct:   direct,
		members:  members,
	}
}

func (m Module) Name() string {
	return "voice-room-config"
}

func (m LockModule) Name() string {
	return "voice-room-lock"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m LockModule) Commands() []commands.Definition {
	return LockDefinitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(VoiceRoomSetCommandName, m.SetHandler()); err != nil {
		return err
	}
	return router.RegisterSlash(VoiceRoomDeleteCommandName, m.DeleteHandler())
}

func (m LockModule) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(VoiceRoomLockCommandName, m.LockHandler()); err != nil {
		return err
	}
	for _, key := range []interactions.RouteKey{
		{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "voice_lock", Action: "prompt"},
		{Kind: interactions.TypeComponent, Version: customid.LegacyVersion, Feature: "voice_lock", Action: "prompt", Legacy: true},
	} {
		if err := router.RegisterRoute(key, m.PromptHandler()); err != nil {
			return err
		}
	}
	return router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeModal,
		Version: customid.LegacyVersion,
		Feature: "voice_lock",
		Action:  "answer",
		Legacy:  true,
	}, m.AnswerHandler())
}

func (m LockEventModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || m.service.Repository == nil || m.messages == nil || m.direct == nil || m.members == nil {
		return
	}
	dispatcher.Register(events.TypeVoiceState, m.VoiceStateHandler())
}
