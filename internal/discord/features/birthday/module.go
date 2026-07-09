package birthday

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/birthday"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	configService  coreservice.ConfigService
	profileService coreservice.ProfileService
	usage          ports.UsageTracker
}

func NewModule(repo ports.BirthdayConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		configService:  coreservice.NewConfigService(repo),
		profileService: coreservice.NewProfileService(repo),
		usage:          usage,
	}
}

func (m Module) Name() string {
	return "birthday-config"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(BirthdayCommandName, m.Handler())
}
