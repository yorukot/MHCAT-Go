package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinResetServiceDeletesGuildBalances(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 100})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 200})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-2", UserID: "user-3", Coins: 300})

	result, err := (CoinResetService{Repository: repo}).Reset(context.Background(), domain.CoinResetCommand{GuildID: " guild-1 "})
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if !result.Deleted || result.AffectedCount != 2 {
		t.Fatalf("result = %#v", result)
	}
	if _, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("user-1 should be deleted, err=%v", err)
	}
	if balance, err := repo.GetCoinBalance(context.Background(), "guild-2", "user-3"); err != nil || balance.Coins != 300 {
		t.Fatalf("other guild balance = %#v err=%v", balance, err)
	}
}

func TestCoinResetServiceDividesGuildBalancesWithLegacyRounding(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 101})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 100})

	result, err := (CoinResetService{Repository: repo}).Reset(context.Background(), domain.CoinResetCommand{GuildID: "guild-1", Divisor: 2})
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if result.Deleted || result.Divisor != 2 || result.AffectedCount != 2 {
		t.Fatalf("result = %#v", result)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); balance.Coins != 51 {
		t.Fatalf("user-1 balance = %#v", balance)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2"); balance.Coins != 50 {
		t.Fatalf("user-2 balance = %#v", balance)
	}
}

func TestCoinResetServiceDividesGuildBalancesWithLegacyNegativeRounding(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 11})

	result, err := (CoinResetService{Repository: repo}).Reset(context.Background(), domain.CoinResetCommand{GuildID: "guild-1", Divisor: -2})
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if result.Deleted || result.Divisor != -2 || result.AffectedCount != 2 {
		t.Fatalf("result = %#v", result)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); balance.Coins != 0 {
		t.Fatalf("user-1 balance = %#v", balance)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2"); balance.Coins != -5 {
		t.Fatalf("user-2 balance = %#v", balance)
	}
}

func TestCoinResetServiceRejectsInvalidQuery(t *testing.T) {
	_, err := (CoinResetService{Repository: fakemongo.NewEconomyRepository()}).Reset(context.Background(), domain.CoinResetCommand{})
	if !errors.Is(err, domain.ErrInvalidCoinResetCommand) {
		t.Fatalf("expected invalid command, got %v", err)
	}
	_, err = (CoinResetService{}).Reset(context.Background(), domain.CoinResetCommand{GuildID: "guild"})
	if !errors.Is(err, domain.ErrInvalidCoinResetCommand) {
		t.Fatalf("expected invalid command for nil repository, got %v", err)
	}
}
