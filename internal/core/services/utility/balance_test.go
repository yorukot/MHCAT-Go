package utility

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestBalanceServiceReturnsStoredBalance(t *testing.T) {
	repo := fakemongo.NewBalanceRepository()
	repo.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "12.5"}

	balance, err := (BalanceService{Repository: repo}).Get(context.Background(), " guild-1 ")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if balance.GuildID != "guild-1" || balance.Amount != "12.5" {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestBalanceServiceDefaultsMissingBalanceToZero(t *testing.T) {
	balance, err := (BalanceService{Repository: fakemongo.NewBalanceRepository()}).Get(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if balance.Amount != "0" {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestBalanceServiceRejectsInvalidInputs(t *testing.T) {
	_, err := (BalanceService{}).Get(context.Background(), "guild-1")
	if !errors.Is(err, domain.ErrInvalidBalanceQuery) {
		t.Fatalf("nil repo error = %v", err)
	}
	_, err = (BalanceService{Repository: fakemongo.NewBalanceRepository()}).Get(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidBalanceQuery) {
		t.Fatalf("missing guild error = %v", err)
	}
}
