package moderation

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/moderation"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	warnings      coremoderation.WarningHistoryService
	settings      coremoderation.WarningSettingsService
	issue         coremoderation.WarningIssueService
	removal       coremoderation.WarningRemovalService
	deleteData    coremoderation.DeleteDataService
	members       ports.DiscordGuildMemberReader
	discord       ports.DiscordInfoProvider
	direct        ports.DiscordDirectMessagePort
	memberActions ports.DiscordMemberPort
	hierarchy     ports.DiscordMemberHierarchyInspector
	messages      ports.DiscordMessagePort
	cleaner       ports.DiscordMessageCleaner
	clock         ports.Clock
	usage         ports.UsageTracker
}

func NewModule(repo ports.WarningHistoryRepository, members ports.DiscordGuildMemberReader, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		warnings: coremoderation.WarningHistoryService{Repository: repo},
		members:  members,
		discord:  discord,
		usage:    usage,
	}
}

func NewRemovalModule(repo ports.WarningRemovalRepository, direct ports.DiscordDirectMessagePort, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		removal: coremoderation.WarningRemovalService{Repository: repo},
		direct:  direct,
		discord: discord,
		usage:   usage,
	}
}

func NewIssueModule(repo ports.WarningIssueRepository, settings ports.WarningSettingsRepository, direct ports.DiscordDirectMessagePort, discord ports.DiscordInfoProvider, hierarchy ports.DiscordMemberHierarchyInspector, memberActions ports.DiscordMemberPort, messages ports.DiscordMessagePort, clock ports.Clock, usage ports.UsageTracker) Module {
	return Module{
		issue:         coremoderation.WarningIssueService{Repository: repo},
		settings:      coremoderation.WarningSettingsService{Repository: settings},
		direct:        direct,
		discord:       discord,
		hierarchy:     hierarchy,
		memberActions: memberActions,
		messages:      messages,
		clock:         clock,
		usage:         usage,
	}
}

func NewSettingsModule(repo ports.WarningSettingsRepository, usage ports.UsageTracker) Module {
	return Module{
		settings: coremoderation.WarningSettingsService{Repository: repo},
		usage:    usage,
	}
}

func NewCleanupModule(cleaner ports.DiscordMessageCleaner, usage ports.UsageTracker) Module {
	return Module{
		cleaner: cleaner,
		usage:   usage,
	}
}

func NewDeleteDataModule(repo ports.DeleteDataRepository, usage ports.UsageTracker) Module {
	return Module{
		deleteData: coremoderation.DeleteDataService{Repository: repo},
		usage:      usage,
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
	if m.removal.Repository != nil {
		definitions = append(definitions, RemovalDefinitions()...)
	}
	if m.issue.Repository != nil {
		definitions = append(definitions, IssueDefinitions()...)
	}
	if m.cleaner != nil {
		definitions = append(definitions, CleanupDefinitions()...)
	}
	if m.deleteData.Repository != nil {
		definitions = append(definitions, DeleteDataDefinitions()...)
	}
	if len(definitions) > 0 {
		return definitions
	}
	definitions = append(definitions, Definitions()...)
	definitions = append(definitions, SettingsDefinitions()...)
	definitions = append(definitions, RemovalDefinitions()...)
	definitions = append(definitions, IssueDefinitions()...)
	definitions = append(definitions, CleanupDefinitions()...)
	definitions = append(definitions, DeleteDataDefinitions()...)
	return definitions
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
	if m.removal.Repository != nil {
		if err := router.RegisterSlash(WarningRemoveCommandName, m.WarningRemoveHandler()); err != nil {
			return err
		}
		if err := router.RegisterSlash(WarningRemoveAllCommandName, m.WarningRemoveAllHandler()); err != nil {
			return err
		}
	}
	if m.issue.Repository != nil {
		if err := router.RegisterSlash(WarningIssueCommandName, m.WarningIssueHandler()); err != nil {
			return err
		}
	}
	if m.cleaner != nil {
		if err := router.RegisterSlash(CleanupCommandName, m.CleanupHandler()); err != nil {
			return err
		}
	}
	if m.deleteData.Repository != nil {
		if err := router.RegisterSlash(DeleteDataCommandName, m.DeleteDataPromptHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "admin", Action: "delete_data_select", Legacy: true}, m.DeleteDataSelectHandler()); err != nil {
			return err
		}
	}
	return nil
}
