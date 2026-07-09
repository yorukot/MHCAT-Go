package stats

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestQueryHandlerRendersLegacyStaticEmbed(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewModuleWithColor(usage, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsQueryCommandName)

	if err := module.QueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	reply := responder.Replies[0]
	if len(reply.Embeds) != 1 {
		t.Fatalf("embeds = %#v", reply.Embeds)
	}
	embed := reply.Embeds[0]
	if embed.Title != "統計系統查詢" || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []string{
		"我的統計系統是每**10分鐘更新一次**",
		"輸入 /統計系統創建",
		"用戶總數 (伺服器的總人數)",
		"文字頻道數量 (文字頻道總數)",
		"語音頻道數量 (語音頻道總數)",
	} {
		if !strings.Contains(embed.Description, want) {
			t.Fatalf("description missing %q: %q", want, embed.Description)
		}
	}
	if reply.AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsQueryCommandName || usage.Events[0].Feature != "stats-query" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestQueryHandlerDoesNotDeferLikeLegacy(t *testing.T) {
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	if err := module.QueryHandler()(context.Background(), fakediscord.SlashInteraction(StatsQueryCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 0 || len(responder.Edits) != 0 {
		t.Fatalf("unexpected defer/edit: defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestDeleteHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	module := NewDeleteModule(repo, nil)
	responder := fakediscord.NewResponder()

	if err := module.DeleteHandler()(context.Background(), fakediscord.SlashInteraction(StatsDeleteCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if _, ok := repo.Configs["guild-1"]; !ok {
		t.Fatal("permission failure should not delete stats config")
	}
}

func TestCreateHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	module := NewCreateModule(repo, discord, discord, nil, "bot-1")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Follow) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("followups = %#v", responder.Follow)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestCreateHandlerCreatesStatsWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 15
	discord.NonBotMembers = 12
	usage := &fakeusage.Tracker{}
	module := NewCreateModule(repo, discord, discord, usage, "bot-1")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("defers=%#v followups=%#v", responder.Defers, responder.Follow)
	}
	if title := responder.Follow[0].Embeds[0].Title; title != "<a:greentick:980496858445135893> | 成功創建!頻道(不要動到數字就沒問題)跟類別的名稱都能自行更改喔!" {
		t.Fatalf("title = %q", title)
	}
	if len(discord.Created) != 4 || discord.Created[1].Name != "總人數: 15" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	if saved := repo.Configs["guild-1"]; saved.MemberNumberName != "15" || saved.UserNumberName != "12" || saved.BotNumberName != "3" {
		t.Fatalf("saved = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsCreateCommandName || usage.Events[0].Feature != "stats-create" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestCreateHandlerExistingStatsRequiresOptionWithLegacyError(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	module := NewCreateModule(repo, discord, discord, nil, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Follow) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "由於你已經創建過了") {
		t.Fatalf("followups = %#v", responder.Follow)
	}
}

func TestCreateHandlerAddsOptionalStatsChannel(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: 4})
	discord.TextChannelCount = 9
	module := NewCreateModule(repo, discord, discord, nil, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
		statsOptionStat:        "文字頻道數量",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(discord.Created) != 1 || discord.Created[0].Name != "總文字頻道數: 9" || discord.Created[0].ParentID != "parent-1" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	if saved := repo.Configs["guild-1"]; saved.TextNumberName != "9" || saved.TextNumberID == "" {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestDeleteHandlerDeletesStatsConfigWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	usage := &fakeusage.Tracker{}
	module := NewDeleteModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if title := responder.Edits[0].Embeds[0].Title; title != "<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:parent-1)" {
		t.Fatalf("title = %q", title)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("stats config should be deleted")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsDeleteCommandName || usage.Events[0].Feature != "stats-delete" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerMissingStatsConfigUsesLegacyError(t *testing.T) {
	module := NewDeleteModule(fakemongo.NewStatsConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你還沒有創建過統計數據") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteHandlerUnknownErrorUsesSafeLegacyStyleError(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Err = ports.ErrCoinLimitExceeded
	module := NewDeleteModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
