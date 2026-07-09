package gacha

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/gacha"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service coreservice.PrizePoolService
	discord ports.DiscordInfoProvider
	usage   ports.UsageTracker
	color   func() int
}

func NewModule(repo ports.GachaPrizePoolRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.PrizePoolService{Repository: repo},
		discord: discord,
		usage:   usage,
		color:   legacyRandomColor,
	}
}

func NewModuleWithColor(repo ports.GachaPrizePoolRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker, color func() int) Module {
	module := NewModule(repo, discord, usage)
	if color != nil {
		module.color = color
	}
	return module
}

func (m Module) Name() string {
	return "gacha-prize-list"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(GachaPrizeListCommandName, m.PrizeListHandler())
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0x5865F2
	}
	return int(value.Int64())
}
