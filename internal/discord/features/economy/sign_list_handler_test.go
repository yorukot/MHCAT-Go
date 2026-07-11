package economy

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSignInListHandlerRendersLegacyDailyList(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Today: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Today: 0})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-3", Today: 1})
	discordInfo := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{
		"user-1": {ID: "user-1", Username: "Alice", Discriminator: "1234"},
		"user-3": {ID: "user-3", Username: "Bob", Discriminator: "0"},
	}}
	usage := &fakeusage.Tracker{}
	module := NewSignInOnlyModule(repo, discordInfo, fixedClock{now: time.Unix(10_000, 0)}, usage)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()

	if err := module.SignInListHandler()(context.Background(), fakediscord.SlashInteraction(SignInListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != signInListTitle || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []string{"**目前共有**`2`**人已經簽到**", "**您是否有簽到:**`有`", "┃ Alice#1234 ┃ Bob#0┃"} {
		if !strings.Contains(embed.Description, want) {
			t.Fatalf("description missing %q: %q", want, embed.Description)
		}
	}
	file := responder.Edits[0].Files[0]
	if file.Name != signInListFileName || file.ContentType != signInListFileType {
		t.Fatalf("file = %#v", file)
	}
	if string(file.Data) != "Alice#1234(id:user-1)\nBob#0(id:user-3)" {
		t.Fatalf("file data = %q", string(file.Data))
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != SignInListCommandName || usage.Events[0].Feature != "economy-signin" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestSignInListHandlerRendersRollingTimesAndMissingUsers(t *testing.T) {
	now := time.Unix(1720063800, 0)
	repo := fakemongo.NewEconomyRepository()
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", ResetMarker: 7200})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Today: 1720060200})
	discordInfo := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{}}
	module := NewSignInOnlyModule(repo, discordInfo, fixedClock{now: now}, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()

	if err := module.SignInListHandler()(context.Background(), fakediscord.SlashInteraction(SignInListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	embed := responder.Edits[0].Embeds[0]
	if !strings.Contains(embed.Description, "**目前共有**`1`**人已經簽到**") || !strings.Contains(embed.Description, "**您是否有簽到:**`沒有`") || !strings.Contains(embed.Description, disappearedUserLabel) {
		t.Fatalf("description = %q", embed.Description)
	}
	fileData := string(responder.Edits[0].Files[0].Data)
	if !strings.Contains(fileData, disappearedUserLabel+"(id:user-2)簽到時間:2024/07/04\u200910:30:00 [台北標準時間]") {
		t.Fatalf("file data = %q", fileData)
	}
}

func TestLegacySignListTimePreservesFractionalNegativeEpoch(t *testing.T) {
	if got := legacySignListTime(-0.5); got != "1970/01/01\u200907:59:59 [台北標準時間]" {
		t.Fatalf("time = %q", got)
	}
}

func TestSignInListHandlerMatchesActorByExactUserID(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "prefix-user-1", Today: 1})
	discordInfo := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{
		"prefix-user-1": {ID: "prefix-user-1", Username: "DifferentUser"},
	}}
	module := NewSignInOnlyModule(repo, discordInfo, fixedClock{now: time.Unix(10_000, 0)}, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(SignInListCommandName)
	interaction.Actor.UserID = "user-1"

	if err := module.SignInListHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "**您是否有簽到:**`沒有`") {
		t.Fatalf("description = %q", responder.Edits[0].Embeds[0].Description)
	}
}

func TestSignInListHandlerUsesOverflowTextForLargeLists(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	users := map[string]ports.DiscordUserInfo{}
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user-%03d", i)
		repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: userID, Today: 1})
		users[userID] = ports.DiscordUserInfo{ID: userID, Username: "User" + fmt.Sprintf("%03d", i)}
	}
	module := NewSignInOnlyModule(repo, &fakebotinfo.DiscordInfoProvider{Users: users}, fixedClock{now: time.Unix(10_000, 0)}, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()

	if err := module.SignInListHandler()(context.Background(), fakediscord.SlashInteraction(SignInListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, signInListOverflowNotice) {
		t.Fatalf("description = %q", responder.Edits[0].Embeds[0].Description)
	}
	if lines := strings.Count(string(responder.Edits[0].Files[0].Data), "\n") + 1; lines != 100 {
		t.Fatalf("file line count = %d", lines)
	}
}
