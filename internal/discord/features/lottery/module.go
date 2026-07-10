package lottery

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/lottery"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service           coreservice.Service
	discord           ports.DiscordInfoProvider
	members           ports.DiscordGuildMemberReader
	messages          ports.DiscordMessagePort
	usage             ports.UsageTracker
	clock             ports.Clock
	commandEnabled    bool
	componentsEnabled bool
	color             func() int
	randomIndex       func(int) (int, error)
}

func NewModule(usage ports.UsageTracker) Module {
	return Module{usage: usage, commandEnabled: true, color: legacyLotteryRandomColor, randomIndex: lotteryCryptoRandomIndex}
}

func NewComponentModule(repo ports.LotteryRepository, discord ports.DiscordInfoProvider, members ports.DiscordGuildMemberReader, messages ports.DiscordMessagePort, clock ports.Clock) Module {
	return Module{
		service:           coreservice.NewService(repo),
		discord:           discord,
		members:           members,
		messages:          messages,
		clock:             clock,
		componentsEnabled: true,
		color:             legacyLotteryRandomColor,
		randomIndex:       lotteryCryptoRandomIndex,
	}
}

func NewModuleWithComponents(repo ports.LotteryRepository, discord ports.DiscordInfoProvider, members ports.DiscordGuildMemberReader, messages ports.DiscordMessagePort, clock ports.Clock, usage ports.UsageTracker) Module {
	module := NewComponentModule(repo, discord, members, messages, clock)
	module.commandEnabled = true
	module.usage = usage
	return module
}

func (m Module) Name() string {
	return "lottery-disabled-command"
}

func (m Module) Commands() []commands.Definition {
	if !m.commandEnabled {
		return nil
	}
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.commandEnabled {
		if err := router.RegisterSlash(LotteryCreateCommandName, m.CreateHandler()); err != nil {
			return err
		}
	}
	if !m.componentsEnabled {
		return nil
	}
	for _, route := range []struct {
		action  string
		handler interactions.Handler
	}{
		{action: "enter", handler: m.EnterHandler()},
		{action: "search", handler: m.SearchHandler()},
		{action: "reroll", handler: m.RerollHandler()},
		{action: "stop", handler: m.StopHandler()},
	} {
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.LegacyVersion, Feature: "lottery", Action: route.action, Legacy: true}, route.handler); err != nil {
			return err
		}
	}
	return nil
}

func (m Module) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func legacyLotteryRandomColor() int {
	value, err := rand.Int(rand.Reader, big.NewInt(0xFFFFFF+1))
	if err != nil {
		return 0x5865F2
	}
	return int(value.Int64())
}

func lotteryCryptoRandomIndex(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}
