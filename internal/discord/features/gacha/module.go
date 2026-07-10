package gacha

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/gacha"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service       coreservice.PrizePoolService
	drawService   coreservice.DrawService
	createService coreservice.PrizeCreateService
	editService   coreservice.PrizeEditService
	deleteService coreservice.PrizeDeleteService
	discord       ports.DiscordInfoProvider
	messages      ports.DiscordMessagePort
	direct        ports.DiscordDirectMessagePort
	usage         ports.UsageTracker
	color         func() int
	drawWait      func(context.Context, time.Duration) error
}

func NewModule(repo ports.GachaPrizePoolRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		service: coreservice.PrizePoolService{Repository: repo},
		discord: discord,
		usage:   usage,
		color:   legacyRandomColor,
	}
}

func NewDrawModule(repo ports.GachaDrawRepository, messages ports.DiscordMessagePort, direct ports.DiscordDirectMessagePort, usage ports.UsageTracker) Module {
	return Module{
		drawService: coreservice.DrawService{Repository: repo},
		messages:    messages,
		direct:      direct,
		usage:       usage,
		color:       legacyRandomColor,
		drawWait:    waitForGachaDraw,
	}
}

func NewDeleteModule(repo ports.GachaPrizeDeleteRepository, usage ports.UsageTracker) Module {
	return Module{
		deleteService: coreservice.PrizeDeleteService{Repository: repo},
		usage:         usage,
		color:         legacyRandomColor,
	}
}

func NewCreateModule(repo ports.GachaPrizeCreateRepository, usage ports.UsageTracker) Module {
	return Module{
		createService: coreservice.PrizeCreateService{Repository: repo},
		usage:         usage,
		color:         legacyRandomColor,
	}
}

func NewEditModule(repo ports.GachaPrizeEditRepository, usage ports.UsageTracker) Module {
	return Module{
		editService: coreservice.PrizeEditService{Repository: repo},
		usage:       usage,
		color:       legacyRandomColor,
	}
}

func NewModuleWithRepositories(listRepo ports.GachaPrizePoolRepository, drawRepo ports.GachaDrawRepository, createRepo ports.GachaPrizeCreateRepository, editRepo ports.GachaPrizeEditRepository, deleteRepo ports.GachaPrizeDeleteRepository, discord ports.DiscordInfoProvider, messages ports.DiscordMessagePort, direct ports.DiscordDirectMessagePort, usage ports.UsageTracker) Module {
	return Module{
		service:       coreservice.PrizePoolService{Repository: listRepo},
		drawService:   coreservice.DrawService{Repository: drawRepo},
		createService: coreservice.PrizeCreateService{Repository: createRepo},
		editService:   coreservice.PrizeEditService{Repository: editRepo},
		deleteService: coreservice.PrizeDeleteService{Repository: deleteRepo},
		discord:       discord,
		messages:      messages,
		direct:        direct,
		usage:         usage,
		color:         legacyRandomColor,
		drawWait:      waitForGachaDraw,
	}
}

func NewModuleWithColor(repo ports.GachaPrizePoolRepository, discord ports.DiscordInfoProvider, usage ports.UsageTracker, color func() int) Module {
	module := NewModule(repo, discord, usage)
	if color != nil {
		module.color = color
	}
	return module
}

func (m Module) WithDrawWait(wait func(context.Context, time.Duration) error) Module {
	if wait != nil {
		m.drawWait = wait
	}
	return m
}

func (m Module) Name() string {
	enabled := 0
	if m.service.Repository != nil {
		enabled++
	}
	if m.drawService.Repository != nil {
		enabled++
	}
	if m.createService.Repository != nil {
		enabled++
	}
	if m.editService.Repository != nil {
		enabled++
	}
	if m.deleteService.Repository != nil {
		enabled++
	}
	if enabled > 1 {
		return "gacha"
	}
	if m.drawService.Repository != nil {
		return "gacha-draw"
	}
	if m.createService.Repository != nil {
		return "gacha-prize-create"
	}
	if m.editService.Repository != nil {
		return "gacha-prize-edit"
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
	if m.drawService.Repository != nil {
		definitions = append(definitions, DrawDefinitions()...)
	}
	if m.createService.Repository != nil {
		definitions = append(definitions, PrizeCreateDefinitions()...)
	}
	if m.editService.Repository != nil {
		definitions = append(definitions, PrizeEditDefinitions()...)
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
	if m.drawService.Repository != nil {
		if err := router.RegisterSlash(GachaDrawCommandName, m.DrawHandler()); err != nil {
			return err
		}
	}
	if m.createService.Repository != nil {
		if err := router.RegisterSlash(GachaPrizeCreateCommandName, m.PrizeCreateHandler()); err != nil {
			return err
		}
	}
	if m.editService.Repository != nil {
		if err := router.RegisterSlash(GachaPrizeEditCommandName, m.PrizeEditHandler()); err != nil {
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

func waitForGachaDraw(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
