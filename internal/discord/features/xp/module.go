package xp

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
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

func (m Module) Name() string {
	return "text-xp-config"
}

func (m VoiceModule) Name() string {
	return "voice-xp-config"
}

func (m DisabledProfileModule) Name() string {
	return "xp-profile-disabled"
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
