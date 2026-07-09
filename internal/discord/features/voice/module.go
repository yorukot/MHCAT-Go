package voice

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/voice"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
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
	return router.RegisterSlash(VoiceRoomLockCommandName, m.LockHandler())
}
