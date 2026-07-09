package safety

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/safety"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	configService coreservice.ConfigService
	reportService coreservice.ReportService
	usage         ports.UsageTracker
	configEnabled bool
	reportEnabled bool
}

func NewModule(repo ports.AntiScamConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		configService: coreservice.NewConfigService(repo),
		usage:         usage,
		configEnabled: true,
	}
}

func NewReportModule(catalog ports.ScamURLCatalog, sender ports.ScamReportSender, usage ports.UsageTracker) Module {
	return Module{
		reportService: coreservice.NewReportService(catalog, sender),
		usage:         usage,
		reportEnabled: true,
	}
}

func NewModuleWithReport(repo ports.AntiScamConfigRepository, catalog ports.ScamURLCatalog, sender ports.ScamReportSender, usage ports.UsageTracker) Module {
	return Module{
		configService: coreservice.NewConfigService(repo),
		reportService: coreservice.NewReportService(catalog, sender),
		usage:         usage,
		configEnabled: repo != nil,
		reportEnabled: catalog != nil && sender != nil,
	}
}

func (m Module) Name() string {
	return "anti-scam-config"
}

func (m Module) Commands() []commands.Definition {
	definitions := []commands.Definition{}
	if m.configEnabled {
		definitions = append(definitions, ConfigDefinitions()...)
	}
	if m.reportEnabled {
		definitions = append(definitions, ReportDefinitions()...)
	}
	return definitions
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.configEnabled {
		if err := router.RegisterSlash(AntiScamCommandName, m.ToggleHandler()); err != nil {
			return err
		}
	}
	if m.reportEnabled {
		if err := router.RegisterSlash(ScamReportCommandName, m.ReportHandler()); err != nil {
			return err
		}
	}
	return nil
}
