package onboarding

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

func TestAccountAgeHandlerSetsHoursWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAccountAgeConfigRepository()
	usage := &fakeusage.Tracker{}
	module := NewAccountAgeModule(repo, usage)
	interaction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "小時數", map[string]string{"小時數": "24"})
	interaction.Actor.PermissionBits = permissionKickMembersBit
	responder := fakediscord.NewResponder()

	if err := module.AccountAgeHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("responses = defers:%#v edits:%#v", responder.Defers, responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614>群組防護系統" {
		t.Fatalf("title = %q", embed.Title)
	}
	if embed.Description != "已為您設定必須創建帳號1天才能加入伺服器" || embed.Color != accountAgeSuccessColor {
		t.Fatalf("embed = %#v", embed)
	}
	if got := repo.Configs["guild-1"]; got.RequiredSeconds != 86400 {
		t.Fatalf("saved = %#v", got)
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "account-age-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestAccountAgeHandlerRejectsPermissionAndInvalidHours(t *testing.T) {
	module := NewAccountAgeModule(fakemongo.NewAccountAgeConfigRepository(), nil)
	interaction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "小時數", map[string]string{"小時數": "24"})
	responder := fakediscord.NewResponder()

	if err := module.AccountAgeHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`踢出用戶`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionKickMembersBit
	interaction.Options["小時數"] = "0"
	responder = fakediscord.NewResponder()
	if err := module.AccountAgeHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("invalid handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "不可為負數或0!!!") {
		t.Fatalf("invalid response = %#v", responder.Edits)
	}
}

func TestAccountAgeHandlerRequiresHoursBeforeLogChannel(t *testing.T) {
	module := NewAccountAgeModule(fakemongo.NewAccountAgeConfigRepository(), nil)
	interaction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "被踢出資訊頻道", map[string]string{"頻道": "channel-1"})
	interaction.Actor.PermissionBits = permissionKickMembersBit
	responder := fakediscord.NewResponder()

	if err := module.AccountAgeHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你必須先設定`/帳號需創建時數 小時數`") {
		t.Fatalf("response = %#v", responder.Edits)
	}
}

func TestAccountAgeHandlerSetsAndDeletesLogChannel(t *testing.T) {
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	module := NewAccountAgeModule(repo, nil)
	interaction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "被踢出資訊頻道", map[string]string{"頻道": "channel-1"})
	interaction.Actor.PermissionBits = permissionKickMembersBit
	responder := fakediscord.NewResponder()

	if err := module.AccountAgeHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set channel: %v", err)
	}
	if repo.Configs["guild-1"].ChannelID != "channel-1" {
		t.Fatalf("config = %#v", repo.Configs["guild-1"])
	}
	if responder.Edits[0].Embeds[0].Description != "已為您設定當未達創建時數時會在:\n<#channel-1>發送使用者資運" {
		t.Fatalf("description = %q", responder.Edits[0].Embeds[0].Description)
	}

	deleteInteraction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "被踢出資訊頻道刪除", nil)
	deleteInteraction.Actor.PermissionBits = permissionKickMembersBit
	responder = fakediscord.NewResponder()
	if err := module.AccountAgeHandler()(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("delete channel: %v", err)
	}
	if repo.Configs["guild-1"].ChannelID != "" {
		t.Fatalf("config = %#v", repo.Configs["guild-1"])
	}
	if responder.Edits[0].Embeds[0].Description != "已刪除被踢出資訊頻道" {
		t.Fatalf("delete response = %#v", responder.Edits)
	}
}

func TestAccountAgeModuleRegistersRoute(t *testing.T) {
	module := NewAccountAgeModule(fakemongo.NewAccountAgeConfigRepository(), nil)
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions(AccountAgeCommandName, "小時數", map[string]string{"小時數": "1"})
	interaction.Actor.PermissionBits = permissionKickMembersBit
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
