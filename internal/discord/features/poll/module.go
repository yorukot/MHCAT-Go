package poll

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	repo       ports.PollRepository
	messages   ports.DiscordMessagePort
	members    ports.DiscordGuildMemberReader
	clock      ports.Clock
	defs       []commands.Definition
	feature    string
	memberPerm int64
}

func NewModule(repo ports.PollRepository) Module {
	return Module{
		repo:       repo,
		defs:       Definitions(),
		feature:    "poll",
		memberPerm: permissionManageMessages,
	}
}

func NewModuleWithSideEffects(repo ports.PollRepository, messages ports.DiscordMessagePort, members ports.DiscordGuildMemberReader, clock ports.Clock) Module {
	module := NewModule(repo)
	module.messages = messages
	module.members = members
	module.clock = clock
	return module
}

func (m Module) Name() string {
	return m.feature
}

func (m Module) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.defs...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash("投票創建", m.CreateHandler()); err != nil {
		return err
	}
	routes := []struct {
		key     interactions.RouteKey
		handler interactions.Handler
	}{
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "poll", Action: "vote", Legacy: true}, m.VoteHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "poll", Action: "result", Legacy: true}, m.ResultHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "poll", Action: "owner_menu", Legacy: true}, m.OwnerMenuHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "poll", Action: "max_choices", Legacy: true}, m.MaxChoicesHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "poll", Action: "vote"}, m.VoteHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "poll", Action: "result"}, m.ResultHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "poll", Action: "owner_menu"}, m.OwnerMenuHandler()},
		{interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "poll", Action: "max_choices"}, m.MaxChoicesHandler()},
	}
	for _, route := range routes {
		if err := router.RegisterRoute(route.key, route.handler); err != nil {
			return err
		}
	}
	return nil
}
