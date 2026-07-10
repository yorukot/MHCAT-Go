package safety

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAntiScamMessageCreateDeletesScamURLAndWarns(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewMessageDeleteModule(configs, catalog, sideEffects)

	if err := module.MessageCreateHandler()(context.Background(), antiScamMessageEvent("check https://bad.example now")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.DeletedMessage) != 1 || sideEffects.DeletedMessage[0].ChannelID != "channel-1" || sideEffects.DeletedMessage[0].MessageID != "message-1" {
		t.Fatalf("deleted messages = %#v", sideEffects.DeletedMessage)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "channel-1" || sideEffects.Sent[0].Message.Content != antiScamDeleteWarning {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
}

func TestAntiScamMessageCreateIgnoresDisabledCleanAndDMEvents(t *testing.T) {
	for name, event := range map[string]events.Event{
		"wrong type": {Type: events.TypeMessageDelete},
		"dm": func() events.Event {
			event := antiScamMessageEvent("https://bad.example")
			event.GuildID = ""
			return event
		}(),
		"clean": antiScamMessageEvent("hello"),
		"empty content": func() events.Event {
			event := antiScamMessageEvent("")
			return event
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			configs := fakemongo.NewAntiScamConfigRepository()
			configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
			catalog := fakemongo.NewScamURLCatalogRepository()
			catalog.Known = []string{"https://bad.example"}
			sideEffects := fakediscord.NewSideEffects()
			module := NewMessageDeleteModule(configs, catalog, sideEffects)

			if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
				t.Fatalf("message create: %v", err)
			}
			if len(sideEffects.DeletedMessage) != 0 || len(sideEffects.Sent) != 0 {
				t.Fatalf("side effects deleted=%#v sent=%#v", sideEffects.DeletedMessage, sideEffects.Sent)
			}
		})
	}

	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: false}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewMessageDeleteModule(configs, catalog, sideEffects)
	if err := module.MessageCreateHandler()(context.Background(), antiScamMessageEvent("https://bad.example")); err != nil {
		t.Fatalf("disabled message create: %v", err)
	}
	if len(sideEffects.DeletedMessage) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("disabled side effects deleted=%#v sent=%#v", sideEffects.DeletedMessage, sideEffects.Sent)
	}
}

func TestAntiScamMessageEventRegisteredOnlyWithPorts(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewMessageDeleteModule(fakemongo.NewAntiScamConfigRepository(), fakemongo.NewScamURLCatalogRepository(), fakediscord.NewSideEffects()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("expected anti-scam message handler")
	}

	empty := events.NewDispatcher(nil)
	Module{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("unexpected anti-scam message handler")
	}
}

func antiScamMessageEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		MessageID: "message-1",
		UserID:    "user-1",
		Content:   content,
	}
}
