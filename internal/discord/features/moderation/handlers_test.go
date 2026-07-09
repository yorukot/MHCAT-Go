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

func TestWarningRemoveRequiresManageMessages(t *testing.T) {
	module := NewRemovalModule(fakemongo.NewWarningRemovalRepository(), nil, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := warningRemoveInteraction("user-2", "1")
	interaction.Actor.PermissionBits = 0

	if err := module.WarningRemoveHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningRemoveSavesRendersLegacyEmbedAndDMs(t *testing.T) {
	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{
			{ModeratorID: "mod-1", Reason: "first"},
			{ModeratorID: "mod-2", Reason: "second"},
		},
	})
	sideEffects := fakediscord.NewSideEffects()
	discordInfo := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試伺服器"}}
	usage := &fakeusage.Tracker{}
	module := NewRemovalModule(repo, sideEffects, discordInfo, usage)
	responder := fakediscord.NewResponder()
	interaction := warningRemoveInteraction("user-2", "1")

	if err := module.WarningRemoveHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!" || responder.Edits[0].Embeds[0].Color != warningRemovalSuccessColor {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	remaining := repo.Histories["guild-1\x00user-2"].Entries
	if len(remaining) != 1 || remaining[0].Reason != "second" {
		t.Fatalf("remaining = %#v", remaining)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "user-2" {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if description := sideEffects.DirectMessages[0].Message.Embeds[0].Description; !strings.Contains(description, "你在測試伺服器的一個__警告__被刪除了") || !strings.Contains(description, "User(id:user-1)") {
		t.Fatalf("dm description = %q", description)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != WarningRemoveCommandName || usage.Events[0].Feature != "warning-removal" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestWarningRemoveTypedIntegerOption(t *testing.T) {
	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{
			{Reason: "first"},
			{Reason: "second"},
		},
	})
	module := NewRemovalModule(repo, nil, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := warningRemoveInteraction("user-2", "")
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		warningOptionUser: {
			Type:   interactions.CommandOptionUser,
			String: "user-2",
		},
		warningRemoveOptionIndex: {
			Type: interactions.CommandOptionInteger,
			Int:  2,
		},
	}

	if err := module.WarningRemoveHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if remaining := repo.Histories["guild-1\x00user-2"].Entries; len(remaining) != 1 || remaining[0].Reason != "first" {
		t.Fatalf("remaining = %#v", remaining)
	}
}

func TestWarningRemoveMissingUsesLegacyError(t *testing.T) {
	module := NewRemovalModule(fakemongo.NewWarningRemovalRepository(), nil, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningRemoveHandler()(context.Background(), warningRemoveInteraction("user-2", "1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "這位使用者沒有任何警告!") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningRemoveAllDeletesRendersLegacyEmbedAndDMs(t *testing.T) {
	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{{Reason: "first"}},
	})
	sideEffects := fakediscord.NewSideEffects()
	discordInfo := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試伺服器"}}
	module := NewRemovalModule(repo, sideEffects, discordInfo, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningRemoveAllHandler()(context.Background(), warningRemoveAllInteraction("user-2"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if _, ok := repo.Histories["guild-1\x00user-2"]; ok {
		t.Fatalf("history should be deleted: %#v", repo.Histories)
	}
	if len(sideEffects.DirectMessages) != 1 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Description, "所有__警告__被刪除了") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
}

func TestWarningRemoveAllMissingUsesLegacyErrorWithoutExclamation(t *testing.T) {
	module := NewRemovalModule(fakemongo.NewWarningRemovalRepository(), nil, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningRemoveAllHandler()(context.Background(), warningRemoveAllInteraction("user-2"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 這位使用者沒有任何警告" {
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

func warningRemoveInteraction(userID string, index string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningRemoveCommandName, "", map[string]string{
		warningOptionUser:        userID,
		warningRemoveOptionIndex: index,
	})
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}

func warningRemoveAllInteraction(userID string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningRemoveAllCommandName, "", map[string]string{warningOptionUser: userID})
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}
