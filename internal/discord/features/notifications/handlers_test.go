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
	module.clock = &autoNotificationTestClock{now: time.UnixMilli(1700000000000)}
	responder := fakediscord.NewResponder()
	interaction := autoNotificationSetupSlash("channel-1")

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
	if len(repo.Completed) != 1 || repo.Completed[0].Cron != "*/30 * * * *" || repo.Completed[0].Message.EmbedTitle != "Title" || repo.Completed[0].Message.EmbedColor != "#123456" {
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

func TestSetupModalPreservesLegacyMessageWhitespace(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "id-1", ChannelID: "channel-1", Pending: true}}
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithMessagePort(repo, nil, sideEffects, nil)
	responder := fakediscord.NewResponder()
	interaction := autoNotificationModal("id-1", "*/30 * * * *", "  hello  ", "#123456", "  Title  ", "  Content  ")
	interaction.ChannelID = "source-channel"

	if err := module.SetupModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Completed) != 1 {
		t.Fatalf("completed = %#v", repo.Completed)
	}
	message := repo.Completed[0].Message
	if message.Content != "  hello  " || message.EmbedTitle != "  Title  " || message.EmbedDescription != "  Content  " {
		t.Fatalf("stored message = %#v", message)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != "  hello  " || sideEffects.Sent[0].Message.Embeds[0].Title != "  Title  " || sideEffects.Sent[0].Message.Embeds[0].Description != "  Content  " {
		t.Fatalf("preview = %#v", sideEffects.Sent)
	}
}

func TestSetupModalAcceptsWhitespaceOnlyLegacyContentAndRejectsWhitespaceColor(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "id-1", ChannelID: "channel-1", Pending: true}}
	module := NewModule(repo, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.SetupModalHandler()(context.Background(), autoNotificationModal("id-1", "*/30 * * * *", "   ", "", "", ""), responder); err != nil {
		t.Fatalf("whitespace content handler: %v", err)
	}
	if len(repo.Completed) != 1 || repo.Completed[0].Message.Content != "   " {
		t.Fatalf("completed = %#v", repo.Completed)
	}

	repo.Schedules["guild-1"] = append(repo.Schedules["guild-1"], domain.AutoNotificationSchedule{GuildID: "guild-1", ID: "id-2", ChannelID: "channel-1", Pending: true})
	responder = fakediscord.NewResponder()
	if err := module.SetupModalHandler()(context.Background(), autoNotificationModal("id-2", "*/30 * * * *", "hello", "   ", "", ""), responder); err != nil {
		t.Fatalf("whitespace color handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你傳送的並不是顏色") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestSetupModalStartsLegacySimplifiedCronWizard(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "1700000000000", ChannelID: "target-channel", Pending: true}}
	module := NewModule(repo, nil, nil)
	module.clock = &autoNotificationTestClock{now: time.Unix(1_700_000_000, 600*time.Millisecond.Nanoseconds())}
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := autoNotificationModal("1700000000000", "cancel", "hello", "", "", "")
	interaction.ChannelID = "source-channel"
	interaction.Actor.AvatarURL = "https://example.test/user.png"

	if err := module.SetupModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	message := responder.Edits[0]
	if len(message.Embeds) != 1 || message.Embeds[0].Title != "<:dailytasks:1022041880394989669> 設定corn" || message.Embeds[0].Color != 0x123456 {
		t.Fatalf("message = %#v", message)
	}
	if !strings.Contains(message.Embeds[0].Description, "<:7days:1022059380725784626>") || !strings.Contains(message.Embeds[0].Description, "<t:1700000301:R>") {
		t.Fatalf("description = %q", message.Embeds[0].Description)
	}
	if message.Embeds[0].Footer == nil || message.Embeds[0].Footer.Text != "有問題都可以前往支援伺服器詢問" || message.Embeds[0].Footer.IconURL != "https://example.test/user.png" {
		t.Fatalf("footer = %#v", message.Embeds[0].Footer)
	}
	if len(message.Components) != 1 || len(message.Components[0].Components) != 1 {
		t.Fatalf("components = %#v", message.Components)
	}
	menu := message.Components[0].Components[0]
	if !strings.HasPrefix(menu.CustomID, "mhcat:v1:cron:week:state=") || menu.Placeholder != "請選擇要在星期幾發送(可複選)" || menu.MinValues != 1 || menu.MaxValues != 7 || len(menu.Options) != 7 {
		t.Fatalf("menu = %#v", menu)
	}
	if menu.Options[0].Label != "禮拜一" || menu.Options[0].Emoji != "<:monday:1022040759614050314>" || menu.Options[6].Value != "0" {
		t.Fatalf("options = %#v", menu.Options)
	}
	if len(repo.Completed) != 0 {
		t.Fatalf("wizard should not complete before selections: %#v", repo.Completed)
	}
}

func TestSimplifiedCronWizardCompletesDraftAndSendsPreview(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "1700000000000", ChannelID: "target-channel", Pending: true}}
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithMessagePort(repo, nil, sideEffects, nil)
	module.clock = &autoNotificationTestClock{now: time.Unix(1_700_000_000, 0)}
	module.color = func() int { return 0x123456 }
	start := fakediscord.NewResponder()
	modal := autoNotificationModal("1700000000000", "取消", "hello", "Random", "Title", "Content")
	modal.ChannelID = "source-channel"
	if err := module.SetupModalHandler()(context.Background(), modal, start); err != nil {
		t.Fatalf("start wizard: %v", err)
	}

	weekResponder := fakediscord.NewResponder()
	week := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	week.Values = []string{"0", "3", "1"}
	if err := module.WeekSelectHandler()(context.Background(), week, weekResponder); err != nil {
		t.Fatalf("week handler: %v", err)
	}
	if len(weekResponder.Updates) != 1 {
		t.Fatalf("week updates = %#v", weekResponder.Updates)
	}
	hourMenu := weekResponder.Updates[0].Components[0].Components[0]
	if !strings.Contains(weekResponder.Updates[0].Embeds[0].Description, "<:24hours:1022059604747747379>") || hourMenu.MaxValues != 24 || len(hourMenu.Options) != 24 || hourMenu.Options[23].Value != "0" {
		t.Fatalf("hour update = %#v", weekResponder.Updates[0])
	}

	hourResponder := fakediscord.NewResponder()
	hour := fakediscord.ComponentInteractionFromID(hourMenu.CustomID)
	hour.Values = []string{"0", "8"}
	if err := module.HourSelectHandler()(context.Background(), hour, hourResponder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	if len(hourResponder.Updates) != 1 {
		t.Fatalf("hour updates = %#v", hourResponder.Updates)
	}
	minuteMenu := hourResponder.Updates[0].Components[0].Components[0]
	if !strings.Contains(hourResponder.Updates[0].Embeds[0].Description, "<:60minutes:1022059603153924156>") || minuteMenu.MaxValues != 6 || len(minuteMenu.Options) != 12 {
		t.Fatalf("minute update = %#v", hourResponder.Updates[0])
	}

	minuteResponder := fakediscord.NewResponder()
	minute := fakediscord.ComponentInteractionFromID(minuteMenu.CustomID)
	minute.ChannelID = "source-channel"
	minute.Values = []string{"30", "0"}
	if err := module.MinuteSelectHandler()(context.Background(), minute, minuteResponder); err != nil {
		t.Fatalf("minute handler: %v", err)
	}
	if len(repo.Completed) != 1 || repo.Completed[0].Cron != "0,30 8,0 * * 1,3,0" || repo.Completed[0].Message.EmbedTitle != "Title" || repo.Completed[0].Message.EmbedColor != "#123456" {
		t.Fatalf("completed = %#v", repo.Completed)
	}
	if len(minuteResponder.Updates) != 1 || len(minuteResponder.Updates[0].Components) != 0 || !strings.Contains(minuteResponder.Updates[0].Embeds[0].Description, "1700000000000") {
		t.Fatalf("completion update = %#v", minuteResponder.Updates)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "source-channel" || sideEffects.Sent[0].Message.Content != "hello" || sideEffects.Sent[0].Message.Embeds[0].Title != "Title" {
		t.Fatalf("preview = %#v", sideEffects.Sent)
	}

	replay := fakediscord.NewResponder()
	if err := module.MinuteSelectHandler()(context.Background(), minute, replay); err != nil {
		t.Fatalf("replay handler: %v", err)
	}
	if len(replay.Replies) != 1 || !replay.Replies[0].Ephemeral {
		t.Fatalf("replay replies = %#v", replay.Replies)
	}
}

func TestSimplifiedCronWizardRejectsWrongActorAndExpiredState(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "1700000000000", ChannelID: "target-channel", Pending: true}}
	clock := &autoNotificationTestClock{now: time.Unix(1_700_000_000, 0)}
	module := NewModule(repo, nil, nil)
	module.clock = clock
	start := fakediscord.NewResponder()
	if err := module.SetupModalHandler()(context.Background(), autoNotificationModal("1700000000000", "cancel", "hello", "", "", ""), start); err != nil {
		t.Fatalf("start wizard: %v", err)
	}
	weekID := start.Edits[0].Components[0].Components[0].CustomID

	wrongActor := fakediscord.ComponentInteractionFromID(weekID)
	wrongActor.Actor.UserID = "user-2"
	wrongActor.Values = []string{"1"}
	wrongResponder := fakediscord.NewResponder()
	if err := module.WeekSelectHandler()(context.Background(), wrongActor, wrongResponder); err != nil {
		t.Fatalf("wrong actor handler: %v", err)
	}
	if len(wrongResponder.Replies) != 1 || !wrongResponder.Replies[0].Ephemeral || len(wrongResponder.Updates) != 0 {
		t.Fatalf("wrong actor replies=%#v updates=%#v", wrongResponder.Replies, wrongResponder.Updates)
	}

	clock.now = clock.now.Add(autoNotificationWizardTTL)
	expired := fakediscord.ComponentInteractionFromID(weekID)
	expired.Values = []string{"1"}
	expiredResponder := fakediscord.NewResponder()
	if err := module.WeekSelectHandler()(context.Background(), expired, expiredResponder); err != nil {
		t.Fatalf("expired handler: %v", err)
	}
	if len(expiredResponder.Replies) != 1 || !expiredResponder.Replies[0].Ephemeral || len(expiredResponder.Updates) != 0 {
		t.Fatalf("expired replies=%#v updates=%#v", expiredResponder.Replies, expiredResponder.Updates)
	}
	if len(repo.Completed) != 0 {
		t.Fatalf("expired wizard completed schedule: %#v", repo.Completed)
	}
}

func TestSetupModalRejectsDirectCronBelowFifteenMinuteInterval(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	module.clock = &autoNotificationTestClock{now: time.Date(2026, time.July, 10, 0, 1, 0, 0, time.UTC)}
	responder := fakediscord.NewResponder()

	if err := module.SetupModalHandler()(context.Background(), autoNotificationModal("id-1", "*/5 * * * *", "hello", "", "", ""), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "傳送訊息的間隔必須大於15分鐘") || len(responder.Edits[0].Components) != 0 {
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

type autoNotificationTestClock struct {
	now time.Time
}

func (c *autoNotificationTestClock) Now() time.Time {
	return c.now
}
