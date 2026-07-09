package xp

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSetHandlerRendersLegacySuccessAndPreview(t *testing.T) {
	repo := fakemongo.NewTextXPConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, sideEffects, usage)
	interaction := fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{
		"頻道": "channel-1",
		"訊息": "  {user} 升到了 {level}  ",
		"顏色": "#00ff00",
	})
	interaction.ChannelID = "invoke-channel"
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "聊天經驗系統" || embed.Description != "您的聊天經驗升等頻道成功創建\n您目前的升等通知頻道為 <#channel-1>" {
		t.Fatalf("embed = %#v", embed)
	}
	saved := repo.Configs["guild-1"]
	if saved.ChannelID != "channel-1" || saved.Color != "#00ff00" || saved.Message != "  {user} 升到了 {level}  " {
		t.Fatalf("saved config = %#v", saved)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "invoke-channel" {
		t.Fatalf("preview sends = %#v", sideEffects.Sent)
	}
	if !strings.Contains(sideEffects.Sent[0].Message.Content, "以下為你的訊息預覽:") || !strings.Contains(sideEffects.Sent[0].Message.Content, legacyLineEmoji+"我") {
		t.Fatalf("preview content = %q", sideEffects.Sent[0].Message.Content)
	}
	if sideEffects.Sent[0].Message.AllowedMentions.ParseUsers || sideEffects.Sent[0].Message.AllowedMentions.ParseRoles || sideEffects.Sent[0].Message.AllowedMentions.ParseEveryone {
		t.Fatalf("preview should suppress mentions: %#v", sideEffects.Sent[0].Message.AllowedMentions)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != TextXPSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestSetHandlerRejectsMissingPermissionAndInvalidColor(t *testing.T) {
	module := NewModule(fakemongo.NewTextXPConfigRepository(), nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.Options["顏色"] = "not-a-color"
	responder = fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你傳送的並不是顏色(色碼)") {
		t.Fatalf("color response = %#v", responder.Edits)
	}
}

func TestDeleteHandlerSuccessAndMissing(t *testing.T) {
	repo := fakemongo.NewTextXPConfigRepository()
	repo.Configs["guild-1"] = domain.TextXPConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, nil, usage)
	interaction := fakediscord.SlashInteraction(TextXPDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("config was not deleted")
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Description != "成功刪除!" {
		t.Fatalf("delete response = %#v", responder.Edits)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != TextXPDeleteCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}

	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler missing: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你本來就沒有對聊天經驗設定喔!") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}

func TestVoiceSetHandlerRendersLegacySuccessAndIgnoresBackground(t *testing.T) {
	repo := fakemongo.NewVoiceXPConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewVoiceModule(repo, sideEffects, usage)
	interaction := fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{
		"頻道": "voice-channel-1",
		"訊息": "  {user} 升到了 {level}  ",
		"顏色": "#00ff00",
		"背景": "https://example.invalid/background.png",
	})
	interaction.ChannelID = "invoke-channel"
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "語音經驗系統" || embed.Description != "您的語音經驗升等頻道成功創建\n您目前的升等通知頻道為 <#voice-channel-1>" {
		t.Fatalf("embed = %#v", embed)
	}
	saved := repo.Configs["guild-1"]
	if saved.ChannelID != "voice-channel-1" || saved.Color != "#00ff00" || saved.Message != "  {user} 升到了 {level}  " {
		t.Fatalf("saved config = %#v", saved)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "invoke-channel" {
		t.Fatalf("preview sends = %#v", sideEffects.Sent)
	}
	if !strings.Contains(sideEffects.Sent[0].Message.Content, "以下為你的訊息預覽:") || !strings.Contains(sideEffects.Sent[0].Message.Content, legacyLineEmoji+"我") {
		t.Fatalf("preview content = %q", sideEffects.Sent[0].Message.Content)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VoiceXPSetCommandName || usage.Events[0].Feature != "voice-xp-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestVoiceSetHandlerRejectsMissingPermissionAndInvalidColor(t *testing.T) {
	module := NewVoiceModule(fakemongo.NewVoiceXPConfigRepository(), nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.Options["顏色"] = "not-a-color"
	responder = fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你傳送的並不是顏色(色碼)") {
		t.Fatalf("color response = %#v", responder.Edits)
	}
}

func TestVoiceDeleteHandlerSuccessAndMissing(t *testing.T) {
	repo := fakemongo.NewVoiceXPConfigRepository()
	repo.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	usage := &fakeusage.Tracker{}
	module := NewVoiceModule(repo, nil, usage)
	interaction := fakediscord.SlashInteraction(VoiceXPDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("voice config was not deleted")
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "語音經驗系統" || responder.Edits[0].Embeds[0].Description != "成功刪除!" {
		t.Fatalf("voice delete response = %#v", responder.Edits)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VoiceXPDeleteCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}

	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler missing: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你本來就沒有對語音經驗設定喔!") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}
