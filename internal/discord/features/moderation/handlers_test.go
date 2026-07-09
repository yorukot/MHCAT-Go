package moderation

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

func TestWarningIssueRequiresManageMessages(t *testing.T) {
	module := NewIssueModule(fakemongo.NewWarningHistoryRepository(), nil, nil, nil, nil, nil, nil, moderationFixedClock{}, nil)
	responder := fakediscord.NewResponder()
	interaction := warningIssueInteraction("user-2", "洗版")
	interaction.Actor.PermissionBits = 0

	if err := module.WarningIssueHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningIssueRejectsTargetWithHigherRole(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	settings := fakemongo.NewWarningSettingsRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.ModerationAllowed["guild-1/user-2"] = false
	module := NewIssueModule(repo, settings, nil, nil, sideEffects, nil, nil, moderationFixedClock{}, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "洗版"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "身分組位階比他低") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(repo.Histories) != 0 {
		t.Fatalf("warning should not be saved: %#v", repo.Histories)
	}
}

func TestWarningIssueSavesRendersLegacyEmbedAndDMs(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	settings := fakemongo.NewWarningSettingsRepository()
	sideEffects := fakediscord.NewSideEffects()
	discordInfo := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試伺服器"}}
	usage := &fakeusage.Tracker{}
	module := NewIssueModule(repo, settings, sideEffects, discordInfo, sideEffects, sideEffects, sideEffects, moderationFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)}, usage)
	responder := fakediscord.NewResponder()

	if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "洗版"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 成功警告這位使用者!" || responder.Edits[0].Embeds[0].Color != warningIssueSuccessColor {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	history := repo.Histories["guild-1\x00user-2"]
	if len(history.Entries) != 1 || history.Entries[0].ModeratorID != "user-1" || history.Entries[0].Reason != "洗版" || history.Entries[0].Time != "2026年07月04日 18點30分" {
		t.Fatalf("history = %#v", history)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "user-2" {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if description := sideEffects.DirectMessages[0].Message.Embeds[0].Description; !strings.Contains(description, "你在測試伺服器被__警告__了") || !strings.Contains(description, "**原因:**洗版") || !strings.Contains(description, "User(id:user-1)") {
		t.Fatalf("dm description = %q", description)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != WarningIssueCommandName || usage.Events[0].Feature != "warning-issue" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestWarningIssueThresholdSkipsNewLegacyRecord(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 1, Action: domain.WarningSettingsActionKick}
	sideEffects := fakediscord.NewSideEffects()
	module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)}, nil)
	responder := fakediscord.NewResponder()
	interaction := warningIssueInteraction("user-2", "洗版")
	interaction.ChannelID = "channel-1"

	if err := module.WarningIssueHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sideEffects.Kicked) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("new legacy record should not trigger action: kicked=%#v sent=%#v", sideEffects.Kicked, sideEffects.Sent)
	}
}

func TestWarningIssueThresholdKicksExistingWarning(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: domain.WarningSettingsActionKick}
	sideEffects := fakediscord.NewSideEffects()
	module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)}, nil)
	responder := fakediscord.NewResponder()
	interaction := warningIssueInteraction("user-2", "second")
	interaction.ChannelID = "channel-1"

	if err := module.WarningIssueHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sideEffects.Kicked) != 1 || sideEffects.Kicked[0].UserID != "user-2" || sideEffects.Kicked[0].Reason != "second" {
		t.Fatalf("kicked = %#v", sideEffects.Kicked)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者已到達警告須執行條件，成功對他執行`踢出`!" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWarningIssueThresholdBansExistingWarning(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: domain.WarningSettingsActionBan}
	sideEffects := fakediscord.NewSideEffects()
	module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)}, nil)
	responder := fakediscord.NewResponder()
	interaction := warningIssueInteraction("user-2", "second")
	interaction.ChannelID = "channel-1"

	if err := module.WarningIssueHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sideEffects.Banned) != 1 || sideEffects.Banned[0].UserID != "user-2" || sideEffects.Banned[0].Reason != "second" {
		t.Fatalf("banned = %#v", sideEffects.Banned)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者已到達警告須執行條件，成功對他執行`停權`!" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
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

func warningIssueInteraction(userID string, reason string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningIssueCommandName, "", map[string]string{
		warningOptionUser:        userID,
		warningIssueOptionReason: reason,
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

type moderationFixedClock struct {
	now time.Time
}

func (c moderationFixedClock) Now() time.Time {
	if c.now.IsZero() {
		return time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)
	}
	return c.now
}

func warningRemoveAllInteraction(userID string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(WarningRemoveAllCommandName, "", map[string]string{warningOptionUser: userID})
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}
