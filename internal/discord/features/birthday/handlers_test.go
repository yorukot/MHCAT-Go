package birthday

import (
	"context"
	"strings"
	"testing"

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
