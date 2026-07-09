package lottery

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	usage ports.UsageTracker
}

func NewModule(usage ports.UsageTracker) Module {
	return Module{usage: usage}
}

func (m Module) Name() string {
	return "lottery-disabled-command"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(LotteryCreateCommandName, m.CreateHandler())
}
