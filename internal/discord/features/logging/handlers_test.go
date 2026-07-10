package logging

import (
	"context"
	"strings"
	"testing"

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
	module := NewModule(&fakemongo.LoggingConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := loggingSlash("channel-1")

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
}

func TestLoggingConfigSelectSavesConfigAndUpdatesMessage(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(loggingConfigCustomID("channel-9"))
	interaction.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "logging", Action: "configure"}
	interaction.Values = []string{"訊息更新", "用戶語音更新"}

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
	if len(usage.Events) != 1 || usage.Events[0].Feature != "logging" || usage.Events[0].CommandName != LoggingConfigCommandName {
		t.Fatalf("usage = %#v", usage.Events)
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
