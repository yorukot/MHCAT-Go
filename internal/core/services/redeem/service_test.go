package redeem

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestServiceRedeemsValidCode(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	repo := fakemongo.NewRedeemRepository()
	repo.Codes[" abc "] = domain.RedeemCode{Code: " abc ", Price: 12.5, CreatedAtMillis: float64(now.Add(-time.Hour).UnixMilli())}
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
	repo.Codes["   "] = domain.RedeemCode{Code: "   ", Price: 1, CreatedAtMillis: float64(now.UnixMilli())}
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
	repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: 12.5, CreatedAtMillis: float64(now.Add(-LegacyCodeTTL - time.Millisecond).UnixMilli())}
	service := NewService(repo, fixedClock{now: now})

	err := service.Redeem(context.Background(), "guild-1", "abc")
	if !errors.Is(err, ports.ErrRedeemCodeExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
	if _, ok := repo.Codes["abc"]; !ok {
		t.Fatal("expired legacy code should not be consumed")
	}
}

func TestServicePreservesLegacyNumberAndExpiryEdges(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	for _, test := range []struct {
		name      string
		price     float64
		createdMS float64
		wantErr   error
		want      float64
	}{
		{name: "exact seven day boundary", price: 2, createdMS: float64(now.Add(-LegacyCodeTTL).UnixMilli()), want: 12},
		{name: "missing time is NaN and remains usable", price: 2, createdMS: math.NaN(), want: 12},
		{name: "null time is zero and expired", price: 2, createdMS: 0, wantErr: ports.ErrRedeemCodeExpired, want: 10},
		{name: "negative price reduces balance", price: -2, createdMS: float64(now.UnixMilli()), want: 8},
		{name: "malformed price is rejected safely", price: math.NaN(), createdMS: float64(now.UnixMilli()), wantErr: domain.ErrInvalidRedeemCode, want: 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewRedeemRepository()
			repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: test.price, CreatedAtMillis: test.createdMS}
			repo.Balances["guild-1"] = 10
			err := NewService(repo, fixedClock{now: now}).Redeem(context.Background(), "guild-1", "abc")
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if repo.Balances["guild-1"] != test.want {
				t.Fatalf("balance = %v, want %v", repo.Balances["guild-1"], test.want)
			}
			_, remains := repo.Codes["abc"]
			if remains != (test.wantErr != nil) {
				t.Fatalf("code remains = %v, want %v", remains, test.wantErr != nil)
			}
		})
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
