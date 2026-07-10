package logging

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	moderationservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/moderation"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service              moderationservice.LoggingConfigService
	configReader         ports.LoggingConfigReader
	messages             ports.DiscordMessagePort
	auditLogs            ports.DiscordAuditLogPort
	usage                ports.UsageTracker
	clock                ports.Clock
	messageEventsEnabled bool
	channelEventsEnabled bool
	voiceEventsEnabled   bool
}

func NewModule(repo ports.LoggingConfigRepository, usage ports.UsageTracker) Module {
	return NewModuleWithClock(repo, usage, nil)
}

func NewModuleWithClock(repo ports.LoggingConfigRepository, usage ports.UsageTracker, clock ports.Clock) Module {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Module{
		service: moderationservice.NewLoggingConfigService(repo),
		usage:   usage,
		clock:   clock,
	}
}

func (m Module) Name() string {
	return "logging"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(LoggingConfigCommandName, m.ConfigPromptHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "logging", Action: "configure"}, m.ConfigSelectHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.LegacyVersion, Feature: "logging", Action: "configure_select", Legacy: true}, m.LegacyConfigSelectHandler())
}
