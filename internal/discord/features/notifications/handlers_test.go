package notifications

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestListHandlerRequiresManageMessagesWithoutDeferring(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AutoNotificationListCommandName)

	if err := module.ListHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 0 {
		t.Fatalf("list should not defer, defers = %#v", responder.Defers)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "訊息管理") || !strings.Contains(responder.Replies[0].Embeds[0].Description, "D43zPrZU5Fw") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestListHandlerRendersLegacyListAndCleansDrafts(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{
		{GuildID: "guild-1", ID: "1700000000000", Cron: "*/30 * * * *", ChannelID: "channel-1"},
		{GuildID: "guild-1", ID: "draft", Pending: true, ChannelID: "channel-2"},
	}
	usage := &fakeusage.Tracker{}
	module := NewModuleWithColor(repo, &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試群"}}, usage, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()

	if err := module.ListHandler()(context.Background(), autoNotificationListSlash(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	embed := responder.Replies[0].Embeds[0]
	if embed.Title != "<:list:992002476360343602> 以下是測試群的所有自動通知id" || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "輸入`/自動通知刪除 id`") || !strings.Contains(embed.Description, "id:`1700000000000`") || strings.Contains(embed.Description, "draft") {
		t.Fatalf("description = %q", embed.Description)
	}
	if len(repo.PendingCleanupGuilds) != 1 {
		t.Fatalf("pending cleanup = %#v", repo.PendingCleanupGuilds)
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "auto-notification-config" || usage.Events[0].CommandName != AutoNotificationListCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerDeletesAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "schedule-1", Cron: "0 9 * * 1", ChannelID: "channel-1"}}
	module := NewModule(repo, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.DeleteHandler()(context.Background(), autoNotificationDeleteSlash("schedule-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Deleted) != 1 || repo.Deleted[0].ID != "schedule-1" {
		t.Fatalf("deleted = %#v", repo.Deleted)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614>自動通知系統" || embed.Description != "<:trashbin:995991389043163257>成功刪除該自動通知" || embed.Color != autoNotificationGreenColor {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestDeleteHandlerMissingUsesLegacyTutorialError(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.DeleteHandler()(context.Background(), autoNotificationDeleteSlash("missing"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "找不到這個id的自動通知") || !strings.Contains(responder.Edits[0].Embeds[0].Description, "D43zPrZU5Fw") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func autoNotificationListSlash() interactions.Interaction {
	interaction := fakediscord.SlashInteraction(AutoNotificationListCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func autoNotificationDeleteSlash(id string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(AutoNotificationDeleteCommandName, "", map[string]string{optionID: id})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}
