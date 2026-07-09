package ticket

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	repo      ports.TicketConfigRepository
	usage     ports.UsageTracker
	channels  ports.DiscordChannelPort
	messages  ports.DiscordMessagePort
	botUserID string
	defs      []commands.Definition
	feature   string
}

func NewModule(repo ports.TicketConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		repo:    repo,
		usage:   usage,
		defs:    Definitions(),
		feature: "ticket",
	}
}

func NewModuleWithSideEffects(repo ports.TicketConfigRepository, usage ports.UsageTracker, channels ports.DiscordChannelPort, messages ports.DiscordMessagePort, botUserID string) Module {
	module := NewModule(repo, usage)
	module.channels = channels
	module.messages = messages
	module.botUserID = botUserID
	return module
}

func (m Module) Name() string {
	return m.feature
}

func (m Module) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.defs...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash("私人頻道設置", m.SetupHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash("私人頻道刪除", m.DeleteHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeModal,
		Version: "v1",
		Feature: "ticket",
		Action:  "setup",
	}, m.SetupModalHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeModal,
		Version: "legacy",
		Feature: "ticket",
		Action:  "panel_submit",
		Legacy:  true,
	}, m.LegacyPanelSubmitHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "legacy",
		Feature: "ticket",
		Action:  "open",
		Legacy:  true,
	}, m.OpenHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "legacy",
		Feature: "ticket",
		Action:  "close",
		Legacy:  true,
	}, m.CloseHandler())
}
