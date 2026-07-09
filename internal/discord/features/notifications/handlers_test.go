package notifications

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSetupHandlerShowsLegacyModalAndCreatesPendingDraft(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, nil, usage)
	responder := fakediscord.NewResponder()
	interaction := autoNotificationSetupSlash("channel-1")
	interaction.CreatedAt = time.UnixMilli(1700000000000)

	if err := module.SetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	modal := responder.Modals[0]
	if modal.CustomID != "1700000000000" || modal.Title != "自動發送通知系統!" {
		t.Fatalf("modal = %#v", modal)
	}
	inputs := flattenNotificationInputs(modal)
	if len(inputs) != 5 || inputs[0].CustomID != fieldCron || inputs[0].Label != "請輸入corn表達式(如想用簡化版，請直接輸入取消或cancel就可以簡易設置corn)" || inputs[0].Style != "short" || !inputs[0].Required {
		t.Fatalf("inputs = %#v", inputs)
	}
	schedules := repo.Schedules["guild-1"]
	if len(schedules) != 1 || schedules[0].ID != "1700000000000" || schedules[0].ChannelID != "channel-1" || !schedules[0].Pending {
		t.Fatalf("schedules = %#v", schedules)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != AutoNotificationSetupCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestSetupHandlerRequiresManageMessages(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.SetupHandler()(context.Background(), fakediscord.SlashInteraction(AutoNotificationSetupCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestSetupModalCompletesDirectCronAndSendsPreview(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "1700000000000", ChannelID: "channel-1", Pending: true}}
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithMessagePort(repo, nil, sideEffects, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := autoNotificationModal("1700000000000", "*/30 * * * *", "hello", "Random", "Title", "Content")
	interaction.ChannelID = "source-channel"

	if err := module.SetupModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Content, "1700000000000") || !strings.Contains(responder.Edits[0].Content, "訊息預覽") {
		t.Fatalf("edit = %#v", responder.Edits[0])
	}
	if len(repo.Completed) != 1 || repo.Completed[0].Cron != "*/30 * * * *" || repo.Completed[0].Message.EmbedTitle != "Title" {
		t.Fatalf("completed = %#v", repo.Completed)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "source-channel" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	msg := sideEffects.Sent[0].Message
	if msg.Content != "hello" || len(msg.Embeds) != 1 || msg.Embeds[0].Title != "Title" || msg.Embeds[0].Color != 0x123456 {
		t.Fatalf("preview = %#v", msg)
	}
}

func TestSetupModalRejectsInvalidColor(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.SetupModalHandler()(context.Background(), autoNotificationModal("id-1", "*/30 * * * *", "", "not-a-color", "Title", "Content"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你傳送的並不是顏色") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

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

func autoNotificationSetupSlash(channelID string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(AutoNotificationSetupCommandName, "", map[string]string{optionChannel: channelID})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func autoNotificationDeleteSlash(id string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(AutoNotificationDeleteCommandName, "", map[string]string{optionID: id})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func autoNotificationModal(id string, cron string, message string, color string, title string, content string) interactions.Interaction {
	interaction := fakediscord.ModalInteraction(interactions.ModalKey{Version: "legacy", Feature: "cron", Action: "submit"})
	interaction.CustomID = id
	interaction.ModalFields = []customid.ModalField{
		{CustomID: fieldCron, Value: cron},
		{CustomID: fieldMessage, Value: message},
		{CustomID: fieldColor, Value: color},
		{CustomID: fieldTitle, Value: title},
		{CustomID: fieldContent, Value: content},
	}
	return interaction
}

func flattenNotificationInputs(modal responses.Modal) []responses.TextInput {
	var inputs []responses.TextInput
	for _, row := range modal.Rows {
		inputs = append(inputs, row.Inputs...)
	}
	return inputs
}
