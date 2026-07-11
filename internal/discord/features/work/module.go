package work

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	workservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/work"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	usage   ports.UsageTracker
	work    *workservice.Service
	discord ports.DiscordInfoProvider
	captcha captchaGenerator
	color   func() int
}

func NewModule(usage ports.UsageTracker) Module {
	return Module{usage: usage, captcha: randomCaptcha, color: randomWorkColor}
}

func NewModuleWithRepository(repo ports.WorkInterfaceRepository, usage ports.UsageTracker, clock ports.Clock) Module {
	return NewModuleWithRepositoryAndDiscordInfo(repo, nil, usage, clock)
}

func NewModuleWithRepositoryAndDiscordInfo(repo ports.WorkInterfaceRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker, clock ports.Clock) Module {
	service := workservice.NewService(repo, clock)
	return Module{usage: usage, work: &service, discord: discordInfo, captcha: randomCaptcha, color: randomWorkColor}
}

func NewModuleWithStartRepository(repo ports.WorkStartRepository, usage ports.UsageTracker, clock ports.Clock) Module {
	return NewModuleWithStartRepositoryAndDiscordInfo(repo, nil, usage, clock)
}

func NewModuleWithStartRepositoryAndDiscordInfo(repo ports.WorkStartRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker, clock ports.Clock) Module {
	service := workservice.NewServiceWithStartRepository(repo, repo, clock)
	return Module{usage: usage, work: &service, discord: discordInfo, captcha: randomCaptcha, color: randomWorkColor}
}

func NewModuleWithAdminRepository(repo ports.WorkAdminRepository, usage ports.UsageTracker, clock ports.Clock) Module {
	return NewModuleWithAdminRepositoryAndDiscordInfo(repo, nil, usage, clock)
}

func NewModuleWithAdminRepositoryAndDiscordInfo(repo ports.WorkAdminRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker, clock ports.Clock) Module {
	service := workservice.NewServiceWithAdminRepository(repo, clock)
	return Module{usage: usage, work: &service, discord: discordInfo, captcha: randomCaptcha, color: randomWorkColor}
}

func (m Module) randomColor() int {
	if m.color == nil {
		return randomWorkColor()
	}
	return m.color()
}

func (m Module) Name() string {
	return "work"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(CommandName, m.Handler()); err != nil {
		return err
	}
	if m.work == nil {
		return nil
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "work", Action: "detail"}, m.DetailHandler()); err != nil {
		return err
	}
	if !m.work.CanStart() {
		return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.VersionV1, Feature: "work", Action: "captcha"}, m.CaptchaHandler())
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "work", Action: "start"}, m.StartHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "work", Action: "override"}, m.OverrideHandler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "work", Action: "cancel"}, m.CancelHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.VersionV1, Feature: "work", Action: "captcha"}, m.CaptchaHandler())
}
