package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinAdminAddCreatesMissingBalance(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	result, err := (CoinAdminService{Repository: repo}).Adjust(context.Background(), domain.CoinAdminCommand{
		GuildID:   " guild ",
		UserID:    " user ",
		Operation: domain.CoinAdminOperationAdd,
		Amount:    25,
	})
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if !result.Created || result.Balance.GuildID != "guild" || result.Balance.UserID != "user" || result.Balance.Coins != 25 || result.Delta != 25 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCoinAdminReduceRejectsNegativeBalance(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 3})
	_, err := (CoinAdminService{Repository: repo}).Adjust(context.Background(), domain.CoinAdminCommand{
		GuildID:   "guild",
		UserID:    "user",
		Operation: domain.CoinAdminOperationReduce,
		Amount:    4,
	})
	if !errors.Is(err, ports.ErrCoinNegativeBalance) {
		t.Fatalf("expected negative balance error, got %v", err)
	}
}

func TestCoinAdminAddRejectsLimitExceeded(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: MaxLegacyCoinBalance})
	_, err := (CoinAdminService{Repository: repo}).Adjust(context.Background(), domain.CoinAdminCommand{
		GuildID:   "guild",
		UserID:    "user",
		Operation: domain.CoinAdminOperationAdd,
		Amount:    1,
	})
	if !errors.Is(err, ports.ErrCoinLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestCoinAdminRejectsInvalidCommand(t *testing.T) {
	_, err := (CoinAdminService{Repository: fakemongo.NewEconomyRepository()}).Adjust(context.Background(), domain.CoinAdminCommand{
		GuildID:   "guild",
		UserID:    "user",
		Operation: "multiply",
		Amount:    1,
	})
	if !errors.Is(err, domain.ErrInvalidCoinAdminCommand) {
		t.Fatalf("expected invalid command, got %v", err)
	}
}
