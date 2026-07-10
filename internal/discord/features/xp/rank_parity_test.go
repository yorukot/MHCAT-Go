package xp

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
)

func TestRankComponentsPreserveLegacyRows(t *testing.T) {
	const viewerID = "123456789012345678"
	tests := []struct {
		name string
		page coreservice.RankPage
		want []responses.ComponentRow
	}{
		{
			name: "text viewer at rank ten",
			page: coreservice.RankPage{ViewerID: viewerID, Page: 0, TotalPages: 2, ViewerRank: 10, ViewerHasProfile: true, Kind: coreservice.RankKindText},
			want: rankComponentRows(viewerID, "text_rank", 0, 2, "1"),
		},
		{
			name: "voice page preserves text label id",
			page: coreservice.RankPage{ViewerID: viewerID, Page: 1, TotalPages: 3, ViewerRank: 20, ViewerHasProfile: true, Kind: coreservice.RankKindVoice},
			want: rankComponentRows(viewerID, "voice_rank", 1, 3, "2"),
		},
		{
			name: "empty leaderboard keeps target row",
			page: coreservice.RankPage{ViewerID: viewerID, Page: 0, TotalPages: 0, Kind: coreservice.RankKindText},
			want: rankComponentRows(viewerID, "text_rank", 0, 0, "NaN"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := rankComponents(test.page); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("rankComponents() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestRankLoadingMessagePreservesLegacyPayload(t *testing.T) {
	tests := map[string]string{
		"missing avatar":         rankDefaultAvatar,
		"Discord default avatar": rankDefaultAvatar,
		"animated avatar":        "https://cdn.discordapp.com/avatars/123/a_hash.png?size=128",
		"static avatar":          "https://example.test/avatar.png",
	}
	inputs := map[string]string{
		"missing avatar":         "",
		"Discord default avatar": "https://cdn.discordapp.com/embed/avatars/2.png",
		"animated avatar":        "https://cdn.discordapp.com/avatars/123/a_hash.gif?size=128",
		"static avatar":          "https://example.test/avatar.png",
	}

	for name, wantAvatar := range tests {
		t.Run(name, func(t *testing.T) {
			want := responses.Message{
				Embeds: []responses.Embed{{
					Author: &responses.EmbedAuthor{Name: rankLoadingAuthor, IconURL: rankLoadingIcon},
					Footer: &responses.EmbedFooter{Text: rankLoadingFooter, IconURL: wantAvatar},
					Color:  rankLoadingColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := rankLoadingMessage(inputs[name]); !reflect.DeepEqual(got, want) {
				t.Fatalf("rankLoadingMessage() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRankMissingUserMessagePreservesLegacyPayload(t *testing.T) {
	want := responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: rankMissingUserTitle,
			Color: rankMissingUserColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := rankMissingUserMessage(); !reflect.DeepEqual(got, want) {
		t.Fatalf("rankMissingUserMessage() = %#v, want %#v", got, want)
	}
}

func TestLookupRankUsernamePreservesLegacyUserTag(t *testing.T) {
	tests := []struct {
		name string
		info ports.DiscordUserInfo
		want string
	}{
		{name: "legacy discriminator", info: ports.DiscordUserInfo{Username: "Yoru", Nickname: "Guild Nick", Discriminator: "1234"}, want: "Yoru#1234"},
		{name: "migrated username", info: ports.DiscordUserInfo{Username: "yoru", Discriminator: "0"}, want: "yoru"},
		{name: "missing discriminator", info: ports.DiscordUserInfo{Username: "Yoru"}, want: "Yoru"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{"user-1": test.info}}
			module := RankModule{guilds: provider}
			if got := module.lookupRankUsername(context.Background(), "guild-1", "user-1"); got != test.want {
				t.Fatalf("lookupRankUsername() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestLookupRankUsernameUsesLegacyMissingText(t *testing.T) {
	provider := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{}}
	module := RankModule{guilds: provider}
	if got := module.lookupRankUsername(context.Background(), "guild-1", "missing"); got != rankMissingUsername {
		t.Fatalf("lookupRankUsername() = %q, want %q", got, rankMissingUsername)
	}
}

func TestTruncateLegacyRankText(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "ASCII boundary", value: "1234567890123456789012345678901234", want: "1234567890123456789012345678901234"},
		{name: "ASCII truncation", value: "12345678901234567890123456789012345", want: "123456789012345678901234567890123"},
		{name: "wide boundary", value: "一二三四五六七八九十一二三四五六七", want: "一二三四五六七八九十一二三四五六七"},
		{name: "wide truncation", value: "一二三四五六七八九十一二三四五六七八", want: "一二三四五六七八九十一二三四五六"},
		{name: "mixed prefix", value: "123456789012345678901234567890123界", want: "123456789012345678901234567890123"},
		{name: "emoji UTF-16 width", value: "1234567890123456789012345678901😀", want: "1234567890123456"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := truncateLegacyRankText(test.value); got != test.want {
				t.Fatalf("truncateLegacyRankText(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestParseLegacyXPRankPageRequestBoundsPages(t *testing.T) {
	const viewerID = "123456789012345678"
	viewer, page, kind, err := parseLegacyXPRankPageRequest("[" + viewerID + "]voice_rank {12}")
	if err != nil || viewer != viewerID || page != 12 || kind != coreservice.RankKindVoice {
		t.Fatalf("valid target request = viewer %q page %d kind %q err %v", viewer, page, kind, err)
	}

	tooLarge := strconv.Itoa(int(^uint(0) >> 1))
	for _, customID := range []string{
		"[" + viewerID + "]{" + tooLarge + "}text_rank",
		"[" + viewerID + "]text_rank {" + tooLarge + "}",
	} {
		if _, _, _, err := parseLegacyXPRankPageRequest(customID); err == nil {
			t.Fatalf("parseLegacyXPRankPageRequest(%q) unexpectedly succeeded", customID)
		}
	}
}

func rankComponentRows(viewerID string, kind string, page int, totalPages int, targetPage string) []responses.ComponentRow {
	button := func(customID string, label string, emoji string, style responses.ButtonStyle, disabled bool) responses.Component {
		return responses.Component{Type: responses.ComponentTypeButton, CustomID: customID, Label: label, Emoji: emoji, Style: style, Disabled: disabled}
	}
	return []responses.ComponentRow{
		{Components: []responses.Component{
			button(formatRankPageID(viewerID, page-10, kind), "", legacyRankYearBackEmoji, responses.ButtonStyleSuccess, page-10 < 0),
			button(formatRankPageID(viewerID, page-1, kind), "", legacyRankPageBackEmoji, responses.ButtonStyleSuccess, page-1 == -1),
			button("text_rank", rankPageLabel(page, totalPages), "", responses.ButtonStyleSecondary, true),
			button(formatRankPageID(viewerID, page+1, kind), "", legacyRankPageForwardEmoji, responses.ButtonStyleSuccess, page+1 >= totalPages),
			button(formatRankPageID(viewerID, page+10, kind), "", legacyRankYearForwardEmoji, responses.ButtonStyleSuccess, page+10 >= totalPages),
		}},
		{Components: []responses.Component{
			button("text_rank1", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
			button("text_rank2", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
			button("["+viewerID+"]"+kind+" {"+targetPage+"}", "", legacyRankTargetViewerEmoji, responses.ButtonStyleSecondary, false),
			button("text_rank4", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
			button("text_rank5", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
		}},
	}
}

func formatRankPageID(viewerID string, page int, kind string) string {
	return "[" + viewerID + "]{" + strconv.Itoa(page) + "}" + kind
}

func rankPageLabel(page int, totalPages int) string {
	return strconv.Itoa(page+1) + "/" + strconv.Itoa(totalPages)
}
