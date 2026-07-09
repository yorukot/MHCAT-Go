package economy

import (
	"bytes"
	"context"
	"errors"
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

func TestProfileSlashRepliesLoadingThenRendersLegacyPNGAndRefreshButton(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	userID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: userID, Coins: 1200, Today: 1})
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", GachaCost: 500, SignCoins: 25, XPMultiple: 1.5})
	repo.PutWorkConfig(domain.WorkConfig{GuildID: "guild-1", DailyEnergy: 10, MaxEnergy: 100})
	repo.PutWorkUser(domain.WorkUserState{GuildID: "guild-1", UserID: userID, Energy: 50, EndTimeUnix: 1_800})
	repo.PutTextXP(domain.XPProfile{GuildID: "guild-1", UserID: userID, XP: 40, Level: 2})
	repo.PutVoiceXP(domain.XPProfile{GuildID: "guild-1", UserID: userID, XP: 80, Level: 3})
	info := &fakebotinfo.DiscordInfoProvider{
		Users: map[string]ports.DiscordUserInfo{userID: {
			ID:            userID,
			Username:      "User",
			Nickname:      "Nick",
			Discriminator: "1234",
			CreatedAt:     time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			JoinedAt:      time.Date(2021, 2, 3, 0, 0, 0, 0, time.UTC),
		}},
		Guild: ports.DiscordGuildInfo{Name: "Guild"},
	}
	usage := &fakeusage.Tracker{}
	module := NewProfileModule(repo, info, fixedClock{now: time.Unix(1_000, 0)}, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(ProfileCommandName)
	interaction.Actor.UserID = userID
	interaction.Actor.AvatarURL = "https://example.test/avatar.png"

	if err := module.ProfileHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 || responder.Replies[0].Embeds[0].Author.Name != signLoadingAuthor {
		t.Fatalf("loading reply = %#v", responder.Replies)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	assertProfileImage(t, responder.Edits[0].Files)
	if len(responder.Edits[0].Components) != 1 || len(responder.Edits[0].Components[0].Components) != 1 {
		t.Fatalf("components = %#v", responder.Edits[0].Components)
	}
	button := responder.Edits[0].Components[0].Components[0]
	if button.CustomID != userID+"my-profile" || button.Label != profileRefreshLabel || button.Emoji != profileRefreshEmoji || button.Style != responses.ButtonStyleSuccess {
		t.Fatalf("refresh button = %#v", button)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != ProfileCommandName || usage.Events[0].Feature != "economy-profile" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestProfileSlashSelectedUserUsesUserOption(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	targetID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: targetID, Coins: 1})
	info := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{targetID: {ID: targetID, Username: "Target"}}}
	module := NewProfileModule(repo, info, fixedClock{now: time.Unix(1_000, 0)}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(ProfileCommandName, "", map[string]string{"user": targetID})

	if err := module.ProfileHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(info.UserCalls) != 1 || info.UserCalls[0] != "guild-1:"+targetID {
		t.Fatalf("user calls = %#v", info.UserCalls)
	}
	button := responder.Edits[0].Components[0].Components[0]
	if button.CustomID != targetID+"my-profile" {
		t.Fatalf("button = %#v", button)
	}
}

func TestProfileRefreshMissingMemberUsesLegacyEphemeralError(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	info := &fakebotinfo.DiscordInfoProvider{UserErr: errors.New("missing")}
	module := NewProfileModule(repo, info, fixedClock{now: time.Unix(1_000, 0)}, nil)
	responder := fakediscord.NewResponder()

	if err := module.ProfileRefreshHandler()(context.Background(), fakediscord.ComponentInteractionFromID("123456789012345678my-profile"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	if title := responder.Replies[0].Embeds[0].Title; !strings.Contains(title, "已退出該伺服器") {
		t.Fatalf("title = %q", title)
	}
}

func TestProfileModuleRegistersLegacyRefreshRoute(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	userID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: userID, Coins: 1})
	info := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{userID: {ID: userID, Username: "User"}}}
	module := NewProfileModule(repo, info, fixedClock{now: time.Unix(1_000, 0)}, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.ComponentInteractionFromID(userID+"my-profile"), responder); err != nil {
		t.Fatalf("route: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	assertProfileImage(t, responder.Updates[0].Files)
}

func assertProfileImage(t *testing.T, files []responses.File) {
	t.Helper()
	if len(files) != 1 || files[0].Name != profileFileName || files[0].ContentType != profileFileContentType {
		t.Fatalf("files = %#v", files)
	}
	img, err := png.Decode(bytes.NewReader(files[0].Data))
	if err != nil {
		t.Fatalf("decode profile png: %v", err)
	}
	if bounds := img.Bounds(); bounds.Dx() != 1500 || bounds.Dy() != 750 {
		t.Fatalf("bounds = %v", bounds)
	}
}
