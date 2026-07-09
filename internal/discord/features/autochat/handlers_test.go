package autochat

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSetHandlerRequiresManageMessages(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := autoChatSetSlash()
	interaction.Actor.PermissionBits = 0

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestSetHandlerSavesAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAutoChatConfigRepository()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), autoChatSetSlash(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || saved.ChannelID != "channel-1" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "自動聊天系統" || embed.Color != autoChatSuccessColor || !strings.Contains(embed.Description, "您的自動聊天頻道成功創建") || !strings.Contains(embed.Description, "<#channel-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatalf("allowed mentions not set: %#v", responder.Edits[0])
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "autochat-config" || usage.Events[0].CommandName != AutoChatSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerDeletesAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAutoChatConfigRepository()
	repo.Configs["guild-1"] = structConfig("guild-1", "channel-1")
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()

	interaction := fakediscord.SlashInteraction(AutoChatDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatalf("config was not deleted: %#v", repo.Configs)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Description != "您的自動聊天頻道成功刪除" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteHandlerMissingConfigUsesLegacyError(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AutoChatDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你沒有設定過，我不知道要刪除甚麼") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func autoChatSetSlash() interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(AutoChatSetCommandName, "", map[string]string{
		optionChannel: "channel-1",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func structConfig(guildID string, channelID string) domain.AutoChatConfig {
	return domain.AutoChatConfig{GuildID: guildID, ChannelID: channelID}
}
