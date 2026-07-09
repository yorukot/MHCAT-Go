package moderation

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/moderation"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	warnings coremoderation.WarningHistoryService
	members  ports.DiscordGuildMemberReader
	discord  ports.DiscordInfoProvider
	usage    ports.UsageTracker
}

func NewModule(repo ports.WarningHistoryRepository, members ports.DiscordGuildMemberReader, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		warnings: coremoderation.WarningHistoryService{Repository: repo},
		members:  members,
		discord:  discord,
		usage:    usage,
	}
}

func (m Module) Name() string {
	return "warnings"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(WarningHistoryCommandName, m.WarningHistoryHandler())
}
