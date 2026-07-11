package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinQueryUsesConfiguredGachaCost(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 125})
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild", GachaCost: 700})
	result, err := (CoinQueryService{Repository: repo}).Query(context.Background(), "guild", "user")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !result.ConfigFound {
		t.Fatal("expected config to be found")
	}
	if result.GachaCost != 700 || result.GachaCostText != "700" || result.BalanceText != "125" || result.MissingCoins != 575 || result.MissingCoinsText != "575" || result.CanGacha {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCoinQueryPreservesLegacyBalanceNumberEdges(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantMissing string
		wantCan     bool
	}{
		{name: "undefined", text: "undefined", wantCan: true},
		{name: "null", text: "null", wantMissing: "500"},
		{name: "decimal", text: "125.5", wantMissing: "374.5"},
		{name: "positive infinity", text: "Infinity", wantCan: true},
		{name: "negative infinity", text: "-Infinity", wantMissing: "Infinity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyRepository()
			repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", CoinsText: test.text})
			repo.PutConfig(domain.EconomyConfig{GuildID: "guild", GachaCost: 500})
			result, err := (CoinQueryService{Repository: repo}).Query(context.Background(), "guild", "user")
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if result.BalanceText != test.text || result.MissingCoinsText != test.wantMissing || result.CanGacha != test.wantCan {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestCoinQueryMissingConfigDefaultsToLegacyGachaCost(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 125})
	result, err := (CoinQueryService{Repository: repo}).Query(context.Background(), "guild", "user")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if result.ConfigFound {
		t.Fatal("expected config to be missing")
	}
	if result.GachaCost != DefaultGachaCost || result.GachaCostText != "500" || result.MissingCoins != 375 || result.MissingCoinsText != "375" || result.CanGacha {
		t.Fatalf("unexpected default result: %#v", result)
	}
}

func TestCoinQueryPreservesLegacyConfiguredNumberEdges(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantMissing string
		wantCan     bool
	}{
		{name: "undefined", text: "undefined", wantCan: true},
		{name: "null", text: "null", wantCan: true},
		{name: "decimal", text: "700.5", wantMissing: "575.5"},
		{name: "infinity", text: "Infinity", wantMissing: "Infinity"},
		{name: "negative infinity", text: "-Infinity", wantCan: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyRepository()
			repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 125})
			repo.PutConfig(domain.EconomyConfig{GuildID: "guild", GachaCostText: test.text})
			result, err := (CoinQueryService{Repository: repo}).Query(context.Background(), "guild", "user")
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if result.GachaCostText != test.text || result.MissingCoinsText != test.wantMissing || result.CanGacha != test.wantCan {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestCoinQueryMissingBalanceReturnsNotFound(t *testing.T) {
	_, err := (CoinQueryService{Repository: fakemongo.NewEconomyRepository()}).Query(context.Background(), "guild", "user")
	if !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("expected balance not found, got %v", err)
	}
}
