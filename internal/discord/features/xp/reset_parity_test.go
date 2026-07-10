package xp

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestXPResetSlashMessagesMatchLegacy(t *testing.T) {
	for _, tc := range []struct {
		name  string
		label string
	}{
		{name: "text", label: "聊天"},
		{name: "voice", label: "語音"},
	} {
		t.Run(tc.name+" success", func(t *testing.T) {
			want := responses.Message{
				Content:         "<a:green_tick:994529015652163614> | 成功清除<@user-2>的" + tc.label + "經驗",
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := xpResetProfileSuccessMessage(" user-2 ", tc.label); !reflect.DeepEqual(got, want) {
				t.Fatalf("profile success = %#v, want %#v", got, want)
			}
		})
	}

	wantWarning := responses.Message{
		Content:         ":warning: | 一但刪除，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!",
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := xpResetWarningMessage(); !reflect.DeepEqual(got, wantWarning) {
		t.Fatalf("warning = %#v, want %#v", got, wantWarning)
	}

	wantMissing := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 這位使用者還沒有任何的經驗值喔!",
			Color: 0xED4245,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	for _, err := range []error{ports.ErrTextXPProfileMissing, ports.ErrVoiceXPProfileMissing} {
		if got := xpResetProfileErrorMessage(err); !reflect.DeepEqual(got, wantMissing) {
			t.Fatalf("missing profile = %#v, want %#v", got, wantMissing)
		}
	}
}

func TestXPResetConfirmationMessagesMatchLegacy(t *testing.T) {
	wantWrong := ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你輸入了錯誤的確認!因此視為取消還原",
			Color: 0xED4245,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
	if got := xpResetWrongConfirmationOutbound(); !reflect.DeepEqual(got, wantWrong) {
		t.Fatalf("wrong confirmation = %#v, want %#v", got, wantWrong)
	}

	for _, tc := range []struct {
		name  string
		kind  resetKind
		label string
	}{
		{name: "text", kind: resetKindText, label: "聊天"},
		{name: "voice", kind: resetKindVoice, label: "語音"},
	} {
		t.Run(tc.name+" success", func(t *testing.T) {
			want := ports.OutboundMessage{
				Embeds: []ports.OutboundEmbed{{
					Title: "<:trashbin:995991389043163257>成功刪除伺服器內所有" + tc.label + "經驗",
					Color: 0x53FF53,
				}},
				AllowedMentions: ports.AllowedMentions{},
			}
			if got := xpResetGuildSuccessOutbound(tc.kind); !reflect.DeepEqual(got, want) {
				t.Fatalf("guild success = %#v, want %#v", got, want)
			}
		})
	}

	for _, tc := range []struct {
		name    string
		err     error
		kind    resetKind
		message string
	}{
		{name: "empty voice", err: ports.ErrVoiceXPProfileMissing, kind: resetKindVoice, message: "伺服器沒有任何語音經驗的資料!"},
		{name: "unknown", err: errors.New("database unavailable"), kind: resetKindText, message: "很抱歉，出現了未知的錯誤，請重試!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := ports.OutboundMessage{
				Embeds: []ports.OutboundEmbed{{
					Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + tc.message,
					Color: 0xED4245,
				}},
				AllowedMentions: ports.AllowedMentions{},
			}
			if got := xpResetGuildErrorOutbound(tc.err, tc.kind); !reflect.DeepEqual(got, want) {
				t.Fatalf("guild error = %#v, want %#v", got, want)
			}
		})
	}
}

func TestXPResetConfirmationStoreScopesExpiresAndUsesLatestRequest(t *testing.T) {
	clock := &xpResetTestClock{now: time.Unix(1000, 0)}
	store := newResetConfirmationStore(clock, time.Minute)
	text := resetConfirmation{GuildID: "guild-1", ChannelID: "channel-1", UserID: "user-1", Kind: resetKindText}
	store.Put(text)

	for _, event := range []events.Event{
		{GuildID: "other-guild", ChannelID: "channel-1", UserID: "user-1", CreatedAt: clock.now.Add(time.Second)},
		{GuildID: "guild-1", ChannelID: "other-channel", UserID: "user-1", CreatedAt: clock.now.Add(time.Second)},
		{GuildID: "guild-1", ChannelID: "channel-1", UserID: "other-user", CreatedAt: clock.now.Add(time.Second)},
	} {
		if _, ok := store.Take(event); ok {
			t.Fatalf("unscoped event consumed the confirmation: %#v", event)
		}
	}
	confirmation, ok := store.Take(events.Event{GuildID: "guild-1", ChannelID: "channel-1", UserID: "user-1", CreatedAt: clock.now.Add(time.Second)})
	if !ok || confirmation.Kind != resetKindText {
		t.Fatalf("scoped confirmation = %#v/%t", confirmation, ok)
	}

	store.Put(text)
	if _, ok := store.Take(events.Event{GuildID: "guild-1", ChannelID: "channel-1", UserID: "user-1", CreatedAt: clock.now.Add(time.Minute)}); ok {
		t.Fatal("confirmation should expire at the 60-second boundary")
	}

	store.Put(text)
	store.Put(resetConfirmation{GuildID: "guild-1", ChannelID: "channel-1", UserID: "user-1", Kind: resetKindVoice})
	confirmation, ok = store.Take(events.Event{GuildID: "guild-1", ChannelID: "channel-1", UserID: "user-1", CreatedAt: clock.now.Add(time.Second)})
	if !ok || confirmation.Kind != resetKindVoice {
		t.Fatalf("latest confirmation = %#v/%t", confirmation, ok)
	}
}
