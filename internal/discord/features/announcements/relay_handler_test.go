package announcements

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRelayHandlerSendsLegacyAnnouncementEmbedAndDeletesOriginal(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.BoundAnnouncements["guild-1:channel-1"] = domain.BoundAnnouncementConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Tag:       "@everyone",
		Color:     "#53FF53",
		Title:     "公告標題",
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewRelayModule(repo, sideEffects)

	err := module.RelayHandler()(context.Background(), relayEvent("這是一則公告"))
	if err != nil {
		t.Fatalf("relay: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0]
	if sent.ChannelID != "channel-1" || sent.Message.Content != "@everyone" {
		t.Fatalf("sent = %#v", sent)
	}
	if sent.Message.AllowedMentions.ParseEveryone || sent.Message.AllowedMentions.ParseRoles || sent.Message.AllowedMentions.ParseUsers {
		t.Fatalf("relay must suppress mention parsing: %#v", sent.Message.AllowedMentions)
	}
	if len(sent.Message.Embeds) != 1 {
		t.Fatalf("embeds = %#v", sent.Message.Embeds)
	}
	embed := sent.Message.Embeds[0]
	if embed.Title != "公告標題" || embed.Description != "這是一則公告" || embed.Color != 0x53FF53 {
		t.Fatalf("embed = %#v", embed)
	}
	if embed.FooterText != "來自Yoru#1234的公告" || embed.FooterIconURL == "" {
		t.Fatalf("footer = %#v", embed)
	}
	if len(sideEffects.DeletedMessage) != 1 || sideEffects.DeletedMessage[0].MessageID != "message-1" {
		t.Fatalf("deleted = %#v", sideEffects.DeletedMessage)
	}
}

func TestRelayHandlerIgnoresBotEmptyContentAndMissingConfig(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewRelayModule(repo, sideEffects)

	botEvent := relayEvent("bot")
	botEvent.IsBot = true
	if err := module.RelayHandler()(context.Background(), botEvent); err != nil {
		t.Fatalf("bot relay: %v", err)
	}
	if err := module.RelayHandler()(context.Background(), relayEvent("   ")); err != nil {
		t.Fatalf("empty relay: %v", err)
	}
	if err := module.RelayHandler()(context.Background(), relayEvent("not configured")); err != nil {
		t.Fatalf("missing config relay: %v", err)
	}
	if len(sideEffects.Sent) != 0 || len(sideEffects.DeletedMessage) != 0 {
		t.Fatalf("side effects = sent %#v deleted %#v", sideEffects.Sent, sideEffects.DeletedMessage)
	}
}

func TestRelayHandlerDoesNotDeleteWhenSendFails(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.BoundAnnouncements["guild-1:channel-1"] = domain.BoundAnnouncementConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Tag:       "@here",
		Color:     "Random",
		Title:     "公告",
	}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Err = errors.New("send failed")
	module := NewRelayModule(repo, sideEffects)

	err := module.RelayHandler()(context.Background(), relayEvent("message"))
	if err == nil || !strings.Contains(err.Error(), "send failed") {
		t.Fatalf("expected send error, got %v", err)
	}
	if len(sideEffects.DeletedMessage) != 0 {
		t.Fatalf("original should not be deleted when send fails: %#v", sideEffects.DeletedMessage)
	}
}

func TestRelayHandlerRegisteredOnlyWhenModuleRelayEnabled(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewRelayModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("expected relay handler")
	}

	empty := events.NewDispatcher(nil)
	Module{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("unexpected relay handler for empty module")
	}
}

func relayEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		MessageID: "message-1",
		UserID:    "user-1",
		UserTag:   "Yoru#1234",
		AvatarURL: "https://cdn.discordapp.com/avatars/user-1/avatar.png",
		Content:   content,
	}
}
