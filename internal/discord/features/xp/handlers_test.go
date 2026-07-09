package xp

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

func TestDisabledProfileHandlersReturnLegacyRemovalMessage(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewDisabledProfileModule(usage)

	for _, tc := range []struct {
		name    string
		handler func() interactions.Handler
	}{
		{name: TextXPProfileCommandName, handler: module.TextHandler},
		{name: VoiceXPProfileCommandName, handler: module.VoiceHandler},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interaction := fakediscord.SlashInteraction(tc.name)
			responder := fakediscord.NewResponder()
			if err := tc.handler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Defers) != 1 {
				t.Fatalf("defers = %#v", responder.Defers)
			}
			if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
				t.Fatalf("edits = %#v", responder.Edits)
			}
			embed := responder.Edits[0].Embeds[0]
			if embed.Title != "<a:Discord_AnimatedNo:1015989839809757295> | "+disabledProfileMessage || embed.Color != textXPErrorColor {
				t.Fatalf("embed = %#v", embed)
			}
			if responder.Edits[0].AllowedMentions == nil {
				t.Fatalf("allowed mentions should be explicitly empty: %#v", responder.Edits[0])
			}
		})
	}

	if len(usage.Events) != 2 {
		t.Fatalf("usage = %#v", usage.Events)
	}
	if usage.Events[0].CommandName != TextXPProfileCommandName || usage.Events[0].Feature != "xp-profile-disabled" {
		t.Fatalf("text usage = %#v", usage.Events[0])
	}
	if usage.Events[1].CommandName != VoiceXPProfileCommandName || usage.Events[1].Feature != "xp-profile-disabled" {
		t.Fatalf("voice usage = %#v", usage.Events[1])
	}
}

func TestAdminHandlerAddsTextXPAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	usage := &fakeusage.Tracker{}
	module := NewAdminModule(repo, usage)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "聊天經驗改變", map[string]string{
		"使用者": "user-2",
		"經驗值": "150",
	})
	interaction.Actor.PermissionBits = permissionKickMembers
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	profile := repo.TextProfiles["guild-1/user-2"]
	if profile.Level != 1 || profile.XP != 50 {
		t.Fatalf("profile = %#v", profile)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:xp:990254386792005663> 經驗系統" || embed.Description != doneEmoji+"成功為:<@user-2>\n增加:`150`" {
		t.Fatalf("embed = %#v", embed)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != XPAdminCommandName || usage.Events[0].Feature != "xp-admin" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestAdminHandlerAddsVoiceXPWithTypedOptions(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", Level: 2, XP: 50}
	module := NewAdminModule(repo, nil)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "語音經驗改變", map[string]string{})
	interaction.Actor.PermissionBits = permissionKickMembers
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		"使用者": {Type: interactions.CommandOptionUser, String: "user-2"},
		"經驗值": {Type: interactions.CommandOptionInteger, Int: 500, String: "500"},
	}
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	profile := repo.VoiceProfiles["guild-1/user-2"]
	if profile.Level != 3 || profile.XP != 250 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestAdminHandlerRequiresKickMembers(t *testing.T) {
	module := NewAdminModule(fakemongo.NewXPAdminRepository(), nil)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "聊天經驗改變", map[string]string{
		"使用者": "user-2",
		"經驗值": "1",
	})
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "踢出用戶") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestRewardRoleAddDeleteAndQueryHandlers(t *testing.T) {
	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	usage := &fakeusage.Tracker{}
	module := NewRewardRoleModule(textRepo, voiceRepo, roles, usage)

	add := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "增加", map[string]string{
		"等級":     "12",
		"身分組":    "role-1",
		"是否自動刪除": "true",
	})
	add.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), add, responder); err != nil {
		t.Fatalf("add handler: %v", err)
	}
	if len(textRepo.Configs) != 1 || textRepo.Configs[0].Level != 12 || textRepo.Configs[0].RoleID != "role-1" || !textRepo.Configs[0].DeleteWhenNot {
		t.Fatalf("saved reward roles = %#v", textRepo.Configs)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != channelEmoji+"聊天經驗系統" || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功`增加`/`修改`該設定") {
		t.Fatalf("add response = %#v", responder.Edits)
	}

	query := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "設定查詢", nil)
	query.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), query, responder); err != nil {
		t.Fatalf("query handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds[0].Fields) != 1 {
		t.Fatalf("query response = %#v", responder.Edits)
	}
	field := responder.Edits[0].Embeds[0].Fields[0]
	if !strings.Contains(field.Value, "**等級:**`12`") || !strings.Contains(field.Value, "**身分組:**<@&role-1>") {
		t.Fatalf("query field = %#v", field)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatalf("allowed mentions should be explicit: %#v", responder.Edits[0])
	}

	del := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "刪除", map[string]string{"等級": "12", "身分組": "role-1"})
	del.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), del, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(textRepo.Configs) != 0 {
		t.Fatalf("reward role was not deleted: %#v", textRepo.Configs)
	}
	if len(usage.Events) != 3 || usage.Events[0].Feature != "text-xp-role-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestRewardRoleRejectsPermissionRoleAndMissingDelete(t *testing.T) {
	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	module := NewRewardRoleModule(textRepo, fakemongo.NewVoiceXPRewardRoleRepository(), fakediscord.NewSideEffects(), nil)

	interaction := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "增加", map[string]string{"等級": "1", "身分組": "role-1"})
	responder := fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("role handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "我沒有權限給大家這個身分組") {
		t.Fatalf("role response = %#v", responder.Edits)
	}

	deleteInteraction := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "刪除", map[string]string{"等級": "1", "身分組": "role-1"})
	deleteInteraction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("missing delete handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你沒有設定過這個選項!") {
		t.Fatalf("missing delete response = %#v", responder.Edits)
	}
}

func TestRewardRoleVoicePaginationUpdatesMessage(t *testing.T) {
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	for i := 0; i < 13; i++ {
		voiceRepo.Configs = append(voiceRepo.Configs, domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i + 1), RoleID: "role"})
	}
	module := NewRewardRoleModule(fakemongo.NewTextXPRewardRoleRepository(), voiceRepo, nil, nil)
	module.color = func() int { return 0x123456 }
	interaction := fakediscord.ComponentInteractionFromID("1voice_leave_role")
	responder := fakediscord.NewResponder()
	if err := module.VoicePageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("page handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Embeds[0].Fields) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	if responder.Updates[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("color = %#v", responder.Updates[0].Embeds[0])
	}
	if responder.Updates[0].Components[0].Components[0].CustomID != "0voice_leave_role" || !responder.Updates[0].Components[0].Components[1].Disabled {
		t.Fatalf("components = %#v", responder.Updates[0].Components)
	}
}
