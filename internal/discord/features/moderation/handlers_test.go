package moderation

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestWarningHistoryDoesNotEnforceAdvertisedManageMessages(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{{ModeratorID: "mod-1", Reason: "reason", Time: "time"}},
	})
	module := NewModule(repo, nil, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(WarningHistoryCommandName, "", map[string]string{"使用者": "user-2"})

	if err := module.WarningHistoryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("legacy defer should be public: %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "以下是user-2的警告紀錄" || responder.Edits[0].Ephemeral {
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
	module := NewModule(repo, members, discordInfo, nil)
	responder := fakediscord.NewResponder()
	interaction := warningHistoryInteraction("user-2")

	if err := module.WarningHistoryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "以下是target的警告紀錄" || embed.Color < 0 || embed.Color >= 0x1000000 {
		t.Fatalf("embed = %#v", embed)
	}
	if !strings.Contains(embed.Description, "- 警告者: admin#0001") || !strings.Contains(embed.Description, "- 原因: 洗版") || !strings.Contains(embed.Description, "- 時間: 2026-07-04") {
		t.Fatalf("description = %q", embed.Description)
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

func TestWarningHistoryBackendFailureUsesGenericError(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Err = errors.New("mongo unavailable")
	responder := fakediscord.NewResponder()

	if err := NewModule(repo, nil, nil, nil).WarningHistoryHandler()(context.Background(), warningHistoryInteraction("user-2"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertWarningGenericError(t, responder)
}

func TestWarningSettingsRequiresManageMessages(t *testing.T) {
	module := NewSettingsModule(fakemongo.NewWarningSettingsRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := warningSettingsInteraction(domain.WarningSettingsActionBan, "3")
	interaction.Actor.PermissionBits = 0

	if err := module.WarningSettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") || responder.Edits[0].Embeds[0].Color != 0xED4245 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningSettingsSavesAndRendersLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	module := NewSettingsModule(repo, nil)
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

func TestWarningSettingsBackendFailureUsesGenericError(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	repo.Err = errors.New("mongo unavailable")
	responder := fakediscord.NewResponder()

	if err := NewSettingsModule(repo, nil).WarningSettingsHandler()(context.Background(), warningSettingsInteraction(domain.WarningSettingsActionBan, "3"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertWarningGenericError(t, responder)
}

func TestWarningSettingsPreservesNonPositiveLegacyThresholds(t *testing.T) {
	for _, threshold := range []string{"0", "-2"} {
		t.Run(threshold, func(t *testing.T) {
			repo := fakemongo.NewWarningSettingsRepository()
			module := NewSettingsModule(repo, nil)
			responder := fakediscord.NewResponder()

			if err := module.WarningSettingsHandler()(context.Background(), warningSettingsInteraction(domain.WarningSettingsActionKick, threshold), responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			want, err := strconv.ParseFloat(threshold, 64)
			if err != nil {
				t.Fatalf("parse threshold: %v", err)
			}
			if got := repo.Settings["guild-1"]; got.Threshold != want || got.Action != domain.WarningSettingsActionKick {
				t.Fatalf("saved settings = %#v", got)
			}
			if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Description != "警告成功設為警告"+threshold+"次後\n執行踢出" {
				t.Fatalf("edits = %#v", responder.Edits)
			}
		})
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
	module := NewIssueModule(repo, settings, sideEffects, discordInfo, sideEffects, sideEffects, sideEffects, moderationFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)}, nil)
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
}

func TestWarningIssuePreservesRawReasonWithoutAddingLegacyAuditReason(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: domain.WarningSettingsActionKick}
	sideEffects := fakediscord.NewSideEffects()
	module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)
	interaction := warningIssueInteraction("user-2", "  raw reason  ")

	if err := module.WarningIssueHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("handler: %v", err)
	}
	history := repo.Histories["guild-1\x00user-2"]
	if len(history.Entries) != 2 || history.Entries[1].Reason != "  raw reason  " {
		t.Fatalf("history = %#v", history)
	}
	if len(sideEffects.Kicked) != 1 || sideEffects.Kicked[0].Reason != "" {
		t.Fatalf("kicks = %#v", sideEffects.Kicked)
	}
	if len(sideEffects.DirectMessages) != 1 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Description, "**原因:**  raw reason  \n") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
}

func TestWarningIssueAcceptsAllSpaceLegacyReason(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	module := NewIssueModule(repo, nil, nil, nil, nil, nil, nil, moderationFixedClock{}, nil)

	if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "   "), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if history := repo.Histories["guild-1\x00user-2"]; len(history.Entries) != 1 || history.Entries[0].Reason != "   " {
		t.Fatalf("history = %#v", history)
	}
}

func TestWarningIssueBackendFailureUsesGenericError(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Err = errors.New("mongo unavailable")
	responder := fakediscord.NewResponder()

	if err := NewIssueModule(repo, nil, nil, nil, nil, nil, nil, moderationFixedClock{}, nil).WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "reason"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertWarningGenericError(t, responder)
}

func TestWarningIssueDMFailureDoesNotReplaceLegacySuccess(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	direct := fakediscord.NewSideEffects()
	direct.Err = errors.New("dm unavailable")
	module := NewIssueModule(repo, nil, direct, nil, nil, nil, nil, moderationFixedClock{}, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "reason"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 成功警告這位使用者!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if history := repo.Histories["guild-1\x00user-2"]; len(history.Entries) != 1 {
		t.Fatalf("history = %#v", history)
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
	if len(sideEffects.Kicked) != 1 || sideEffects.Kicked[0].UserID != "user-2" || sideEffects.Kicked[0].Reason != "" {
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
	if len(sideEffects.Banned) != 1 || sideEffects.Banned[0].UserID != "user-2" || sideEffects.Banned[0].Reason != "" {
		t.Fatalf("banned = %#v", sideEffects.Banned)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者已到達警告須執行條件，成功對他執行`停權`!" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWarningIssueThresholdUnknownActionFallsBackToKick(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: "mute"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)
	interaction := warningIssueInteraction("user-2", "second")
	interaction.ChannelID = "channel-1"

	if err := module.WarningIssueHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sideEffects.Kicked) != 1 || len(sideEffects.Banned) != 0 {
		t.Fatalf("kicked = %#v, banned = %#v", sideEffects.Kicked, sideEffects.Banned)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者已到達警告須執行條件，成功對他執行`踢出`!" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWarningIssueThresholdMessageFailureDoesNotReplaceSuccess(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	settings := fakemongo.NewWarningSettingsRepository()
	settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: domain.WarningSettingsActionKick}
	actions := fakediscord.NewSideEffects()
	messages := fakediscord.NewSideEffects()
	messages.Err = errors.New("channel unavailable")
	module := NewIssueModule(repo, settings, nil, nil, nil, actions, messages, moderationFixedClock{}, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "second"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(actions.Kicked) != 1 {
		t.Fatalf("kicked = %#v", actions.Kicked)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 成功警告這位使用者!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestWarningIssueUsesJavaScriptNumberThresholdComparison(t *testing.T) {
	t.Run("decimal", func(t *testing.T) {
		repo := fakemongo.NewWarningHistoryRepository()
		repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
		settings := fakemongo.NewWarningSettingsRepository()
		settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2.5, Action: domain.WarningSettingsActionKick}
		sideEffects := fakediscord.NewSideEffects()
		module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)
		interaction := warningIssueInteraction("user-2", "next")
		interaction.ChannelID = "channel-1"

		if err := module.WarningIssueHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
			t.Fatalf("second warning: %v", err)
		}
		if len(sideEffects.Kicked) != 0 {
			t.Fatalf("count 2 should not reach 2.5: %#v", sideEffects.Kicked)
		}
		if err := module.WarningIssueHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
			t.Fatalf("third warning: %v", err)
		}
		if len(sideEffects.Kicked) != 1 {
			t.Fatalf("count 3 should reach 2.5: %#v", sideEffects.Kicked)
		}
	})

	t.Run("malformed", func(t *testing.T) {
		repo := fakemongo.NewWarningHistoryRepository()
		repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
		settings := fakemongo.NewWarningSettingsRepository()
		settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: math.NaN(), Action: domain.WarningSettingsActionBan}
		sideEffects := fakediscord.NewSideEffects()
		module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)

		if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "next"), fakediscord.NewResponder()); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if len(sideEffects.Banned) != 0 {
			t.Fatalf("NaN threshold should never match: %#v", sideEffects.Banned)
		}
	})

	t.Run("zero", func(t *testing.T) {
		repo := fakemongo.NewWarningHistoryRepository()
		repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
		settings := fakemongo.NewWarningSettingsRepository()
		settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 0, Action: domain.WarningSettingsActionKick}
		sideEffects := fakediscord.NewSideEffects()
		module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)

		if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "next"), fakediscord.NewResponder()); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if len(sideEffects.Kicked) != 1 {
			t.Fatalf("zero threshold should match existing warning: %#v", sideEffects.Kicked)
		}
	})
}

func TestWarningIssueThresholdFailureFollowsLegacySuccessAndDM(t *testing.T) {
	for _, test := range []struct {
		name      string
		action    string
		wantError string
		configure func(*fakediscord.SideEffects)
	}{
		{name: "kick", action: domain.WarningSettingsActionKick, wantError: "我沒有權限踢出他", configure: func(sideEffects *fakediscord.SideEffects) { sideEffects.KickErr = errors.New("missing permission") }},
		{name: "ban", action: domain.WarningSettingsActionBan, wantError: "我沒有權限ban掉他", configure: func(sideEffects *fakediscord.SideEffects) { sideEffects.BanErr = errors.New("missing permission") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewWarningHistoryRepository()
			repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
			settings := fakemongo.NewWarningSettingsRepository()
			settings.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: test.action}
			sideEffects := fakediscord.NewSideEffects()
			test.configure(sideEffects)
			module := NewIssueModule(repo, settings, sideEffects, nil, sideEffects, sideEffects, sideEffects, moderationFixedClock{}, nil)
			responder := fakediscord.NewResponder()

			if err := module.WarningIssueHandler()(context.Background(), warningIssueInteraction("user-2", "reason"), responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Edits) != 2 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 成功警告這位使用者!" || !strings.Contains(responder.Edits[1].Embeds[0].Title, test.wantError) {
				t.Fatalf("edits = %#v", responder.Edits)
			}
			if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "user-2" {
				t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
			}
		})
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
	module := NewRemovalModule(repo, sideEffects, discordInfo, nil)
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
}

func TestWarningRemoveDMFailureDoesNotReplaceLegacySuccess(t *testing.T) {
	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	direct := fakediscord.NewSideEffects()
	direct.Err = errors.New("dm unavailable")
	module := NewRemovalModule(repo, direct, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.WarningRemoveHandler()(context.Background(), warningRemoveInteraction("user-2", "1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!" {
		t.Fatalf("edits = %#v", responder.Edits)
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

func TestWarningRemovePreservesJavaScriptSpliceIndexes(t *testing.T) {
	tests := []struct {
		name  string
		index string
		want  []string
	}{
		{name: "zero removes last", index: "0", want: []string{"one", "two"}},
		{name: "negative counts from end", index: "-1", want: []string{"one", "three"}},
		{name: "very negative clamps first", index: "-99", want: []string{"two", "three"}},
		{name: "large index is successful no-op", index: "99", want: []string{"one", "two", "three"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewWarningRemovalRepository()
			repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "one"}, {Reason: "two"}, {Reason: "three"}}})
			module := NewRemovalModule(repo, nil, nil, nil)
			responder := fakediscord.NewResponder()

			if err := module.WarningRemoveHandler()(context.Background(), warningRemoveInteraction("user-2", test.index), responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			entries := repo.Histories["guild-1\x00user-2"].Entries
			if len(entries) != len(test.want) {
				t.Fatalf("entries = %#v", entries)
			}
			for index, want := range test.want {
				if entries[index].Reason != want {
					t.Fatalf("entries = %#v", entries)
				}
			}
			if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!" {
				t.Fatalf("edits = %#v", responder.Edits)
			}
		})
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

func TestWarningRemovalBackendFailuresUseGenericError(t *testing.T) {
	for _, test := range []struct {
		name    string
		handler func(Module) interactions.Handler
		input   interactions.Interaction
	}{
		{name: "one", handler: func(module Module) interactions.Handler { return module.WarningRemoveHandler() }, input: warningRemoveInteraction("user-2", "1")},
		{name: "all", handler: func(module Module) interactions.Handler { return module.WarningRemoveAllHandler() }, input: warningRemoveAllInteraction("user-2")},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewWarningRemovalRepository()
			repo.Err = errors.New("mongo unavailable")
			responder := fakediscord.NewResponder()

			if err := test.handler(NewRemovalModule(repo, nil, nil, nil))(context.Background(), test.input, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			assertWarningGenericError(t, responder)
		})
	}
}

func assertWarningGenericError(t *testing.T, responder *fakediscord.Responder) {
	t.Helper()
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!" || embed.Color != 0xED4245 {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestCleanupRequiresManageMessages(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewCleanupModule(sideEffects, nil)
	responder := fakediscord.NewResponder()
	interaction := cleanupInteraction("5", "")
	interaction.Actor.PermissionBits = 0

	if err := module.CleanupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理(刪除超過200則需要有權限)`才能使用此指令" || responder.Edits[0].Embeds[0].Color != 0xED4245 || !responder.Edits[0].Ephemeral {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(sideEffects.CleanupRequests) != 0 {
		t.Fatalf("cleanup should not be requested: %#v", sideEffects.CleanupRequests)
	}
}

func TestCleanupRequiresAdministratorAboveLegacyThreshold(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewCleanupModule(sideEffects, nil)
	responder := fakediscord.NewResponder()

	if err := module.CleanupHandler()(context.Background(), cleanupInteraction("201", ""), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, cleanupPermissionLabel) {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(sideEffects.CleanupRequests) != 0 {
		t.Fatalf("cleanup should not be requested: %#v", sideEffects.CleanupRequests)
	}
}

func TestCleanupRejectsMoreThanLegacyMaximum(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewCleanupModule(sideEffects, nil)
	responder := fakediscord.NewResponder()

	if err := module.CleanupHandler()(context.Background(), cleanupInteraction("1001", ""), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "不可刪除超過1000則消息!!!!!") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(sideEffects.CleanupRequests) != 0 {
		t.Fatalf("cleanup should not be requested: %#v", sideEffects.CleanupRequests)
	}
}

func TestCleanupRejectsNonPositiveCount(t *testing.T) {
	for _, count := range []string{"0", "-1"} {
		sideEffects := fakediscord.NewSideEffects()
		module := NewCleanupModule(sideEffects, nil)
		responder := fakediscord.NewResponder()

		if err := module.CleanupHandler()(context.Background(), cleanupInteraction(count, ""), responder); err != nil {
			t.Fatalf("handler count %s: %v", count, err)
		}
		if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "未知的錯誤") {
			t.Fatalf("edits count %s = %#v", count, responder.Edits)
		}
		if len(sideEffects.CleanupRequests) != 0 {
			t.Fatalf("cleanup should not be requested for count %s: %#v", count, sideEffects.CleanupRequests)
		}
	}
}

func TestCleanupCleanerErrorUsesGenericLegacyError(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.CleanupErr = errors.New("discord unavailable")
	module := NewCleanupModule(sideEffects, nil)
	responder := fakediscord.NewResponder()

	if err := module.CleanupHandler()(context.Background(), cleanupInteraction("5", ""), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestCleanupRequestsMessagesAndRendersLegacyCompletion(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.CleanupDeleted = 7
	module := NewCleanupModule(sideEffects, nil)
	responder := fakediscord.NewResponder()
	interaction := cleanupInteraction("", "user-2")
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		cleanupOptionCount: {
			Type: interactions.CommandOptionInteger,
			Int:  10,
		},
		cleanupOptionUser: {
			Type:   interactions.CommandOptionUser,
			String: "user-2",
		},
	}

	if err := module.CleanupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(sideEffects.CleanupRequests) != 1 {
		t.Fatalf("cleanup requests = %#v", sideEffects.CleanupRequests)
	}
	request := sideEffects.CleanupRequests[0]
	if request.ChannelID != "channel-1" || request.Limit != 10 || request.UserID != "user-2" {
		t.Fatalf("cleanup request = %#v", request)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614> | 清理完成!" || embed.Color != 0x53FF53 {
		t.Fatalf("embed = %#v", embed)
	}
	if embed.Description != "**成功清除:**`7`/`10`\n**<:deletebutton:981971559679950848> 如果沒有成功清完全\n代表可能超過14天或沒這麼多訊息給清**" || !responder.Edits[0].Ephemeral {
		t.Fatalf("description = %q", embed.Description)
	}
}

func TestDeleteDataPromptRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()
	interaction := deleteDataSlashInteraction()
	interaction.Actor.PermissionBits = 0

	if err := module.DeleteDataPromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("legacy slash defer should be public: %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令" || responder.Edits[0].Embeds[0].Color != 0xED4245 || responder.Edits[0].Ephemeral {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(responder.Follow) != 0 {
		t.Fatalf("prompt should not be sent without permission: %#v", responder.Follow)
	}
}

func TestDeleteDataPromptRendersLegacyMenu(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataPromptHandler()(context.Background(), deleteDataSlashInteraction(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 {
		t.Fatalf("followups = %#v", responder.Follow)
	}
	embed := responder.Follow[0].Embeds[0]
	warningImageURL := "https://media.discordapp.net/attachments/991337796960784424/996749656161779853/6lnjr0.gif"
	if embed.Title != "<:trashbin:995991389043163257> 刪除資料" || embed.Description != "<a:NukeExplosion:986558305885368321>這邊刪除的都是全刪!!!\n<:warning:985590881698590730> 一但刪除將__**無法復原**__，請三思!\n<:warning:985590881698590730> 一但刪除將__**無法復原**__，請三思!" || embed.Color < 0 || embed.Color >= 0x1000000 || embed.Footer == nil || embed.Footer.Text != "請三思!!!" || embed.Footer.IconURL != warningImageURL || embed.Thumbnail == nil || embed.Thumbnail.URL != warningImageURL {
		t.Fatalf("embed = %#v", embed)
	}
	if responder.Follow[0].Ephemeral || responder.Follow[0].AllowedMentions == nil {
		t.Fatalf("legacy prompt visibility/mentions = %#v", responder.Follow[0])
	}
	if len(responder.Follow[0].Components) != 1 || len(responder.Follow[0].Components[0].Components) != 1 {
		t.Fatalf("components = %#v", responder.Follow[0].Components)
	}
	selectMenu := responder.Follow[0].Components[0].Components[0]
	if selectMenu.CustomID != "delete-data" || selectMenu.Placeholder != "🗑 選擇你要刪除的資料!" || selectMenu.MinValues != 1 || selectMenu.MaxValues != 1 || len(selectMenu.Options) != 9 {
		t.Fatalf("select menu = %#v", selectMenu)
	}
	wantEmojis := []string{
		"<:joines:953970547849592884>",
		"<:leaves:956444050792280084>",
		"<:logfile:985948561625710663>",
		"<:statistics:986108146747600928>",
		"<:ChatBot:956863473910947850>",
		"<:tickmark:985949769224556614>",
		"<:xp:990254386792005663>",
		"<:Voice:994844272790610011>",
		"<:ticket:985945491093205073>",
	}
	for index, target := range domain.LegacyDeleteDataTargets() {
		option := selectMenu.Options[index]
		if option.Label != string(target) || option.Value != string(target) || option.Description != "🗑 "+string(target)+" 刪除!" || option.Emoji != wantEmojis[index] || option.Default {
			t.Fatalf("option %d = %#v", index, option)
		}
	}
}

func TestDeleteDataSelectDeletesTargetWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataSelectHandler()(context.Background(), deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("legacy component defer should be public: %#v", responder.Defers)
	}
	if len(repo.Deleted) != 1 || repo.Deleted[0].Target != domain.DeleteDataTargetAutoChat {
		t.Fatalf("deleted = %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != "<a:green_tick:994529015652163614> **| 成功刪除該設定!**" || responder.Edits[0].Ephemeral {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteDataSelectDispatchesEveryLegacyTargetInGuild(t *testing.T) {
	for _, target := range domain.LegacyDeleteDataTargets() {
		t.Run(string(target), func(t *testing.T) {
			repo := fakemongo.NewDeleteDataRepository()
			for _, candidate := range domain.LegacyDeleteDataTargets() {
				repo.Put("guild-1", candidate)
				repo.Put("guild-2", candidate)
			}
			module := NewDeleteDataModule(repo)

			if err := module.DeleteDataSelectHandler()(context.Background(), deleteDataComponentInteraction(target), fakediscord.NewResponder()); err != nil {
				t.Fatalf("handle target %q: %v", target, err)
			}
			if len(repo.Deleted) != 1 || repo.Deleted[0] != (domain.DeleteDataRequest{GuildID: "guild-1", Target: target}) {
				t.Fatalf("deleted = %#v", repo.Deleted)
			}
			if repo.Existing["guild-1"][target] {
				t.Fatalf("target %q still exists for invoking guild", target)
			}
			if !repo.Existing["guild-2"][target] {
				t.Fatalf("target %q was deleted from another guild", target)
			}
			for _, candidate := range domain.LegacyDeleteDataTargets() {
				if candidate != target && !repo.Existing["guild-1"][candidate] {
					t.Fatalf("unselected target %q was deleted", candidate)
				}
			}
		})
	}
}

func TestDeleteDataSelectRejectsForeignPromptUser(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	module := NewDeleteDataModule(repo)
	interaction := deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat)
	interaction.Actor.UserID = "user-2"
	interaction.OriginalInteractionUserID = "user-1"
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Deleted) != 0 {
		t.Fatalf("foreign user deleted data: %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != "<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**" || responder.Edits[0].Ephemeral {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteDataSelectHonorsLegacyCollectorLifetime(t *testing.T) {
	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		now         time.Time
		wantDeleted bool
	}{
		{name: "just before one hour", now: createdAt.Add(time.Hour - time.Millisecond), wantDeleted: true},
		{name: "exactly one hour", now: createdAt.Add(time.Hour), wantDeleted: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewDeleteDataRepository()
			repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
			module := NewDeleteDataModuleWithClock(repo, moderationFixedClock{now: test.now})
			interaction := deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat)
			interaction.OriginalInteractionID = deleteDataTestSnowflake(createdAt)
			responder := fakediscord.NewResponder()

			if err := module.DeleteDataSelectHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handle delete data select: %v", err)
			}
			if got := len(repo.Deleted) == 1; got != test.wantDeleted {
				t.Fatalf("deleted = %#v", repo.Deleted)
			}
			if !test.wantDeleted && (len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "你沒有設定過這個選項")) {
				t.Fatalf("expiry response = %#v", responder.Edits)
			}
		})
	}
}

func TestDeleteDataSelectKeepsCompatibilityForMissingOrMalformedPromptID(t *testing.T) {
	for _, originalInteractionID := range []string{"", "not-a-snowflake"} {
		t.Run(originalInteractionID, func(t *testing.T) {
			repo := fakemongo.NewDeleteDataRepository()
			repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
			module := NewDeleteDataModuleWithClock(repo, moderationFixedClock{now: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
			interaction := deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat)
			interaction.OriginalInteractionID = originalInteractionID

			if err := module.DeleteDataSelectHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
				t.Fatalf("handle delete data select: %v", err)
			}
			if len(repo.Deleted) != 1 {
				t.Fatalf("deleted = %#v", repo.Deleted)
			}
		})
	}
}

func TestDeleteDataSelectMissingUsesLegacyContent(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataSelectHandler()(context.Background(), deleteDataComponentInteraction(domain.DeleteDataTargetTicket), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != "<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteDataSelectBackendFailureUsesLegacyMissingContent(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	repo.Err = errors.New("mongo unavailable")
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataSelectHandler()(context.Background(), deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Deleted) != 0 {
		t.Fatalf("backend failure recorded deletion: %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != "<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**" || responder.Edits[0].Ephemeral {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteDataSelectMetadataLessPanelRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	module := NewDeleteDataModule(repo)
	responder := fakediscord.NewResponder()
	interaction := deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat)
	interaction.Actor.PermissionBits = 0
	interaction.OriginalInteractionUserID = ""

	if err := module.DeleteDataSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Deleted) != 0 {
		t.Fatalf("delete should not run without permission: %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteDataSelectPromptOwnerDoesNotRecheckManageMessages(t *testing.T) {
	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	module := NewDeleteDataModule(repo)
	interaction := deleteDataComponentInteraction(domain.DeleteDataTargetAutoChat)
	interaction.Actor.PermissionBits = 0
	responder := fakediscord.NewResponder()

	if err := module.DeleteDataSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.Deleted) != 1 || repo.Deleted[0].Target != domain.DeleteDataTargetAutoChat {
		t.Fatalf("prompt owner delete = %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "成功刪除該設定") {
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

func cleanupInteraction(count string, userID string) interactions.Interaction {
	options := map[string]string{cleanupOptionCount: count}
	if strings.TrimSpace(userID) != "" {
		options[cleanupOptionUser] = userID
	}
	interaction := fakediscord.SlashInteractionWithOptions(CleanupCommandName, "", options)
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	interaction.ChannelID = "channel-1"
	return interaction
}

func deleteDataSlashInteraction() interactions.Interaction {
	interaction := fakediscord.SlashInteraction(DeleteDataCommandName)
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	return interaction
}

func deleteDataComponentInteraction(target domain.DeleteDataTarget) interactions.Interaction {
	interaction := fakediscord.ComponentInteractionFromID("delete-data")
	interaction.Actor.PermissionBits = warningManageMessagesPermission
	interaction.OriginalInteractionUserID = interaction.Actor.UserID
	interaction.Values = []string{string(target)}
	return interaction
}

func deleteDataTestSnowflake(createdAt time.Time) string {
	const discordEpoch = int64(1420070400000)
	return strconv.FormatUint(uint64(createdAt.UnixMilli()-discordEpoch)<<22, 10)
}
