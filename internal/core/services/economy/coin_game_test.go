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

func TestCoinGamePreservesLegacyBalanceScalars(t *testing.T) {
	for _, test := range []struct {
		name     string
		text     string
		wager    int64
		returned int64
		want     string
	}{
		{name: "decimal", text: "50.5", wager: 10, returned: 20, want: "60.5"},
		{name: "null", text: "null", wager: 0, returned: 0, want: "0"},
		{name: "infinity", text: "Infinity", wager: 10, returned: 20, want: "Infinity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyRepository()
			repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "main", CoinsText: test.text})
			repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", CoinsText: test.text})
			service := CoinGameService{Repository: repo}
			command := domain.CoinGameCommand{GuildID: "guild", ChallengerID: "main", OpponentID: "user", Wager: test.wager, Kind: domain.CoinGameKindHigherLower}
			if _, err := service.Reserve(context.Background(), command); err != nil {
				t.Fatalf("reserve: %v", err)
			}
			result, err := service.Settle(context.Background(), domain.CoinGameSettlementCommand{GuildID: "guild", ChallengerID: "main", OpponentID: "user", ChallengerReturn: test.returned, OpponentReturn: test.returned})
			if err != nil {
				t.Fatalf("settle: %v", err)
			}
			if result.Challenger.CoinsText != test.want || result.Opponent.CoinsText != test.want {
				t.Fatalf("settled = %#v", result)
			}
		})
	}
}
