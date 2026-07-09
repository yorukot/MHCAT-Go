package roles

import (
	"context"
	"fmt"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/roles"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service     coreservice.SelectionService
	messages    ports.DiscordMessagePort
	direct      ports.DiscordDirectMessagePort
	usage       ports.UsageTracker
	idGenerator func() string
}

func NewModule(repo ports.RoleSelectionRepository, roles ports.DiscordRolePort, inspector ports.DiscordRoleInspector, reactions ports.DiscordReactionPort, messages ports.DiscordMessagePort, direct ports.DiscordDirectMessagePort, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.SelectionService{
			Repository:    repo,
			RoleInspector: inspector,
			Roles:         roles,
			Reactions:     reactions,
		},
		messages:    messages,
		direct:      direct,
		usage:       usage,
		idGenerator: legacyRoleButtonID,
	}
}

func NewModuleWithIDGenerator(repo ports.RoleSelectionRepository, roles ports.DiscordRolePort, inspector ports.DiscordRoleInspector, reactions ports.DiscordReactionPort, messages ports.DiscordMessagePort, direct ports.DiscordDirectMessagePort, usage ports.UsageTracker, idGenerator func() string) Module {
	module := NewModule(repo, roles, inspector, reactions, messages, direct, usage)
	if idGenerator != nil {
		module.idGenerator = idGenerator
	}
	return module
}

func (m Module) Name() string {
	return "role-selection"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(RoleReactionSetCommandName, m.ReactionSetHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(RoleReactionDeleteCommandName, m.ReactionDeleteHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(RoleButtonCommandName, m.ButtonSetupHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.LegacyVersion, Feature: "role_button", Action: "modal_submit", Legacy: true}, m.ButtonModalHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.LegacyVersion, Feature: "role_button", Action: "add", Legacy: true}, m.ButtonApplyHandler(false)); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.LegacyVersion, Feature: "role_button", Action: "remove", Legacy: true}, m.ButtonApplyHandler(true))
}

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil {
		return
	}
	dispatcher.Register(events.TypeReactionAdd, m.ReactionEventHandler(false))
	dispatcher.Register(events.TypeReactionRemove, m.ReactionEventHandler(true))
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		GuildID:     interaction.Actor.GuildID,
		UserID:      interaction.Actor.UserID,
		CommandName: commandName,
		Feature:     "role-selection",
	})
}

func legacyRoleButtonID() string {
	now := time.Now().UTC().Add(8 * time.Hour)
	return fmt.Sprintf("%s%d", now.Format("200601021504"), now.UnixNano()%10000000000)
}
