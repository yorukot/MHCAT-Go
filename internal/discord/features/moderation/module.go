package moderation

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/moderation"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	warnings coremoderation.WarningHistoryService
	settings coremoderation.WarningSettingsService
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

func NewSettingsModule(repo ports.WarningSettingsRepository, usage ports.UsageTracker) Module {
	return Module{
		settings: coremoderation.WarningSettingsService{Repository: repo},
		usage:    usage,
	}
}

func (m Module) Name() string {
	return "warnings"
}

func (m Module) Commands() []commands.Definition {
	definitions := []commands.Definition{}
	if m.warnings.Repository != nil {
		definitions = append(definitions, WarningHistoryDefinition())
	}
	if m.settings.Repository != nil {
		definitions = append(definitions, WarningSettingsDefinition())
	}
	if len(definitions) > 0 {
		return definitions
	}
	return append(Definitions(), SettingsDefinitions()...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.warnings.Repository != nil {
		if err := router.RegisterSlash(WarningHistoryCommandName, m.WarningHistoryHandler()); err != nil {
			return err
		}
	}
	if m.settings.Repository != nil {
		if err := router.RegisterSlash(WarningSettingsCommandName, m.WarningSettingsHandler()); err != nil {
			return err
		}
	}
	return nil
}
