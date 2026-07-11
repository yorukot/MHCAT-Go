package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinGameReserveAndSettlePot(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "main", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 50})
	service := CoinGameService{Repository: repo}

	_, err := service.Reserve(context.Background(), domain.CoinGameCommand{
		GuildID:      "guild",
		ChallengerID: "main",
		OpponentID:   "user",
		Wager:        10,
		Kind:         domain.CoinGameKindHigherLower,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	_, err = service.Settle(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild",
		ChallengerID:     "main",
		OpponentID:       "user",
		ChallengerReturn: 20,
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	main, _ := repo.GetCoinBalance(context.Background(), "guild", "main")
	user, _ := repo.GetCoinBalance(context.Background(), "guild", "user")
	if main.Coins != 60 || user.Coins != 40 {
		t.Fatalf("balances main=%#v user=%#v", main, user)
	}
}

func TestCoinGamePreservesLegacyNegativeOneWagerArithmetic(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "main", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 50})
	service := CoinGameService{Repository: repo}

	reserved, err := service.Reserve(context.Background(), domain.CoinGameCommand{
		GuildID:      "guild",
		ChallengerID: "main",
		OpponentID:   "user",
		Wager:        -1,
		Kind:         domain.CoinGameKindHigherLower,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if reserved.Challenger.Coins != 51 || reserved.Opponent.Coins != 51 {
		t.Fatalf("reserved = %#v", reserved)
	}

	settled, err := service.Settle(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild",
		ChallengerID:     "main",
		OpponentID:       "user",
		ChallengerReturn: -2,
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if settled.Challenger.Coins != 49 || settled.Opponent.Coins != 51 {
		t.Fatalf("settled = %#v", settled)
	}
}

func TestCoinGameRejectsSettlementBelowLegacyNegativeOneOutput(t *testing.T) {
	_, err := (CoinGameService{Repository: fakemongo.NewEconomyRepository()}).Settle(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild",
		ChallengerID:     "main",
		OpponentID:       "user",
		ChallengerReturn: -3,
	})
	if !errors.Is(err, domain.ErrInvalidCoinGameCommand) {
		t.Fatalf("expected invalid settlement, got %v", err)
	}
}
