package xp

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceXPEventMarksJoinMoveAndLeave(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo)

	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-1", "")); err != nil {
		t.Fatalf("join: %v", err)
	}
	profile := repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionJoined || profile.XP != 0 || profile.Level != 0 {
		t.Fatalf("joined profile = %#v", profile)
	}

	profile.XP = 75
	profile.Level = 2
	repo.VoiceProfiles["guild-1/user-1"] = profile
	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-2", "voice-1")); err != nil {
		t.Fatalf("move: %v", err)
	}
	profile = repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionJoined || profile.XP != 75 || profile.Level != 2 {
		t.Fatalf("moved profile = %#v", profile)
	}

	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("", "voice-2")); err != nil {
		t.Fatalf("leave: %v", err)
	}
	profile = repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionLeft || profile.XP != 75 || profile.Level != 2 {
		t.Fatalf("left profile = %#v", profile)
	}
}

func TestVoiceXPEventIgnoresBotSameChannelAndMissingPayload(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo)
	for _, event := range []events.Event{
		{Type: events.TypeMessageCreate},
		func() events.Event {
			event := voiceXPEvent("voice-1", "")
			event.IsBot = true
			return event
		}(),
		voiceXPEvent("voice-1", "voice-1"),
		{Type: events.TypeVoiceState, GuildID: "guild-1", VoiceState: &events.VoiceState{ChannelID: "voice-1"}},
	} {
		if err := module.VoiceStateHandler()(context.Background(), event); err != nil {
			t.Fatalf("ignored event returned error: %v", err)
		}
	}
	if len(repo.VoiceProfiles) != 0 {
		t.Fatalf("unexpected profiles = %#v", repo.VoiceProfiles)
	}
}

func TestVoiceXPEventRegisteredOnlyWithRepository(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewVoiceEventModule(fakemongo.NewXPAdminRepository()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeVoiceState) {
		t.Fatal("expected voice XP event handler")
	}

	empty := events.NewDispatcher(nil)
	VoiceEventModule{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeVoiceState) {
		t.Fatal("unexpected voice XP event handler")
	}
}

func voiceXPEvent(channelID string, beforeChannelID string) events.Event {
	return events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		UserID:  "user-1",
		VoiceState: &events.VoiceState{
			GuildID:       "guild-1",
			UserID:        "user-1",
			ChannelID:     channelID,
			BeforeChannel: beforeChannelID,
		},
	}
}
