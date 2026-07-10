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

func TestMessageUpdateLoggingSendsLegacyEmbed(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MessageUpdate: true},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewMessageEventModule(repo, discord, discord)

	err := module.MessageUpdateHandler()(context.Background(), events.Event{
		Type:          events.TypeMessageUpdate,
		GuildID:       "guild-1",
		ChannelID:     "source-channel",
		MessageID:     "message-1",
		UserID:        "user-1",
		Username:      "Yoru",
		AvatarURL:     "https://example.test/avatar.png",
		Content:       "new text",
		OldContent:    "old text",
		HasOldContent: true,
		BotAvatarURL:  "https://example.test/bot.png",
		Attachments:   []events.Attachment{{URL: "https://example.test/file.png"}},
	})
	if err != nil {
		t.Fatalf("message update log: %v", err)
	}
	if len(discord.Sent) != 1 || discord.Sent[0].ChannelID != "log-channel" {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Yoru | 訊息編輯" || embed.Color != 0x46A3FF || !strings.Contains(embed.Description, "訊息編輯者: <@user-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if len(embed.Fields) != 3 || !strings.Contains(embed.Fields[0].Value, "old text") || !strings.Contains(embed.Fields[1].Value, "new text") || embed.Fields[2].Value != "https://example.test/file.png" {
		t.Fatalf("fields = %#v", embed.Fields)
	}
	if discord.Sent[0].Message.AllowedMentions.ParseUsers || discord.Sent[0].Message.AllowedMentions.ParseRoles || discord.Sent[0].Message.AllowedMentions.ParseEveryone {
		t.Fatalf("allowed mentions = %#v", discord.Sent[0].Message.AllowedMentions)
	}
}

func TestMessageDeleteLoggingUsesAuditActorWhenAvailable(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MessageDelete: true},
	}}
	discord := fakediscord.NewSideEffects()
	discord.AuditEntries = []ports.AuditLogEntry{{
		UserID:    "moderator-1",
		TargetID:  "user-1",
		ChannelID: "source-channel",
		Action:    loggingMessageDeleteAuditAction,
	}}
	module := NewMessageEventModule(repo, discord, discord)

	err := module.MessageDeleteHandler()(context.Background(), events.Event{
		Type:        events.TypeMessageDelete,
		GuildID:     "guild-1",
		ChannelID:   "source-channel",
		MessageID:   "message-1",
		UserID:      "user-1",
		Username:    "Yoru",
		Content:     "deleted text",
		Attachments: []events.Attachment{{URL: "https://example.test/deleted.png"}},
	})
	if err != nil {
		t.Fatalf("message delete log: %v", err)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	embed := discord.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "Yoru | 訊息刪除" || embed.Color != 0x84C1FF || !strings.Contains(embed.Description, "訊息刪除者: <@moderator-1>") || !strings.Contains(embed.Description, "訊息發送者:<@user-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if len(embed.Fields) != 2 || !strings.Contains(embed.Fields[0].Value, "deleted text") || embed.Fields[1].Value != "https://example.test/deleted.png" {
		t.Fatalf("fields = %#v", embed.Fields)
	}
}

func TestMessageDeleteLoggingRequiresExactAuditTargetAndChannel(t *testing.T) {
	for _, entry := range []ports.AuditLogEntry{
		{UserID: "moderator-1", ChannelID: "source-channel", Action: loggingMessageDeleteAuditAction},
		{UserID: "moderator-1", TargetID: "user-1", Action: loggingMessageDeleteAuditAction},
		{UserID: "moderator-1", TargetID: "other-user", ChannelID: "source-channel", Action: loggingMessageDeleteAuditAction},
		{UserID: "moderator-1", TargetID: "user-1", ChannelID: "other-channel", Action: loggingMessageDeleteAuditAction},
	} {
		repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
			"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MessageDelete: true},
		}}
		discord := fakediscord.NewSideEffects()
		discord.AuditEntries = []ports.AuditLogEntry{entry}
		module := NewMessageEventModule(repo, discord, discord)

		err := module.MessageDeleteHandler()(context.Background(), events.Event{
			Type:      events.TypeMessageDelete,
			GuildID:   "guild-1",
			ChannelID: "source-channel",
			UserID:    "user-1",
			Username:  "Yoru",
		})
		if err != nil {
			t.Fatalf("message delete log: %v", err)
		}
		if len(discord.Sent) != 1 || !strings.Contains(discord.Sent[0].Message.Embeds[0].Description, "訊息刪除者: <@user-1>") {
			t.Fatalf("entry %#v sent = %#v", entry, discord.Sent)
		}
	}
}

func TestMessageLoggingSkipsDisabledBotAndUncachedUpdate(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{Configs: map[string]domain.LoggingConfig{
		"guild-1": {GuildID: "guild-1", ChannelID: "log-channel", MessageUpdate: true, MessageDelete: true},
	}}
	discord := fakediscord.NewSideEffects()
	module := NewMessageEventModule(repo, discord, discord)

	cases := []events.Event{
		{Type: events.TypeMessageUpdate, GuildID: "guild-1", ChannelID: "source-channel", UserID: "user-1", Content: "new", OldContent: "old"},
		{Type: events.TypeMessageUpdate, GuildID: "guild-1", ChannelID: "source-channel", UserID: "user-1", Content: "same", OldContent: "same", HasOldContent: true},
		{Type: events.TypeMessageDelete, GuildID: "guild-1", ChannelID: "source-channel", UserID: "bot-1", IsBot: true},
	}
	for _, event := range cases {
		switch event.Type {
		case events.TypeMessageUpdate:
			if err := module.MessageUpdateHandler()(context.Background(), event); err != nil {
				t.Fatalf("update handler: %v", err)
			}
		case events.TypeMessageDelete:
			if err := module.MessageDeleteHandler()(context.Background(), event); err != nil {
				t.Fatalf("delete handler: %v", err)
			}
		}
	}
	if len(discord.Sent) != 0 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
}

func TestMessageEventModuleRegistersRoutes(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewMessageEventModule(&fakemongo.LoggingConfigRepository{}, fakediscord.NewSideEffects(), nil).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageUpdate) || !dispatcher.HasHandlers(events.TypeMessageDelete) {
		t.Fatal("expected message logging handlers")
	}
}

func TestLoggingCodeBlockPreservesLegacySpacing(t *testing.T) {
	if got := loggingCodeBlock(""); got != "``` ```" {
		t.Fatalf("empty code block = %q", got)
	}
	if got := loggingCodeBlock("content"); got != "```content ```" {
		t.Fatalf("content code block = %q", got)
	}
}
