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
	service       coreservice.PrizePoolService
	deleteService coreservice.PrizeDeleteService
	discord       ports.DiscordInfoProvider
	usage         ports.UsageTracker
	color         func() int
}

func NewModule(repo ports.GachaPrizePoolRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.PrizePoolService{Repository: repo},
		discord: discord,
		usage:   usage,
		color:   legacyRandomColor,
	}
}

func NewDeleteModule(repo ports.GachaPrizeDeleteRepository, usage ports.UsageTracker) Module {
	return Module{
		deleteService: coreservice.PrizeDeleteService{Repository: repo},
		usage:         usage,
		color:         legacyRandomColor,
	}
}

func NewModuleWithRepositories(listRepo ports.GachaPrizePoolRepository, deleteRepo ports.GachaPrizeDeleteRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		service:       coreservice.PrizePoolService{Repository: listRepo},
		deleteService: coreservice.PrizeDeleteService{Repository: deleteRepo},
		discord:       discord,
		usage:         usage,
		color:         legacyRandomColor,
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
	if m.service.Repository != nil && m.deleteService.Repository != nil {
		return "gacha"
	}
	if m.deleteService.Repository != nil {
		return "gacha-prize-delete"
	}
	return "gacha-prize-list"
}

func (m Module) Commands() []commands.Definition {
	var definitions []commands.Definition
	if m.service.Repository != nil {
		definitions = append(definitions, PrizeListDefinitions()...)
	}
	if m.deleteService.Repository != nil {
		definitions = append(definitions, PrizeDeleteDefinitions()...)
	}
	return definitions
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.service.Repository != nil {
		if err := router.RegisterSlash(GachaPrizeListCommandName, m.PrizeListHandler()); err != nil {
			return err
		}
	}
	if m.deleteService.Repository != nil {
		if err := router.RegisterSlash(GachaPrizeDeleteCommandName, m.PrizeDeleteHandler()); err != nil {
			return err
		}
	}
	return nil
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0x5865F2
	}
	return int(value.Int64())
}
