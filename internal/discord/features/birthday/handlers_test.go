package birthday

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestHandlerRequiresManageMessagesForConfig(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := birthdayConfigSlash()
	interaction.Actor.PermissionBits = 0

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerSavesBirthdayConfigAndRendersLegacySuccess(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()

	if err := module.Handler()(context.Background(), birthdayConfigSlash(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.GuildID != "guild-1" || saved.Message != "{user} 生日快樂" || saved.UTCOffset != "+08:00" || saved.ChannelID != "channel-1" || !saved.EveryoneCanSetBirthdayDate || saved.RoleID != "role-1" {
		t.Fatalf("saved = %#v", saved)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:cake:1065654305983570041> 生日系統祝福語設定" || embed.Color != birthdaySuccessColor || !strings.Contains(embed.Description, "UTC+08:00") || !strings.Contains(embed.Description, "<#channel-1>") || !strings.Contains(embed.Description, "<@&role-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatalf("allowed mentions not set: %#v", responder.Edits[0])
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "birthday-config" || usage.Events[0].CommandName != BirthdayCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestHandlerAcceptsTypedBooleanCommandOption(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := birthdayConfigSlash()
	delete(interaction.Options, optionEveryoneCanSet)
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		optionEveryoneCanSet: {Type: interactions.CommandOptionBoolean, Bool: false, String: "false"},
	}

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.EveryoneCanSetBirthdayDate {
		t.Fatalf("typed boolean was not used: %#v", saved)
	}
}

func TestHandlerPreservesBirthdayMessageWhitespace(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := birthdayConfigSlash()
	interaction.Options[optionMessage] = "   "

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.Message != "   " {
		t.Fatalf("saved = %#v", saved)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "\n   \n") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerAddMissingConfigUsesLegacyError(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := birthdayAddSlash()

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "請先請管理員進行祝福語設定") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerAddInvalidYearUsesLegacyError(t *testing.T) {
	module := birthdayAddModule(birthdayAddRepo(true), nil)
	responder := fakediscord.NewResponder()
	interaction := birthdayAddSlash()
	interaction.Options[optionBirthdayYear] = "2027"

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "請輸入有效的年份") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerAddPreservesExplicitZeroYear(t *testing.T) {
	repo := birthdayAddRepo(true)
	module := birthdayAddModule(repo, nil)
	start := fakediscord.NewResponder()
	interaction := birthdayAddSlash()
	interaction.Options[optionBirthdayYear] = "0"

	if err := module.Handler()(context.Background(), interaction, start); err != nil {
		t.Fatalf("handler: %v", err)
	}
	hourResponder := fakediscord.NewResponder()
	hourInteraction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	hourInteraction.Values = []string{"8"}
	if err := module.HourSelectHandler()(context.Background(), hourInteraction, hourResponder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	minuteResponder := fakediscord.NewResponder()
	minuteInteraction := fakediscord.ComponentInteractionFromID(hourResponder.Updates[0].Components[0].Components[0].CustomID)
	minuteInteraction.Values = []string{"30"}
	if err := module.MinuteSelectHandler()(context.Background(), minuteInteraction, minuteResponder); err != nil {
		t.Fatalf("minute handler: %v", err)
	}

	saved := repo.Profiles["guild-1/user-1"]
	if saved.BirthdayYear == nil || *saved.BirthdayYear != 0 {
		t.Fatalf("saved = %#v", saved)
	}
	if len(minuteResponder.Updates) != 1 || !strings.Contains(minuteResponder.Updates[0].Embeds[0].Description, "`0/7/9`") {
		t.Fatalf("updates = %#v", minuteResponder.Updates)
	}
}

func TestHandlerAddRendersLegacyHourSelect(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := birthdayAddModule(birthdayAddRepo(true), usage)
	responder := fakediscord.NewResponder()
	interaction := birthdayAddSlash()

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	embed := edit.Embeds[0]
	if embed.Title != "<:cake:1065654305983570041> 生日系統祝福語設定" || embed.Color != 0x123456 || embed.Footer == nil || embed.Footer.IconURL != "https://example.test/avatar.png" {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "<:24hours:1022059604747747379>") || !strings.Contains(embed.Description, "<t:1700000300:R>") {
		t.Fatalf("description = %q", embed.Description)
	}
	selectMenu := edit.Components[0].Components[0]
	if selectMenu.Placeholder != "請選擇要在幾點發送(24hr制)" || !strings.HasPrefix(selectMenu.CustomID, "mhcat:v1:birthday:hour:state=") || len(selectMenu.Options) != 24 {
		t.Fatalf("select = %#v", selectMenu)
	}
	if selectMenu.Options[0].Label != "1點" || selectMenu.Options[23].Label != "24點(0點)" || selectMenu.Options[23].Value != "0" {
		t.Fatalf("options = %#v", selectMenu.Options)
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "birthday-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestHandlerAddRoundsLegacyExpiryTimestamp(t *testing.T) {
	now := time.Unix(1700000000, 600*int64(time.Millisecond))
	module := NewModuleWithClock(birthdayAddRepo(true), nil, birthdayFixedClock{now: now})
	responder := fakediscord.NewResponder()

	if err := module.Handler()(context.Background(), birthdayAddSlash(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "<t:1700000301:R>") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHourSelectUpdatesToMinuteSelect(t *testing.T) {
	module := birthdayAddModule(birthdayAddRepo(true), nil)
	start := fakediscord.NewResponder()
	if err := module.Handler()(context.Background(), birthdayAddSlash(), start); err != nil {
		t.Fatalf("handler: %v", err)
	}
	hourCustomID := start.Edits[0].Components[0].Components[0].CustomID
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(hourCustomID)
	interaction.Values = []string{"8"}

	if err := module.HourSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Components) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	update := responder.Updates[0]
	if !strings.Contains(update.Embeds[0].Description, "<:60minutes:1022059603153924156>") {
		t.Fatalf("embed = %#v", update.Embeds[0])
	}
	selectMenu := update.Components[0].Components[0]
	if selectMenu.Placeholder != "請選擇要在幾分發送" || !strings.HasPrefix(selectMenu.CustomID, "mhcat:v1:birthday:minute:state=") || len(selectMenu.Options) != 12 {
		t.Fatalf("select = %#v", selectMenu)
	}
	if selectMenu.Options[3].Label != "15分" || selectMenu.Options[11].Value != "55" {
		t.Fatalf("options = %#v", selectMenu.Options)
	}
}

func TestMinuteSelectSavesProfileAndRendersLegacySuccess(t *testing.T) {
	repo := birthdayAddRepo(true)
	module := birthdayAddModule(repo, nil)
	start := fakediscord.NewResponder()
	if err := module.Handler()(context.Background(), birthdayAddSlash(), start); err != nil {
		t.Fatalf("handler: %v", err)
	}
	hourResponder := fakediscord.NewResponder()
	hourInteraction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	hourInteraction.Values = []string{"8"}
	if err := module.HourSelectHandler()(context.Background(), hourInteraction, hourResponder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	minuteResponder := fakediscord.NewResponder()
	minuteInteraction := fakediscord.ComponentInteractionFromID(hourResponder.Updates[0].Components[0].Components[0].CustomID)
	minuteInteraction.Values = []string{"30"}

	if err := module.MinuteSelectHandler()(context.Background(), minuteInteraction, minuteResponder); err != nil {
		t.Fatalf("minute handler: %v", err)
	}
	saved := repo.Profiles["guild-1/user-1"]
	if saved.BirthdayYear == nil || *saved.BirthdayYear != 2000 || saved.BirthdayMonth == nil || *saved.BirthdayMonth != 7 || saved.SendHour == nil || *saved.SendHour != 8 || saved.SendMinute == nil || *saved.SendMinute != 30 || !saved.AllowAdmin {
		t.Fatalf("saved = %#v", saved)
	}
	if len(minuteResponder.Updates) != 1 || len(minuteResponder.Updates[0].Components) != 0 {
		t.Fatalf("updates = %#v", minuteResponder.Updates)
	}
	description := minuteResponder.Updates[0].Embeds[0].Description
	if !strings.Contains(description, "以下是<@user-1>的生日日期:**`2000/7/9`") || !strings.Contains(description, "通知時間為:**`8:30`") {
		t.Fatalf("description = %q", description)
	}
	if minuteResponder.Updates[0].AllowedMentions == nil {
		t.Fatalf("allowed mentions not set: %#v", minuteResponder.Updates[0])
	}
}

func TestBirthdayAddWrongActorCannotConsumeState(t *testing.T) {
	repo := birthdayAddRepo(true)
	module := birthdayAddModule(repo, nil)
	start := fakediscord.NewResponder()
	if err := module.Handler()(context.Background(), birthdayAddSlash(), start); err != nil {
		t.Fatalf("handler: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	interaction.Actor.UserID = "user-2"
	interaction.Values = []string{"8"}

	if err := module.HourSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || len(responder.Updates) != 0 {
		t.Fatalf("replies=%#v updates=%#v", responder.Replies, responder.Updates)
	}
	if len(repo.ProfileSaved) != 0 {
		t.Fatalf("profile should not be saved: %#v", repo.ProfileSaved)
	}
}

func TestHandlerReturnsStagedUnavailableForUnknownBirthdaySubcommand(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, "未實作", map[string]string{})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "尚未在Go版本啟用") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerAllowAdminPersistsPreferenceAndRendersLegacySuccess(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandAllowAdmin, map[string]string{
		optionAllowAdmin: "false",
	})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved := repo.Profiles["guild-1/user-1"]
	if saved.GuildID != "guild-1" || saved.UserID != "user-1" || saved.AllowAdmin {
		t.Fatalf("saved = %#v", saved)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "設為**`false`") || responder.Edits[0].Embeds[0].Footer == nil {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerDeleteRequiresManageMessages(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandDelete, map[string]string{
		optionUser: "user-2",
	})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerDeleteProfileRendersLegacySuccess(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{Profiles: map[string]domain.BirthdayProfile{
		"guild-1/user-2": {GuildID: "guild-1", UserID: "user-2", AllowAdmin: true},
	}}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandDelete, map[string]string{
		optionUser: "user-2",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Profiles["guild-1/user-2"]; ok {
		t.Fatalf("profile was not deleted: %#v", repo.Profiles)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "你成功刪除了<@user-2>的資料") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestHandlerListProfilesRendersLegacyAttachment(t *testing.T) {
	year, month, day := 2002, 3, 4
	repo := &fakemongo.BirthdayConfigRepository{Profiles: map[string]domain.BirthdayProfile{
		"guild-1/user-2": {GuildID: "guild-1", UserID: "user-2", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day},
	}}
	module := NewModule(repo, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandList, map[string]string{})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	if edit.Embeds[0].Title != "🎂 生日列表" || edit.Embeds[0].Color != 0x123456 || !strings.Contains(edit.Embeds[0].Description, "<@user-2>  | 生日日期(YYYY/MM/DD):2002/3/4") {
		t.Fatalf("embed = %#v", edit.Embeds[0])
	}
	if edit.Files[0].Name != "discord.txt" || !strings.Contains(string(edit.Files[0].Data), "找不到使用者!(user-2)  | 生日日期(YYYY/MM/DD):2002/3/4") {
		t.Fatalf("files = %#v content=%s", edit.Files, string(edit.Files[0].Data))
	}
	if edit.AllowedMentions == nil {
		t.Fatalf("allowed mentions not set: %#v", edit)
	}
}

func TestHandlerListAttachmentUsesCachedLegacyUserTags(t *testing.T) {
	year, month, day := 2002, 3, 4
	repo := &fakemongo.BirthdayConfigRepository{Profiles: map[string]domain.BirthdayProfile{
		"guild-1/legacy":   {GuildID: "guild-1", UserID: "legacy", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day},
		"guild-1/migrated": {GuildID: "guild-1", UserID: "migrated", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day},
		"guild-1/missing":  {GuildID: "guild-1", UserID: "missing", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day},
	}}
	cachedUsers := &fakebotinfo.DiscordInfoProvider{CachedUsers: map[string]ports.DiscordUserInfo{
		"legacy":   {ID: "legacy", Username: "Yoru", Discriminator: "1234"},
		"migrated": {ID: "migrated", Username: "yoru", Discriminator: "0"},
	}}
	module := NewModuleWithCachedUsers(repo, cachedUsers, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandList, map[string]string{})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	content := string(responder.Edits[0].Files[0].Data)
	for _, want := range []string{
		"Yoru#1234(legacy)  | 生日日期(YYYY/MM/DD):2002/3/4",
		"yoru#0(migrated)  | 生日日期(YYYY/MM/DD):2002/3/4",
		"找不到使用者!(missing)  | 生日日期(YYYY/MM/DD):2002/3/4",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("attachment missing %q: %q", want, content)
		}
	}
}

func TestHandlerListProfilesMissingDataUsesLegacyError(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandList, map[string]string{})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "還沒有任何人有進行生日設置喔") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func birthdayConfigSlash() interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandConfig, map[string]string{
		optionMessage:        "{user} 生日快樂",
		optionChannel:        "channel-1",
		optionEveryoneCanSet: "true",
		optionUTC:            "+08:00",
		optionRole:           "role-1",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func birthdayAddSlash() interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandAdd, map[string]string{
		optionBirthdayMonth: "7",
		optionBirthdayDay:   "9",
		optionBirthdayYear:  "2000",
	})
	interaction.Actor.AvatarURL = "https://example.test/avatar.png"
	return interaction
}

func birthdayAddModule(repo *fakemongo.BirthdayConfigRepository, usage *fakeusage.Tracker) Module {
	module := NewModuleWithClock(repo, nil, birthdayFixedClock{now: time.Unix(1700000000, 0)})
	if usage != nil {
		module = NewModuleWithClock(repo, usage, birthdayFixedClock{now: time.Unix(1700000000, 0)})
	}
	module.color = func() int { return 0x123456 }
	return module
}

func birthdayAddRepo(everyoneCanSet bool) *fakemongo.BirthdayConfigRepository {
	return &fakemongo.BirthdayConfigRepository{Configs: map[string]domain.BirthdayConfig{
		"guild-1": {
			GuildID:                    "guild-1",
			Message:                    "{user} 生日快樂",
			UTCOffset:                  "+08:00",
			ChannelID:                  "channel-1",
			EveryoneCanSetBirthdayDate: everyoneCanSet,
		},
	}}
}

type birthdayFixedClock struct {
	now time.Time
}

func (c birthdayFixedClock) Now() time.Time {
	return c.now
}
