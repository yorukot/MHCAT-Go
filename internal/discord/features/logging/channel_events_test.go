package logging

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestChannelTopicLoggingSendsLegacyEmbed(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", ChannelUpdate: true},
	}}
	discord := fakediscord.NewSideEffects()
	discord.AuditEntries = []ports.AuditLogEntry{{
		UserID:    "moderator-1",
		Username:  "Mio",
		AvatarURL: "https://example.test/mod.png",
		TargetID:  "channel-1",
		Action:    loggingChannelUpdateAuditAction,
	}}
	module := NewChannelEventModule(repo, discord, discord)

	err := module.ChannelUpdateHandler()(context.Background(), events.Event{
		Type:         events.TypeChannelUpdate,
		GuildID:      "guild-1",
		ChannelID:    "channel-1",
		BotAvatarURL: "https://example.test/bot.png",
		ChannelUpdate: &events.ChannelUpdate{
			ChannelID:     "channel-1",
			GuildID:       "guild-1",
			HasOldChannel: true,
			OldTopic:      "old topic",
			NewTopic:      "new topic",
		},
	})
	if err != nil {
		t.Fatalf("channel topic log: %v", err)
	}
	if len(discord.Sent) != 1 || discord.Sent[0].ChannelID != "log-channel" {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Mio | 頻道主題更新" || embed.AuthorIconURL != "https://example.test/mod.png" || embed.Color != 0xFF8040 {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "頻道主題編輯者: <@moderator-1>") || !strings.Contains(embed.Description, "頻道: <#channel-1>") {
		t.Fatalf("description = %q", embed.Description)
	}
	if len(embed.Fields) != 2 || !strings.Contains(embed.Fields[0].Value, "old topic") || !strings.Contains(embed.Fields[1].Value, "new topic") {
		t.Fatalf("fields = %#v", embed.Fields)
	}
	if discord.Sent[0].Message.AllowedMentions.ParseUsers || discord.Sent[0].Message.AllowedMentions.ParseRoles || discord.Sent[0].Message.AllowedMentions.ParseEveryone {
		t.Fatalf("allowed mentions = %#v", discord.Sent[0].Message.AllowedMentions)
	}
}

func TestChannelPermissionLoggingSendsChangedOverwriteEmbeds(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", ChannelUpdate: true},
	}}
	discord := fakediscord.NewSideEffects()
	discord.AuditEntries = []ports.AuditLogEntry{{
		UserID:   "moderator-1",
		Username: "Mio",
		TargetID: "channel-1",
		Action:   loggingChannelUpdateAuditAction,
	}}
	module := NewChannelEventModule(repo, discord, discord)

	err := module.ChannelUpdateHandler()(context.Background(), events.Event{
		Type:      events.TypeChannelUpdate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		ChannelUpdate: &events.ChannelUpdate{
			HasOldChannel: true,
			OldPermissionOverwrites: []events.PermissionOverwrite{{
				ID:    "role-1",
				Type:  0,
				Allow: loggingPermissionSendMessages,
				Deny:  loggingPermissionConnect,
			}},
			NewPermissionOverwrites: []events.PermissionOverwrite{{
				ID:    "role-1",
				Type:  0,
				Allow: loggingPermissionSendMessages | loggingPermissionManageMessages,
			}},
		},
	})
	if err != nil {
		t.Fatalf("channel permission log: %v", err)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Mio | 頻道權限更新" || embed.Color != 0xFF5809 || !strings.Contains(embed.Description, "頻道權限編輯者: <@moderator-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if len(embed.Fields) != 1 {
		t.Fatalf("fields = %#v", embed.Fields)
	}
	value := embed.Fields[0].Value
	for _, want := range []string{
		"<:icons_text1:1000814305068986590><@&role-1>",
		"<:YellowSmallDot:1023970607429328946> Connect",
		"<:check:1085240252978966548> Manage Messages",
	} {
		if !strings.Contains(value, want) {
			t.Fatalf("field value %q missing %q", value, want)
		}
	}
}

func TestChannelLoggingSkipsDisabledUncachedAndUnchanged(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", ChannelUpdate: true},
		"guild-2": {GuildID: "guild-2", ChannelID: "log-channel"},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewChannelEventModule(repo, discord, discord)

	cases := []events.Event{
		{Type: events.TypeChannelUpdate, GuildID: "guild-1", ChannelID: "channel-1", ChannelUpdate: &events.ChannelUpdate{}},
		{Type: events.TypeChannelUpdate, GuildID: "guild-2", ChannelID: "channel-1", ChannelUpdate: &events.ChannelUpdate{HasOldChannel: true, OldTopic: "old", NewTopic: "new"}},
		{Type: events.TypeChannelUpdate, GuildID: "guild-1", ChannelID: "channel-1", ChannelUpdate: &events.ChannelUpdate{HasOldChannel: true, OldTopic: "same", NewTopic: "same"}},
	}
	for _, event := range cases {
		if err := module.ChannelUpdateHandler()(context.Background(), event); err != nil {
			t.Fatalf("channel handler: %v", err)
		}
	}
	if len(discord.Sent) != 0 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
}

func TestChannelEventModuleRegistersRoutes(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewChannelEventModule(&fakemongo.LoggingConfigRepository{}, fakediscord.NewSideEffects(), nil).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeChannelUpdate) {
		t.Fatal("expected channel logging handler")
	}
}
