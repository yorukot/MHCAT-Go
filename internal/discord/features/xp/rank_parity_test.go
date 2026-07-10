package xp

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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

func TestParseLegacyXPRankPageRequestRejectsMalformedLegacyIDs(t *testing.T) {
	const viewerID = "123456789012345678"
	for _, customID := range []string{
		"[" + viewerID + "]{-1}text_rank",
		"[" + viewerID + "]{1.5}text_rank",
		"[" + viewerID + "]text_rank {NaN}",
		"[123]{1}text_rank",
		"[" + viewerID + "]{1}text_rank-extra",
		"prefix[" + viewerID + "]{1}voice_rank",
	} {
		if _, _, _, err := parseLegacyXPRankPageRequest(customID); !errors.Is(err, domain.ErrInvalidXPRankQuery) {
			t.Fatalf("parseLegacyXPRankPageRequest(%q) error = %v", customID, err)
		}
	}
}

func TestLegacyRankCanvasNumberLayout(t *testing.T) {
	if got := legacyRankNumber(0, 0); got != 1 {
		t.Fatalf("first rank = %d, want 1", got)
	}
	if got := legacyRankNumber(12, 9); got != 130 {
		t.Fatalf("last rank = %d, want 130", got)
	}
	for page, want := range map[int]int{0: 40, 99: 40, 100: 30, 999: 30, 1000: 25} {
		if got := legacyRankNumberFontSize(page); got != want {
			t.Fatalf("legacyRankNumberFontSize(%d) = %d, want %d", page, got, want)
		}
	}
	if x, row := legacyRankSlotPosition(4); x != 0 || row != 4 {
		t.Fatalf("left slot = (%d, %d)", x, row)
	}
	if x, row := legacyRankSlotPosition(5); x != 484 || row != 0 {
		t.Fatalf("right slot = (%d, %d)", x, row)
	}
}

func TestRankFontFamilyForNumericText(t *testing.T) {
	for value, want := range map[string]rankFontFamily{
		"2020/01/02": rankNumericFont,
		"1.3K":       rankNumericFont,
		"10":         rankNumericFont,
		"沒有資料!":      rankLanguageFont,
	} {
		if got := rankFontFamilyForNumericText(value); got != want {
			t.Fatalf("rankFontFamilyForNumericText(%q) = %d, want %d", value, got, want)
		}
	}
	wantLanguageFonts := []string{
		"fonts/language/TC.otf",
		"fonts/language/SC.otf",
		"fonts/language/JP.otf",
		"fonts/language/HK.otf",
		"fonts/language/NotoSans.ttf",
		"fonts/language/Bengali.ttf",
		"fonts/language/Arabic.ttf",
		"fonts/language/emoji.ttf",
		"fonts/TaipeiSansTCBeta-Regular.ttf",
	}
	if got := rankFontCandidates(rankLanguageFont); !reflect.DeepEqual(got, wantLanguageFonts) {
		t.Fatalf("language font candidates = %#v", got)
	}
	if got := rankFontCandidates(rankNumericFont); len(got) != len(wantLanguageFonts)+1 || got[0] != "fonts/Comic-Sans-MS-copy-5-.ttf" || !reflect.DeepEqual(got[1:], wantLanguageFonts) {
		t.Fatalf("numeric font candidates = %#v", got)
	}
}

func TestRenderRankPNGIncludesSlotsOnEmptyPages(t *testing.T) {
	view := rankCanvasView{
		GuildName:      "Guild",
		GuildCreatedAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		ViewerRankText: "沒有資料!",
		Title:          "聊天經驗排行榜",
	}
	first, err := renderRankPNG(view)
	if err != nil {
		t.Fatalf("render first page: %v", err)
	}
	view.Page = 1
	second, err := renderRankPNG(view)
	if err != nil {
		t.Fatalf("render second page: %v", err)
	}
	if bytes.Equal(first, second) {
		t.Fatal("empty rank pages rendered identically")
	}
}

func TestRenderRankPNGUsesFontCachesConcurrently(t *testing.T) {
	view := rankCanvasView{
		GuildName:      "多語言 Guild",
		GuildCreatedAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		ViewerRankText: "1",
		Title:          "聊天經驗排行榜",
		Entries: []rankCanvasEntry{{
			Rank:        1,
			DisplayName: "مرحبا বাংলা 日本語",
			TotalXP:     1250,
		}},
	}
	var wait sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := renderRankPNG(view)
			errs <- err
		}()
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("render rank PNG: %v", err)
		}
	}
}

func TestDrawRankGuildIconUsesProvidedImage(t *testing.T) {
	icon := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			icon.Set(x, y, color.RGBA{R: 220, G: 20, B: 30, A: 255})
		}
	}
	var data bytes.Buffer
	if err := png.Encode(&data, icon); err != nil {
		t.Fatalf("encode icon: %v", err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 120, 90))
	drawRankGuildIcon(canvas, data.Bytes())
	center := color.RGBAModel.Convert(canvas.At(68, 45)).(color.RGBA)
	if center.R < 200 || center.G > 40 || center.B > 50 || center.A != 255 {
		t.Fatalf("center pixel = %#v", center)
	}
	corner := color.RGBAModel.Convert(canvas.At(33, 10)).(color.RGBA)
	if corner.A != 0 {
		t.Fatalf("rounded corner pixel = %#v", corner)
	}
}

func TestFetchRankGuildIconReadsSuccessfulResponse(t *testing.T) {
	payload := []byte("icon bytes")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(payload)
	}))
	defer server.Close()

	if got := fetchRankGuildIcon(context.Background(), server.URL); !bytes.Equal(got, payload) {
		t.Fatalf("fetchRankGuildIcon() = %q, want %q", got, payload)
	}
}

func TestFetchRankGuildIconRejectsInvalidAndOversizedResponses(t *testing.T) {
	if got := fetchRankGuildIcon(context.Background(), "file:///tmp/icon.png"); got != nil {
		t.Fatalf("invalid URL returned %d bytes", len(got))
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(make([]byte, (2<<20)+1))
	}))
	defer server.Close()
	if got := fetchRankGuildIcon(context.Background(), server.URL); got != nil {
		t.Fatalf("oversized response returned %d bytes", len(got))
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
