package xp

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextXPEventAccruesMessageXP(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewTextEventModule(repo)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	profile := repo.TextProfiles["guild-1/user-1"]
	if profile.GuildID != "guild-1" || profile.UserID != "user-1" || profile.XP != 5 || profile.Level != 0 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestTextXPEventIgnoresBotDMAndNonMessageEvents(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewTextEventModule(repo)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	bot := textXPEvent("bot")
	bot.IsBot = true
	for _, event := range []events.Event{
		{Type: events.TypeMessageUpdate, GuildID: "guild-1", UserID: "user-1", Content: "edit"},
		{Type: events.TypeMessageCreate, UserID: "user-1", Content: "dm"},
		{Type: events.TypeMessageCreate, GuildID: "guild-1", Content: "missing user"},
		bot,
	} {
		if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
			t.Fatalf("ignored event returned error: %v", err)
		}
	}
	if len(repo.TextProfiles) != 0 {
		t.Fatalf("unexpected profiles = %#v", repo.TextProfiles)
	}
}

func TestTextXPEventRegisteredOnlyWithRepository(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewTextEventModule(fakemongo.NewXPAdminRepository()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("expected text XP message handler")
	}

	empty := events.NewDispatcher(nil)
	TextEventModule{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("unexpected text XP message handler")
	}
}

func textXPEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   content,
	}
}

func fixedTextEventMultiplier(value int64) func() int64 {
	return func() int64 { return value }
}
