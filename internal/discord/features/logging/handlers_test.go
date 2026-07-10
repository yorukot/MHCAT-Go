package logging

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestLoggingConfigPromptRequiresManageMessages(t *testing.T) {
	module := NewModule(&fakemongo.LoggingConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(LoggingConfigCommandName, "", map[string]string{"channel": "channel-1"})

	if err := module.ConfigPromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") || responder.Edits[0].Embeds[0].Color != 0xED4245 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestLoggingConfigPromptRendersLegacySelect(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_000)
	module := NewModuleWithClock(&fakemongo.LoggingConfigRepository{}, nil, loggingFixedClock{now: now})
	responder := fakediscord.NewResponder()
	interaction := loggingSlash("channel-1")
	interaction.BotAvatarURL = "https://example.test/bot.png"

	if err := module.ConfigPromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	message := responder.Edits[0]
	embed := message.Embeds[0]
	if embed.Title != "<:logfile:985948561625710663> 日誌系統" || embed.Color != loggingEmbedColor || !strings.Contains(embed.Description, "目前的選擇:") {
		t.Fatalf("embed = %#v", embed)
	}
	if embed.Footer == nil || embed.Footer.Text != loggingFooterText || embed.Footer.IconURL != "https://example.test/bot.webp" {
		t.Fatalf("footer = %#v", embed.Footer)
	}
	if len(message.Components) != 1 || len(message.Components[0].Components) != 1 {
		t.Fatalf("components = %#v", message.Components)
	}
	selectMenu := message.Components[0].Components[0]
	if selectMenu.Placeholder != "請選擇您需要的日誌" || selectMenu.MinValues != 1 || selectMenu.MaxValues != 4 || len(selectMenu.Options) != 4 {
		t.Fatalf("select = %#v", selectMenu)
	}
	if !strings.HasPrefix(selectMenu.CustomID, "mhcat:v1:logging:configure:") {
		t.Fatalf("custom id = %q", selectMenu.CustomID)
	}
	channelID, ownerID, expiresAt, ok := loggingConfigPayload(selectMenu.CustomID)
	if !ok || channelID != "channel-1" || ownerID != interaction.Actor.UserID || !expiresAt.Equal(now.Add(loggingConfigCollectorTTL)) {
		t.Fatalf("custom id payload = channel:%q owner:%q expiry:%v ok:%v", channelID, ownerID, expiresAt, ok)
	}
}

func TestLoggingConfigSelectSavesConfigAndUpdatesMessage(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{}
	usage := &fakeusage.Tracker{}
	now := time.UnixMilli(1_700_000_000_000)
	module := NewModuleWithClock(repo, usage, loggingFixedClock{now: now})
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(loggingConfigCustomID("channel-9", "user-1", now.Add(loggingConfigCollectorTTL)))
	interaction.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "logging", Action: "configure"}
	interaction.Values = []string{"訊息更新", "用戶語音更新"}
	interaction.BotAvatarURL = "https://example.test/bot.png"

	if err := module.ConfigSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.GuildID != "guild-1" || saved.ChannelID != "channel-9" || !saved.MessageUpdate || saved.MessageDelete || saved.ChannelUpdate || !saved.MemberVoiceUpdate {
		t.Fatalf("saved = %#v", saved)
	}
	if responder.DeferredUpdates != 1 {
		t.Fatalf("deferred updates = %d", responder.DeferredUpdates)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "`訊息更新`,`用戶語音更新`") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if footer := responder.Edits[0].Embeds[0].Footer; footer == nil || footer.IconURL != "https://example.test/bot.webp" {
		t.Fatalf("footer = %#v", footer)
	}
	for _, option := range responder.Edits[0].Components[0].Components[0].Options {
		if option.Default {
			t.Fatalf("legacy update should not persist selected defaults: %#v", option)
		}
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "logging" || usage.Events[0].CommandName != LoggingConfigCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestLoggingConfigSelectRejectsAnotherUser(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_000)
	repo := &fakemongo.LoggingConfigRepository{}
	module := NewModuleWithClock(repo, nil, loggingFixedClock{now: now})
	interaction := fakediscord.ComponentInteractionFromID(loggingConfigCustomID("channel-9", "owner-1", now.Add(loggingConfigCollectorTTL)))
	interaction.Actor.UserID = "other-1"
	interaction.Values = []string{"訊息更新"}
	responder := fakediscord.NewResponder()

	if err := module.ConfigSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Saved) != 0 {
		t.Fatalf("saved = %#v", repo.Saved)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "不能操作") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestLoggingConfigSelectExpiresAtLegacyDeadline(t *testing.T) {
	deadline := time.UnixMilli(1_700_000_600_000)
	repo := &fakemongo.LoggingConfigRepository{}
	module := NewModuleWithClock(repo, nil, loggingFixedClock{now: deadline})
	interaction := fakediscord.ComponentInteractionFromID(loggingConfigCustomID("channel-9", "user-1", deadline))
	interaction.Values = []string{"訊息更新"}
	responder := fakediscord.NewResponder()

	if err := module.ConfigSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Saved) != 0 {
		t.Fatalf("saved = %#v", repo.Saved)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "重新執行") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestLoggingConfigCustomIDFitsDiscordLimit(t *testing.T) {
	deadline := time.UnixMilli(9_999_999_999_999)
	customID := loggingConfigCustomID("1234567890123456789", "9876543210987654321", deadline)
	if len(customID) > customid.MaxCustomIDLength {
		t.Fatalf("custom id length = %d: %q", len(customID), customID)
	}
	channelID, ownerID, expiresAt, ok := loggingConfigPayload(customID)
	if !ok || channelID != "1234567890123456789" || ownerID != "9876543210987654321" || !expiresAt.Equal(deadline) {
		t.Fatalf("payload = channel:%q owner:%q expiry:%v ok:%v", channelID, ownerID, expiresAt, ok)
	}
}

func TestLegacyLoggingSelectCannotSaveWithoutChannelPayload(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("loggin_create")
	interaction.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "logging", Action: "configure_select", Legacy: true}
	interaction.Values = []string{"訊息更新"}

	if err := module.LegacyConfigSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Saved) != 0 {
		t.Fatalf("legacy handler saved config = %#v", repo.Saved)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "重新執行") {
		t.Fatalf("updates = %#v", responder.Updates)
	}
}

func loggingSlash(channelID string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(LoggingConfigCommandName, "", map[string]string{"channel": channelID})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

type loggingFixedClock struct {
	now time.Time
}

func (c loggingFixedClock) Now() time.Time {
	return c.now
}
