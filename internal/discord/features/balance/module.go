package balance

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service coreutility.BalanceService
	usage   ports.UsageTracker
}

func NewModule(repo ports.BalanceRepository, usage ports.UsageTracker) Module {
	return Module{
		service: coreutility.BalanceService{Repository: repo},
		usage:   usage,
	}
}

func (m Module) Name() string {
	return "balance-query"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(CommandName, m.Handler())
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: CommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     m.Name(),
	})
}
