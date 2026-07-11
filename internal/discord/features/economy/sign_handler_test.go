package economy

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

const (
	signTestGuildID = "111111111111111111"
	signTestUserID  = "222222222222222222"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestSignInHandlerUsesLegacyLoadingThenSignPNG(t *testing.T) {
	now := time.Date(2026, 7, 4, 10, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	repo := fakemongo.NewEconomyRepository()
	repo.PutSignInResult(domain.SignInResult{
		Balance:  domain.CoinBalance{GuildID: signTestGuildID, UserID: signTestUserID, Coins: 125, Today: 1},
		Calendar: signTestCalendar(),
		Reward:   25,
		SignedAt: now,
	})
	usage := &fakeusage.Tracker{}
	module := NewModuleWithSignIn(repo, repo, nil, fixedClock{now: now}, usage)
	responder := fakediscord.NewResponder()
	interaction := signSlashInteraction()

	if err := module.SignInHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle sign in: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 {
		t.Fatalf("expected one loading reply embed, got %#v", responder.Replies)
	}
	embed := responder.Replies[0].Embeds[0]
	if embed.Author == nil || embed.Author.Name != signLoadingAuthor || embed.Author.IconURL != signLoadingIcon || embed.Color != signLoadingColor {
		t.Fatalf("unexpected loading embed: %#v", embed)
	}
	if embed.Footer == nil || embed.Footer.Text != signLoadingFooter {
		t.Fatalf("unexpected loading footer: %#v", embed.Footer)
	}
	if embed.Footer.IconURL != "https://cdn.discordapp.com/avatars/test/a_hash.png" {
		t.Fatalf("loading avatar = %q", embed.Footer.IconURL)
	}
	assertSignCalendarEdit(t, responder.Edits)
	if len(repo.SignInCommands) != 1 || repo.SignInCommands[0].Day != "4" || repo.SignInCommands[0].Month != "07" {
		t.Fatalf("unexpected sign-in command: %#v", repo.SignInCommands)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != signInCommandName {
		t.Fatalf("unexpected usage events: %#v", usage.Events)
	}
}

func TestSignInDuplicateUsesDailyOrRollingLegacyText(t *testing.T) {
	module := NewModuleWithSignIn(fakemongo.NewEconomyRepository(), fakemongo.NewEconomyRepository(), nil, nil, nil)
	if got := module.signDuplicateText(context.Background(), signTestGuildID); got != signDailyDuplicateText {
		t.Fatalf("missing config duplicate text = %q", got)
	}
	repo := fakemongo.NewEconomyRepository()
	repo.PutConfig(domain.EconomyConfig{GuildID: signTestGuildID, ResetMarker: 3600})
	module = NewModuleWithSignIn(repo, repo, nil, nil, nil)
	if got := module.signDuplicateText(context.Background(), signTestGuildID); got != signRollingDuplicateText {
		t.Fatalf("rolling duplicate text = %q", got)
	}
}

func TestSignInHandlerRendersDuplicateCalendarInsteadOfRawError(t *testing.T) {
	now := time.Date(2026, 7, 4, 10, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	repo := fakemongo.NewEconomyRepository()
	repo.SignInErr = ports.ErrAlreadySignedIn
	repo.PutCalendar(signTestCalendar())
	module := NewModuleWithSignIn(repo, repo, nil, fixedClock{now: now}, nil)
	responder := fakediscord.NewResponder()

	if err := module.SignInHandler()(context.Background(), signSlashInteraction(), responder); err != nil {
		t.Fatalf("handle duplicate sign in: %v", err)
	}
	assertSignCalendarEdit(t, responder.Edits)
}

func TestSignNavigationButtonsUseVersionedIDs(t *testing.T) {
	rows, err := signNavigationButtons(signTestUserID, 2026, time.July)
	if err != nil {
		t.Fatalf("build buttons: %v", err)
	}
	if len(rows) != 1 || len(rows[0].Components) != 4 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
	for _, component := range rows[0].Components {
		if !strings.HasPrefix(component.CustomID, "mhcat:v1:economy:sign_page:") {
			t.Fatalf("button did not use versioned id: %#v", component)
		}
		if len(component.CustomID) > customid.MaxCustomIDLength {
			t.Fatalf("custom id too long: %d %q", len(component.CustomID), component.CustomID)
		}
		parsed, err := customid.ParseComponent(component.CustomID)
		if err != nil {
			t.Fatalf("parse custom id %q: %v", component.CustomID, err)
		}
		if parsed.Feature != "economy" || parsed.Action != "sign_page" || parsed.Payload.Values["u"] != signTestUserID {
			t.Fatalf("unexpected parsed id: %#v", parsed)
		}
	}
}

func TestSignCanvasUsesLegacyBlurredBackgroundLogoAndVerificationIcon(t *testing.T) {
	encoded, err := renderSignPNG(signCalendarView{
		Year: 2026, Month: time.July, Username: "Tester", Calendar: signTestCalendar(),
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	rendered, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expectedBackground := image.NewRGBA(image.Rect(0, 0, 1000, 707))
	drawSignBackground(expectedBackground)
	draw.Draw(expectedBackground, expectedBackground.Bounds(), &image.Uniform{C: color.RGBA{A: 128}}, image.Point{}, draw.Over)
	rgba := func(value color.Color) color.RGBA { return color.RGBAModel.Convert(value).(color.RGBA) }
	if got, want := rgba(rendered.At(10, 10)), rgba(expectedBackground.At(10, 10)); got != want {
		t.Fatalf("background pixel = %#v want %#v", got, want)
	}
	logo := loadSignAsset("asset/mhcat_white.png")
	drawSignAsset(expectedBackground, "asset/mhcat_white.png", image.Pt(20, 35))
	if logo == nil || rgba(rendered.At(50, 65)) != rgba(expectedBackground.At(50, 65)) {
		t.Fatalf("legacy logo was not rendered")
	}
	verification := loadSignAsset("asset/verify_icon.png")
	if verification == nil {
		t.Fatal("legacy verification asset was not loaded")
	}
	opaque := image.Point{-1, -1}
	for y := verification.Bounds().Min.Y; y < verification.Bounds().Max.Y && opaque.X < 0; y++ {
		for x := verification.Bounds().Min.X; x < verification.Bounds().Max.X; x++ {
			if _, _, _, alpha := verification.At(x, y).RGBA(); alpha == 0xffff {
				opaque = image.Pt(x, y)
				break
			}
		}
	}
	if opaque.X < 0 || rgba(rendered.At(883+opaque.X, 202+opaque.Y)) != rgba(verification.At(opaque.X, opaque.Y)) {
		t.Fatalf("legacy verification icon was not rendered")
	}
}

func TestSignPageHandlerSupportsLegacyButtonID(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutCalendar(signTestCalendar())
	module := NewSignInModule(repo, nil, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("/" + signTestUserID + "_sing{2026}-[7]")
	interaction.Actor.GuildID = signTestGuildID
	interaction.Actor.AvatarURL = "https://cdn.discordapp.com/avatars/clicker/a_hash.gif"

	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route legacy sign button: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("expected loading update, got %#v", responder.Updates)
	}
	loading := responder.Updates[0]
	if !loading.ClearAttachments || len(loading.Files) != 0 || len(loading.Components) != 0 || len(loading.Embeds) != 1 {
		t.Fatalf("loading update = %#v", loading)
	}
	if loading.Embeds[0].Footer == nil || loading.Embeds[0].Footer.IconURL != "https://cdn.discordapp.com/avatars/clicker/a_hash.png" {
		t.Fatalf("loading footer = %#v", loading.Embeds[0].Footer)
	}
	assertSignCalendarEdit(t, responder.Edits)
}

func TestParseSignPageRequestRejectsUnknownID(t *testing.T) {
	_, err := parseSignPageRequest("not:legacy:sign")
	if err == nil {
		t.Fatal("expected unknown sign page id to fail")
	}
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected legacy unknown error, got %v", err)
	}
}

func assertSignCalendarEdit(t *testing.T, edits []responses.Message) {
	t.Helper()
	if len(edits) != 1 || len(edits[0].Files) != 1 {
		t.Fatalf("expected one sign image edit, got %#v", edits)
	}
	file := edits[0].Files[0]
	if file.Name != signFileName || file.ContentType != signFileContentType {
		t.Fatalf("unexpected file metadata: %#v", file)
	}
	if _, err := png.Decode(bytes.NewReader(file.Data)); err != nil {
		t.Fatalf("decode sign png: %v", err)
	}
	if len(edits[0].Components) != 1 || len(edits[0].Components[0].Components) != 4 {
		t.Fatalf("expected four sign navigation buttons, got %#v", edits[0].Components)
	}
}

func signSlashInteraction() interactions.Interaction {
	interaction := fakediscord.SlashInteraction(signInCommandName)
	interaction.Actor.GuildID = signTestGuildID
	interaction.Actor.UserID = signTestUserID
	interaction.Actor.UserTag = "Tester#0001"
	interaction.Actor.AvatarURL = "https://cdn.discordapp.com/avatars/test/a_hash.gif"
	return interaction
}

func signTestCalendar() domain.SignCalendar {
	return domain.SignCalendar{
		GuildID: signTestGuildID,
		UserID:  signTestUserID,
		Date: map[string]map[string][]string{
			"2026": {"07": {"4"}},
		},
	}
}
