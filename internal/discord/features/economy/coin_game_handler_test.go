package economy

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinGameHigherLowerAcceptSettlesPot(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})
	module.gameRandInt = fixedCoinGameRandom(90, 10)
	module.color = func() int { return 0x123456 }

	start := coinGameSlash(domain.CoinGameKindHigherLower, "10")
	responder := fakediscord.NewResponder()
	if err := module.CoinGameHandler()(context.Background(), start, responder); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("start response defers=%#v follow=%#v", responder.Defers, responder.Follow)
	}
	if responder.Follow[0].Components[0].Components[0].CustomID != "yesssss" {
		t.Fatalf("invite buttons = %#v", responder.Follow[0].Components)
	}

	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-2", Username: "Opponent", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder = fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("accept game: %v", err)
	}
	challenger, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get challenger balance: %v", err)
	}
	opponent, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if err != nil {
		t.Fatalf("get opponent balance: %v", err)
	}
	if challenger.Coins != 60 || opponent.Coins != 40 {
		t.Fatalf("balances challenger=%#v opponent=%#v", challenger, opponent)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "比大小結果") {
		t.Fatalf("accept update = %#v", responder.Updates)
	}
}

func TestCoinGameChallengerAcceptDoesNotDebitPlayers(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})

	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("start game: %v", err)
	}
	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-1", Username: "User", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("challenger accept: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "你不是被邀請者") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
	challenger, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	opponent, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if challenger.Coins != 50 || opponent.Coins != 50 {
		t.Fatalf("balances mutated challenger=%#v opponent=%#v", challenger, opponent)
	}
}

func TestCoinGameRejectsOpponentWithoutEnoughCoins(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 1})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})

	responder := fakediscord.NewResponder()
	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), responder); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "對方沒有這麼多代幣") {
		t.Fatalf("insufficient response = %#v", responder.Edits)
	}
}

func coinGameSlash(kind domain.CoinGameKind, wager string) interactions.Interaction {
	return fakediscord.SlashInteractionWithOptions(CoinGameCommandName, string(kind), map[string]string{
		coinGameOptionOpponent: "user-2",
		coinGameOptionWager:    wager,
	})
}

func fixedCoinGameRandom(values ...int) func(int) int {
	index := 0
	return func(max int) int {
		if len(values) == 0 {
			return 0
		}
		value := values[index%len(values)]
		index++
		if max <= 0 {
			return value
		}
		return value % max
	}
}
