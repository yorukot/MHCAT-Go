package birthday

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

func TestHandlerReturnsStagedUnavailableForBirthdayDateSubcommands(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandAdd, map[string]string{})

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
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(BirthdayCommandName, subcommandList, map[string]string{})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	if edit.Embeds[0].Title != "🎂 生日列表" || !strings.Contains(edit.Embeds[0].Description, "<@user-2>  | 生日日期(YYYY/MM/DD):2002/3/4") {
		t.Fatalf("embed = %#v", edit.Embeds[0])
	}
	if edit.Files[0].Name != "discord.txt" || !strings.Contains(string(edit.Files[0].Data), "找不到使用者!(user-2)  | 生日日期(YYYY/MM/DD):2002/3/4") {
		t.Fatalf("files = %#v content=%s", edit.Files, string(edit.Files[0].Data))
	}
	if edit.AllowedMentions == nil {
		t.Fatalf("allowed mentions not set: %#v", edit)
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
