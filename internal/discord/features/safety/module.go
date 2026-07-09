package safety

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/safety"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service coreservice.ConfigService
	usage   ports.UsageTracker
}

func NewModule(repo ports.AntiScamConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.NewConfigService(repo),
		usage:   usage,
	}
}

func (m Module) Name() string {
	return "anti-scam-config"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(AntiScamCommandName, m.ToggleHandler())
}
