package economy

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestCoinRankSlashRepliesLoadingThenRendersPNGAndLegacyButtons(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	viewerID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: viewerID, Coins: 1000})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "222222222222222222", Coins: 5000})
	info := &fakebotinfo.DiscordInfoProvider{
		Users: map[string]ports.DiscordUserInfo{
			viewerID:             {Username: "Viewer"},
			"222222222222222222": {Username: "Leader"},
		},
		Guild: ports.DiscordGuildInfo{Name: "Guild", CreatedAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
	}
	usage := &fakeusage.Tracker{}
	module := NewCoinRankModule(repo, info, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(CoinRankCommandName)
	interaction.Actor.UserID = viewerID
	interaction.Actor.AvatarURL = "https://example.test/avatar.png"

	if err := module.CoinRankHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 || responder.Replies[0].Embeds[0].Author.Name != signLoadingAuthor {
		t.Fatalf("loading reply = %#v", responder.Replies)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	assertCoinRankImage(t, responder.Edits[0].Files)
	if len(responder.Edits[0].Components) != 2 {
		t.Fatalf("components = %#v", responder.Edits[0].Components)
	}
	buttons := responder.Edits[0].Components[0].Components
	if buttons[0].CustomID != "["+viewerID+"]{-10}coin_rank" || !buttons[0].Disabled || buttons[2].Label != "1/1" {
		t.Fatalf("pagination buttons = %#v", buttons)
	}
	target := responder.Edits[0].Components[1].Components[2]
	if target.CustomID != "["+viewerID+"]coin_rank {0}" || target.Emoji != legacyRankTargetViewerEmoji || target.Disabled {
		t.Fatalf("target button = %#v", target)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CoinRankCommandName || usage.Events[0].Feature != "economy-coin-rank" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestCoinRankComponentUpdatesRequestedPage(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	viewerID := "123456789012345678"
	users := map[string]ports.DiscordUserInfo{viewerID: {Username: "Viewer"}}
	for i := 0; i < 12; i++ {
		userID := fmt.Sprintf("%018d", 222222222222222222+i)
		if i == 0 {
			userID = viewerID
		}
		repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: userID, Coins: int64(100 - i)})
		users[userID] = ports.DiscordUserInfo{Username: "User"}
	}
	module := NewCoinRankModule(repo, &fakebotinfo.DiscordInfoProvider{Users: users, Guild: ports.DiscordGuildInfo{Name: "Guild"}}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[" + viewerID + "]{1}coin_rank")

	if err := module.CoinRankPageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	assertCoinRankImage(t, responder.Updates[0].Files)
	if label := responder.Updates[0].Components[0].Components[2].Label; label != "2/2" {
		t.Fatalf("page label = %q", label)
	}
}

func TestCoinRankComponentMissingUserUsesLegacyEphemeralError(t *testing.T) {
	module := NewCoinRankModule(fakemongo.NewEconomyRepository(), &fakebotinfo.DiscordInfoProvider{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[123456789012345678]{1}coin_rank")

	if err := module.CoinRankPageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	if title := responder.Replies[0].Embeds[0].Title; !strings.Contains(title, "找不到資料") {
		t.Fatalf("title = %q", title)
	}
}

func TestCoinRankModuleRegistersLegacyComponentRoute(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	viewerID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: viewerID, Coins: 1})
	info := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{viewerID: {Username: "Viewer"}}}
	module := NewCoinRankModule(repo, info, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[" + viewerID + "]{0}coin_rank")
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
}

func assertCoinRankImage(t *testing.T, files []responses.File) {
	t.Helper()
	if len(files) != 1 || files[0].Name != coinRankFileName || files[0].ContentType != coinRankFileContentType {
		t.Fatalf("files = %#v", files)
	}
	img, err := png.Decode(bytes.NewReader(files[0].Data))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if bounds := img.Bounds(); bounds.Dx() != 1000 || bounds.Dy() != 500 {
		t.Fatalf("bounds = %v", bounds)
	}
}
