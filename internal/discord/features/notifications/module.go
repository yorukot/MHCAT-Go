package notifications

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/notifications"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service  coreservice.ScheduleService
	discord  ports.DiscordInfoProvider
	messages ports.DiscordMessagePort
	usage    ports.UsageTracker
	color    func() int
}

func NewModule(repo ports.AutoNotificationScheduleRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.NewScheduleService(repo),
		discord: discord,
		usage:   usage,
		color:   legacyRandomColor,
	}
}

func NewModuleWithMessagePort(repo ports.AutoNotificationScheduleRepository, discord ports.DiscordInfoProvider, messages ports.DiscordMessagePort, usage ports.UsageTracker) Module {
	module := NewModule(repo, discord, usage)
	module.messages = messages
	return module
}

func NewModuleWithColor(repo ports.AutoNotificationScheduleRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker, color func() int) Module {
	module := NewModule(repo, discord, usage)
	if color != nil {
		module.color = color
	}
	return module
}

func (m Module) Name() string {
	return "auto-notification-config"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(AutoNotificationSetupCommandName, m.SetupHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(AutoNotificationListCommandName, m.ListHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash(AutoNotificationDeleteCommandName, m.DeleteHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.LegacyVersion, Feature: "cron", Action: "submit", Legacy: true}, m.SetupModalHandler())
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0x5865F2
	}
	return int(value.Int64())
}
