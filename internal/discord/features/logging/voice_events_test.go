package logging

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceJoinLoggingSendsLegacyEmbed(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MemberVoiceUpdate: true},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewVoiceEventModule(repo, discord)

	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:         events.TypeVoiceState,
		GuildID:      "guild-1",
		UserID:       "user-1",
		Username:     "Yoru",
		AvatarURL:    "https://example.test/user.png",
		BotAvatarURL: "https://example.test/bot.png",
		VoiceState: &events.VoiceState{
			GuildID:     "guild-1",
			UserID:      "user-1",
			ChannelID:   "voice-1",
			ChannelName: "General",
		},
	})
	if err != nil {
		t.Fatalf("voice join log: %v", err)
	}
	if len(discord.Sent) != 1 || discord.Sent[0].ChannelID != "log-channel" {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Yoru | 使用者加入語音頻道" || embed.AuthorIconURL != "https://example.test/user.png" || embed.Color != 0xF235FA {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "使用者: <@user-1>") || !strings.Contains(embed.Description, "頻道: <#voice-1>") {
		t.Fatalf("description = %q", embed.Description)
	}
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "**<:joines:1086217186256900098> 加入頻道:**" || embed.Fields[0].Value != "<#voice-1>(`General`)" {
		t.Fatalf("fields = %#v", embed.Fields)
	}
	if embed.FooterText != loggingFooterText || embed.FooterIconURL != "https://example.test/bot.webp" || embed.Timestamp.IsZero() {
		t.Fatalf("footer/timestamp = %#v", embed)
	}
	if discord.Sent[0].Message.AllowedMentions.ParseUsers || discord.Sent[0].Message.AllowedMentions.ParseRoles || discord.Sent[0].Message.AllowedMentions.ParseEveryone {
		t.Fatalf("allowed mentions = %#v", discord.Sent[0].Message.AllowedMentions)
	}
}

func TestVoiceLeaveLoggingSendsLegacyEmbed(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MemberVoiceUpdate: true},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewVoiceEventModule(repo, discord)

	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:      events.TypeVoiceState,
		GuildID:   "guild-1",
		UserID:    "user-1",
		Username:  "Yoru",
		AvatarURL: "https://example.test/user.png",
		VoiceState: &events.VoiceState{
			GuildID:           "guild-1",
			UserID:            "user-1",
			BeforeChannel:     "voice-1",
			BeforeChannelName: "General",
		},
	})
	if err != nil {
		t.Fatalf("voice leave log: %v", err)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Yoru | 使用者退出語音頻道" || embed.Color != 0xFA359A {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "頻道: <#voice-1>") || len(embed.Fields) != 1 || embed.Fields[0].Name != "**<:leaves:1086219523264356513> 退出頻道:**" || embed.Fields[0].Value != "<#voice-1>(`General`)" {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestVoiceLoggingSkipsDisabledMovesAndNonChannelChanges(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MemberVoiceUpdate: true},
		"guild-2": {GuildID: "guild-2", ChannelID: "log-channel"},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewVoiceEventModule(repo, discord)

	for _, event := range []events.Event{
		{Type: events.TypeVoiceState, GuildID: "guild-1", UserID: "user-1", VoiceState: &events.VoiceState{GuildID: "guild-1", UserID: "user-1", BeforeChannel: "voice-1", ChannelID: "voice-2"}},
		{Type: events.TypeVoiceState, GuildID: "guild-1", UserID: "user-1", VoiceState: &events.VoiceState{GuildID: "guild-1", UserID: "user-1", BeforeChannel: "voice-1", ChannelID: "voice-1"}},
		{Type: events.TypeVoiceState, GuildID: "guild-2", UserID: "user-1", VoiceState: &events.VoiceState{GuildID: "guild-2", UserID: "user-1", ChannelID: "voice-1"}},
		{Type: events.TypeVoiceState, GuildID: "guild-1", VoiceState: &events.VoiceState{GuildID: "guild-1", ChannelID: "voice-1"}},
	} {
		if err := module.VoiceStateHandler()(context.Background(), event); err != nil {
			t.Fatalf("voice handler: %v", err)
		}
	}
	if len(discord.Sent) != 0 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
}

func TestVoiceEventModuleRegistersRoute(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewVoiceEventModule(&fakemongo.LoggingConfigRepository{}, fakediscord.NewSideEffects()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeVoiceState) {
		t.Fatal("expected voice logging handler")
	}
}
