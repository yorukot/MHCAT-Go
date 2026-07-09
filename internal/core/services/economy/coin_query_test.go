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
	if result.GachaCost != 700 || result.MissingCoins != 575 {
		t.Fatalf("unexpected result: %#v", result)
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
	if result.GachaCost != DefaultGachaCost || result.MissingCoins != 375 {
		t.Fatalf("unexpected default result: %#v", result)
	}
}

func TestCoinQueryMissingBalanceReturnsNotFound(t *testing.T) {
	_, err := (CoinQueryService{Repository: fakemongo.NewEconomyRepository()}).Query(context.Background(), "guild", "user")
	if !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("expected balance not found, got %v", err)
	}
}
