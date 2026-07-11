package onboarding

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLeaveMessageDeliveryHandlerSendsLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewLeaveMessageConfigRepository()
	repo.Configs["guild-1"] = domain.LeaveMessageConfig{
		GuildID:        "guild-1",
		ChannelID:      "channel-1",
		Title:          "Bye (MEMBERNAME)",
		MessageContent: "ID: {ID}",
		Color:          "#df1f2f",
	}
	sideEffects := fakediscord.NewSideEffects()
	cacheEventDeliveryChannel(sideEffects, "guild-1", "channel-1")
	module := NewLeaveMessageDeliveryModule(repo, sideEffects, sideEffects)
	now := time.Date(2026, 7, 4, 2, 3, 4, 0, time.UTC)

	err := module.LeaveMessageDeliveryHandler()(context.Background(), events.Event{
		Type:      events.TypeMemberRemove,
		GuildID:   "guild-1",
		CreatedAt: now,
		Member: &events.Member{
			UserID:    "user-1",
			Username:  "Tester",
			UserTag:   "Tester#0001",
			AvatarURL: "https://cdn.example/avatar.png",
		},
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0]
	if sent.ChannelID != "channel-1" || len(sent.Message.Embeds) != 1 {
		t.Fatalf("sent = %#v", sent)
	}
	embed := sent.Message.Embeds[0]
	if embed.Title != "Bye (MEMBERNAME)" || embed.Description != "ID: user-1" {
		t.Fatalf("embed = %#v", embed)
	}
	if embed.ThumbnailURL != "https://cdn.example/avatar.png" || !embed.Timestamp.Equal(now) {
		t.Fatalf("embed shape = %#v", embed)
	}
}

func TestLeaveMessageDeliveryHandlerIgnoresOtherEventsAndIncompleteEvents(t *testing.T) {
	repo := fakemongo.NewLeaveMessageConfigRepository()
	repo.Configs["guild-1"] = domain.LeaveMessageConfig{
		GuildID:        "guild-1",
		ChannelID:      "channel-1",
		Title:          "Bye",
		MessageContent: "Bye",
		Color:          "#df1f2f",
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewLeaveMessageDeliveryModule(repo, sideEffects, sideEffects)
	for _, event := range []events.Event{
		{Type: events.TypeMemberAdd, GuildID: "guild-1", UserID: "user-1"},
		{Type: events.TypeMemberRemove, UserID: "user-1"},
		{Type: events.TypeMemberRemove, GuildID: "guild-1"},
	} {
		if err := module.LeaveMessageDeliveryHandler()(context.Background(), event); err != nil {
			t.Fatalf("handler: %v", err)
		}
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestLeaveMessageDeliveryEventRouteRegistration(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	sideEffects := fakediscord.NewSideEffects()
	NewLeaveMessageDeliveryModule(fakemongo.NewLeaveMessageConfigRepository(), sideEffects, sideEffects).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMemberRemove) {
		t.Fatal("expected member remove handler")
	}
	if dispatcher.HasHandlers(events.TypeMemberAdd) {
		t.Fatal("leave delivery module should not register member add handler")
	}
}
