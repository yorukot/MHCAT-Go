package announcements

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/announcements"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service     coreservice.ConfigService
	reader      ports.AnnouncementChannelReader
	boundReader ports.BoundAnnouncementReader
	messages    ports.DiscordMessagePort
	usage       ports.UsageTracker
	drafts      *DraftStore
	config      bool
	send        bool
	relay       bool
}

func NewModule(repo ports.AnnouncementConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		service:     coreservice.NewConfigService(repo),
		reader:      repo,
		boundReader: repo,
		usage:       usage,
		drafts:      NewDraftStore(),
		config:      repo != nil,
	}
}

func NewSendModule(reader ports.AnnouncementChannelReader, messages ports.DiscordMessagePort, usage ports.UsageTracker) Module {
	return Module{
		reader:   reader,
		messages: messages,
		usage:    usage,
		drafts:   NewDraftStore(),
		send:     reader != nil && messages != nil,
	}
}

func NewModuleWithSend(repo ports.AnnouncementConfigRepository, messages ports.DiscordMessagePort, usage ports.UsageTracker) Module {
	return Module{
		service:     coreservice.NewConfigService(repo),
		reader:      repo,
		boundReader: repo,
		messages:    messages,
		usage:       usage,
		drafts:      NewDraftStore(),
		config:      repo != nil,
		send:        repo != nil && messages != nil,
	}
}

func NewRelayModule(reader ports.BoundAnnouncementReader, messages ports.DiscordMessagePort) Module {
	return Module{
		boundReader: reader,
		messages:    messages,
		relay:       reader != nil && messages != nil,
	}
}

func NewModuleWithRelay(repo ports.AnnouncementConfigRepository, messages ports.DiscordMessagePort, usage ports.UsageTracker) Module {
	return Module{
		service:     coreservice.NewConfigService(repo),
		reader:      repo,
		boundReader: repo,
		messages:    messages,
		usage:       usage,
		drafts:      NewDraftStore(),
		config:      repo != nil,
		send:        repo != nil && messages != nil,
		relay:       repo != nil && messages != nil,
	}
}

func (m Module) Name() string {
	return "announcement-config"
}

func (m Module) Commands() []commands.Definition {
	definitions := make([]commands.Definition, 0, 2)
	if m.config {
		definitions = append(definitions, ConfigDefinitions()...)
	}
	if m.send {
		definitions = append(definitions, SendDefinitions()...)
	}
	return definitions
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.config {
		if err := router.RegisterSlash(ConfigCommandName, m.ConfigHandler()); err != nil {
			return err
		}
	}
	if m.send {
		if err := router.RegisterSlash(SendCommandName, m.SendHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.VersionV1, Feature: announcementFeature, Action: sendModalAction}, m.SendModalHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeModal, Version: customid.LegacyVersion, Feature: announcementFeature, Action: sendModalAction, Legacy: true}, m.SendModalHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: announcementFeature, Action: confirmAction}, m.ConfirmHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: announcementFeature, Action: cancelAction}, m.CancelHandler()); err != nil {
			return err
		}
	}
	return nil
}

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if m.relay && dispatcher != nil {
		dispatcher.Register(events.TypeMessageCreate, m.RelayHandler())
	}
}
