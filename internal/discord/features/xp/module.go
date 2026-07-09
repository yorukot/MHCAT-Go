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

type RewardRoleModule struct {
	textService  coreservice.TextRewardRoleService
	voiceService coreservice.VoiceRewardRoleService
	usage        ports.UsageTracker
	color        func() int
}

type AdminModule struct {
	service coreservice.AdminService
	usage   ports.UsageTracker
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
	return RewardRoleModule{
		textService:  coreservice.TextRewardRoleService{Repository: textRepo, RoleInspector: roles},
		voiceService: coreservice.VoiceRewardRoleService{Repository: voiceRepo, RoleInspector: roles},
		usage:        usage,
		color:        randomXPColor,
	}
}

func NewAdminModule(repo ports.XPAdminRepository, usage ports.UsageTracker) AdminModule {
	return AdminModule{
		service: coreservice.AdminService{Repository: repo},
		usage:   usage,
	}
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
