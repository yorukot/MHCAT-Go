package redeem

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestServiceRedeemsValidCode(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	repo := fakemongo.NewRedeemRepository()
	repo.Codes[" abc "] = domain.RedeemCode{Code: " abc ", Price: 12.5, CreatedAtMillis: now.Add(-time.Hour).UnixMilli()}
	service := NewService(repo, fixedClock{now: now})

	if err := service.Redeem(context.Background(), " guild-1 ", " abc "); err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if _, ok := repo.Codes[" abc "]; ok {
		t.Fatal("expected code to be consumed")
	}
	if got := repo.Balances["guild-1"]; got != 12.5 {
		t.Fatalf("balance = %v", got)
	}
}

func TestServicePreservesAllSpaceCode(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	repo := fakemongo.NewRedeemRepository()
	repo.Codes["   "] = domain.RedeemCode{Code: "   ", Price: 1, CreatedAtMillis: now.UnixMilli()}
	service := NewService(repo, fixedClock{now: now})

	if err := service.Redeem(context.Background(), "guild-1", "   "); err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if _, ok := repo.Codes["   "]; ok || repo.Balances["guild-1"] != 1 {
		t.Fatalf("codes=%#v balances=%#v", repo.Codes, repo.Balances)
	}
}

func TestServiceRejectsExpiredCodeWithoutConsuming(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	repo := fakemongo.NewRedeemRepository()
	repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: 12.5, CreatedAtMillis: now.Add(-LegacyCodeTTL - time.Millisecond).UnixMilli()}
	service := NewService(repo, fixedClock{now: now})

	err := service.Redeem(context.Background(), "guild-1", "abc")
	if !errors.Is(err, ports.ErrRedeemCodeExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
	if _, ok := repo.Codes["abc"]; !ok {
		t.Fatal("expired legacy code should not be consumed")
	}
}

func TestServiceRejectsInvalidInputs(t *testing.T) {
	service := NewService(fakemongo.NewRedeemRepository(), fixedClock{now: time.UnixMilli(1700000000000)})
	if err := service.Redeem(context.Background(), "", "abc"); !errors.Is(err, domain.ErrInvalidRedeemCode) {
		t.Fatalf("expected invalid guild, got %v", err)
	}
	if err := service.Redeem(context.Background(), "guild", ""); !errors.Is(err, domain.ErrInvalidRedeemCode) {
		t.Fatalf("expected invalid code, got %v", err)
	}
	if err := (Service{}).Redeem(context.Background(), "guild", "abc"); !errors.Is(err, domain.ErrInvalidRedeemCode) {
		t.Fatalf("expected missing repo to be invalid, got %v", err)
	}
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

var _ ports.Clock = fixedClock{}
