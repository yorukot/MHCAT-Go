package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSetHandlerCreatesJoinRoleConfigWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewJoinRoleConfigRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	module := NewModule(repo, roles)
	interaction := fakediscord.SlashInteractionWithOptions(JoinRoleSetCommandName, "", map[string]string{
		"身分組":      "role-1",
		"給人還是給機器人": domain.JoinRoleGiveBots,
	})
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
	if embed.Title != "🪂 加入身分組系統" || embed.Description != "<a:green_tick:994529015652163614> **成功創建加入給身分組!**\n**身分組:** <@role-1>!" {
		t.Fatalf("embed = %#v", embed)
	}
	saved := repo.Configs["guild-1/role-1"]
	if saved.GiveTo != domain.JoinRoleGiveBots {
		t.Fatalf("saved config = %#v", saved)
	}
	if responder.Edits[0].AllowedMentions == nil || responder.Edits[0].AllowedMentions.ParseRoles {
		t.Fatalf("mentions should be suppressed: %#v", responder.Edits[0].AllowedMentions)
	}
}

func TestSetHandlerRejectsMissingPermissionUnassignableAndDuplicate(t *testing.T) {
	repo := fakemongo.NewJoinRoleConfigRepository()
	roles := fakediscord.NewSideEffects()
	module := NewModule(repo, roles)
	interaction := fakediscord.SlashInteractionWithOptions(JoinRoleSetCommandName, "", map[string]string{"身分組": "role-1"})
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("unassignable handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "我沒有權限為大家增加這個身分組") {
		t.Fatalf("unassignable response = %#v", responder.Edits)
	}

	roles.AssignableRoles["guild-1/role-1"] = true
	responder = fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("first create: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("duplicate create: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "這個身分組已經被註冊了") {
		t.Fatalf("duplicate response = %#v", responder.Edits)
	}
}

func TestDeleteHandlerDeletesJoinRoleConfigWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewJoinRoleConfigRepository()
	repo.Configs["guild-1/role-1"] = domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "role-1"}
	module := NewModule(repo, nil)
	interaction := fakediscord.SlashInteractionWithOptions(JoinRoleDeleteCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1/role-1"]; ok {
		t.Fatal("config was not deleted")
	}
	if got := responder.Edits[0].Embeds[0].Description; got != "<:trashbin:986308183674990592>**成功刪除:**\n身分組: <@role-1>!" {
		t.Fatalf("description = %q", got)
	}
}

func TestDeleteHandlerMissingConfig(t *testing.T) {
	module := NewModule(fakemongo.NewJoinRoleConfigRepository(), nil)
	interaction := fakediscord.SlashInteractionWithOptions(JoinRoleDeleteCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "找不到這個身份組!") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}

func TestJoinMessageDashboardHandlerUsesLegacyRedirectUI(t *testing.T) {
	module := NewMessageModule(fakemongo.NewLeaveMessageConfigRepository())
	interaction := fakediscord.SlashInteractionWithOptions(JoinMessageSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()

	if err := module.JoinMessageDashboardHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	reply := responder.Replies[0]
	if reply.Embeds[0].Title != "<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定" {
		t.Fatalf("embed = %#v", reply.Embeds[0])
	}
	button := reply.Components[0].Components[0]
	if button.Label != "點我前往儀錶板設定!" || button.Emoji != "<a:arrow:986268851786375218>" || button.URL != "https://mhcat.yorukot.meguilds/guild-1/welcome" {
		t.Fatalf("button = %#v", button)
	}
}

func TestLeaveMessagePromptHandlerShowsLegacyModal(t *testing.T) {
	repo := fakemongo.NewLeaveMessageConfigRepository()
	repo.Configs["guild-1"] = domain.LeaveMessageConfig{
		GuildID:        "guild-1",
		ChannelID:      "old-channel",
		Color:          "#df1f2f",
		Title:          "Bye",
		MessageContent: "Goodbye {MEMBERNAME}",
	}
	module := NewMessageModule(repo)
	interaction := fakediscord.SlashInteractionWithOptions(LeaveMessageSetCommandName, "", map[string]string{"頻道": "channel-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.LeaveMessagePromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	modal := responder.Modals[0]
	if modal.CustomID != "nal" || modal.Title != "退出訊息設置!" {
		t.Fatalf("modal = %#v", modal)
	}
	inputs := flattenModalInputs(modal)
	if len(inputs) != 3 {
		t.Fatalf("inputs = %#v", inputs)
	}
	if inputs[0].CustomID != "leave_msgcolor" || inputs[0].Label != "請輸入你的加入訊息要甚麼顏色(要隨機顏色可輸入:Random)" || inputs[0].Value != "#df1f2f" {
		t.Fatalf("color input = %#v", inputs[0])
	}
	if inputs[1].CustomID != "leave_msgtitle" || inputs[1].Value != "Bye" {
		t.Fatalf("title input = %#v", inputs[1])
	}
	if inputs[2].CustomID != "leave_msgcontent" || inputs[2].Value != "Goodbye {MEMBERNAME}" {
		t.Fatalf("content input = %#v", inputs[2])
	}
	if repo.Configs["guild-1"].ChannelID != "channel-1" {
		t.Fatalf("channel was not updated: %#v", repo.Configs["guild-1"])
	}
}

func TestLeaveMessagePromptHandlerPermissionDenied(t *testing.T) {
	module := NewMessageModule(fakemongo.NewLeaveMessageConfigRepository())
	interaction := fakediscord.SlashInteractionWithOptions(LeaveMessageSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()
	if err := module.LeaveMessagePromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
}

func TestLeaveMessageModalHandlerSavesAndShowsLegacyPreview(t *testing.T) {
	repo := fakemongo.NewLeaveMessageConfigRepository()
	repo.Configs["guild-1"] = domain.LeaveMessageConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	module := NewMessageModule(repo)
	interaction := leaveMessageModalInteraction("#df1f2f", "Bye", "Goodbye {MEMBERNAME}")
	interaction.Actor.AvatarURL = "https://cdn.example/avatar.png"
	responder := fakediscord.NewResponder()

	if err := module.LeaveMessageModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	if edit.Content != "下面為預覽，想修改嗎?再次輸入指令即可修改((MEMBERNAME)在到時候會變正常喔)" {
		t.Fatalf("content = %q", edit.Content)
	}
	if edit.Embeds[0].Title != "Bye" || edit.Embeds[0].Description != "Goodbye {MEMBERNAME}" || edit.Embeds[0].Color != 0xDF1F2F {
		t.Fatalf("embed = %#v", edit.Embeds[0])
	}
	if repo.Configs["guild-1"].MessageContent != "Goodbye {MEMBERNAME}" || repo.Configs["guild-1"].Title != "Bye" || repo.Configs["guild-1"].Color != "#df1f2f" {
		t.Fatalf("saved = %#v", repo.Configs["guild-1"])
	}
}

func TestLeaveMessageModalHandlerRejectsInvalidColor(t *testing.T) {
	repo := fakemongo.NewLeaveMessageConfigRepository()
	repo.Configs["guild-1"] = domain.LeaveMessageConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	module := NewMessageModule(repo)
	responder := fakediscord.NewResponder()
	if err := module.LeaveMessageModalHandler()(context.Background(), leaveMessageModalInteraction("not-a-color", "Bye", "Goodbye"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "你傳送的並不是顏色(色碼)" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if repo.Configs["guild-1"].MessageContent != "" {
		t.Fatalf("config should not save invalid color: %#v", repo.Configs["guild-1"])
	}
}

func leaveMessageModalInteraction(color string, title string, content string) interactions.Interaction {
	return interactions.Interaction{
		Type:     interactions.TypeModal,
		CustomID: "nal",
		ModalKey: interactions.ModalKey{
			Version: "legacy",
			Feature: "welcome",
			Action:  "leave_submit",
		},
		Actor: interactions.Actor{UserID: "user-1", GuildID: "guild-1"},
		ModalFields: []customid.ModalField{
			{CustomID: "leave_msgcolor", Value: color},
			{CustomID: "leave_msgtitle", Value: title},
			{CustomID: "leave_msgcontent", Value: content},
		},
	}
}

func flattenModalInputs(modal responses.Modal) []responses.TextInput {
	var inputs []responses.TextInput
	for _, row := range modal.Rows {
		inputs = append(inputs, row.Inputs...)
	}
	return inputs
}
