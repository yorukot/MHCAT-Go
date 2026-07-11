package utility

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	ping           coreutility.PingService
	help           coreutility.HelpService
	status         coreutility.StatusService
	translate      coreutility.TranslateService
	translateColor func() int
	discord        ports.DiscordInfoProvider
	usage          ports.UsageTracker
	defs           []commands.Definition
	feature        string
}

func NewModule(registry commands.Registry, botInfo ports.BotInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return NewModuleWithDiscordInfo(registry, botInfo, nil, clock, usage)
}

func NewModuleWithDiscordInfo(registry commands.Registry, botInfo ports.BotInfoProvider, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	if len(registry.Commands) == 0 {
		registry = commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	}
	return Module{
		ping:           coreutility.PingService{Clock: clock},
		help:           coreutility.NewHelpService(registry),
		status:         coreutility.StatusService{Provider: botInfo},
		discord:        discordInfo,
		usage:          usage,
		defs:           commands.BuiltinDefinitions(),
		feature:        "utility",
		translateColor: legacyTranslateRandomColor,
	}
}

func NewModuleWithTranslator(registry commands.Registry, botInfo ports.BotInfoProvider, discordInfo ports.DiscordInfoProvider, translator ports.Translator, clock ports.Clock, usage ports.UsageTracker) Module {
	module := NewModuleWithDiscordInfo(registry, botInfo, discordInfo, clock, usage)
	module.translate = coreutility.TranslateService{Translator: translator}
	module.defs = append(module.defs, commands.TranslateDefinition())
	return module
}

func (m Module) Name() string {
	return m.feature
}

func (m Module) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.defs...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash("help", m.HelpHandler()); err != nil {
		return err
	}
	if err := router.RegisterSlash("info", m.InfoHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "v1",
		Feature: "help",
		Action:  "category",
	}, m.HelpComponentHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "legacy",
		Feature: "help",
		Action:  "category_select",
		Legacy:  true,
	}, m.HelpComponentHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "legacy",
		Feature: "info",
		Action:  "bot_refresh",
		Legacy:  true,
	}, m.InfoBotRefreshHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "legacy",
		Feature: "info",
		Action:  "shard_refresh",
		Legacy:  true,
	}, m.InfoShardRefreshHandler()); err != nil {
		return err
	}
	if m.translate.Translator != nil {
		if err := router.RegisterSlash("翻譯", m.TranslateHandler()); err != nil {
			return err
		}
	}
	return router.RegisterSlash("ping", m.PingHandler())
}

func legacyTranslateRandomColor() int {
	value, err := rand.Int(rand.Reader, big.NewInt(0xFFFFFF+1))
	if err != nil {
		return 0x5865F2
	}
	return int(value.Int64())
}

func (m Module) randomTranslateColor() int {
	if m.translateColor == nil {
		return legacyTranslateRandomColor()
	}
	return m.translateColor()
}
