package moderation

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

func TestWarningHistoryRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	module := NewModule(repo, nil, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(WarningHistoryCommandName, "", map[string]string{"使用者": "user-2"})

	if err := module.WarningHistoryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningHistoryRendersLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{{
			ModeratorID: "mod-1",
			Reason:      "洗版",
			Time:        "2026-07-04",
		}},
	})
	members := fakediscord.NewSideEffects()
	members.MemberTagValues["mod-1"] = "admin#0001"
	discordInfo := &fakebotinfo.DiscordInfoProvider{User: ports.DiscordUserInfo{ID: "user-2", Username: "target"}}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, members, discordInfo, usage)
	responder := fakediscord.NewResponder()
	interaction := warningHistoryInteraction("user-2")

	if err := module.WarningHistoryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "以下是target的警告紀錄" || embed.Color != warningHistoryColor {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "- 警告者: admin#0001") || !strings.Contains(embed.Description, "- 原因: 洗版") || !strings.Contains(embed.Description, "- 時間: 2026-07-04") {
		t.Fatalf("description = %q", embed.Description)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != WarningHistoryCommandName || usage.Events[0].Feature != "warnings" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestWarningHistoryModeratorFallbackAvoidsLegacyCrash(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{{ModeratorID: "missing-mod", Reason: "reason", Time: "time"}},
	})
	module := NewModule(repo, fakediscord.NewSideEffects(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningHistoryHandler()(context.Background(), warningHistoryInteraction("user-2"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "- 警告者: missing-mod") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningHistoryMissingUsesLegacyError(t *testing.T) {
	module := NewModule(fakemongo.NewWarningHistoryRepository(), nil, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningHistoryHandler()(context.Background(), warningHistoryInteraction("user-2"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "這位使用者沒有任何警告") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningSettingsRequiresManageMessages(t *testing.T) {
	module := NewSettingsModule(fakemongo.NewWarningSettingsRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := warningSettingsInteraction(domain.WarningSettingsActionBan, "3")
	interaction.Actor.PermissionBits = 0

	if err := module.WarningSettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningSettingsSavesAndRendersLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	usage := &fakeusage.Tracker{}
	module := NewSettingsModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := warningSettingsInteraction(domain.WarningSettingsActionKick, "4")

	if err := module.WarningSettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "警告系統" || embed.Description != "警告成功設為警告4次後\n執行踢出" || embed.Color != warningSettingsSuccessColor {
		t.Fatalf("embed = %#v", embed)
	}
	got := repo.Settings["guild-1"]
	if got.Threshold != 4 || got.Action != domain.WarningSettingsActionKick {
		t.Fatalf("saved settings = %#v", got)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != WarningSettingsCommandName || usage.Events[0].Feature != "warning-settings" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestWarningSettingsTypedIntegerOption(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	module := NewSettingsModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := warningSettingsInteraction(domain.WarningSettingsActionBan, "")
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		warningSettingsOptionAction: {
			Type:   interactions.CommandOptionString,
			String: domain.WarningSettingsActionBan,
		},
		warningSettingsOptionThreshold: {
			Type: interactions.CommandOptionInteger,
			Int:  2,
		},
	}

	if err := module.WarningSettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := repo.Settings["guild-1"]; got.Threshold != 2 || got.Action != domain.WarningSettingsActionBan {
		t.Fatalf("saved settings = %#v", got)
	}
}

func TestWarningSettingsInvalidInputUsesGenericError(t *testing.T) {
	module := NewSettingsModule(fakemongo.NewWarningSettingsRepository(), nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningSettingsHandler()(context.Background(), warningSettingsInteraction("mute", "0"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func warningHistoryInteraction(userID string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningHistoryCommandName, "", map[string]string{"使用者": userID})
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}

func warningSettingsInteraction(action string, threshold string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningSettingsCommandName, "", map[string]string{
		warningSettingsOptionAction:    action,
		warningSettingsOptionThreshold: threshold,
	})
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}
