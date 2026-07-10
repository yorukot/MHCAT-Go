package gacha

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLegacyDrawCounts(t *testing.T) {
	cases := []struct {
		choice     string
		paidDraws  int
		actualDraw int
	}{
		{"", 1, 1},
		{"5", 5, 5},
		{"11", 10, 11},
		{"17", 15, 17},
		{"23", 20, 23},
	}
	for _, tc := range cases {
		paid, actual, err := LegacyDrawCounts(tc.choice)
		if err != nil {
			t.Fatalf("LegacyDrawCounts(%q): %v", tc.choice, err)
		}
		if paid != tc.paidDraws || actual != tc.actualDraw {
			t.Fatalf("LegacyDrawCounts(%q) = %d/%d", tc.choice, paid, actual)
		}
	}
	if _, _, err := LegacyDrawCounts("bad"); !errors.Is(err, domain.ErrInvalidGachaDraw) {
		t.Fatalf("invalid choice error = %v", err)
	}
}

func TestDrawServiceUsesLegacyCountsAndAppliesDraw(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Balances["guild-1/user-1"] = domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 6000}
	repo.Configs["guild-1"] = domain.EconomyConfig{GuildID: "guild-1", GachaCost: 100, ChannelID: "notify"}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "code-1",
		Chance:     100,
		AutoDelete: true,
		Count:      12,
		GiveCoin:   3,
	}}
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 100, Count: 12}}
	service := DrawService{Repository: repo, Random: func() float64 { return 0 }}

	result, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1", Choice: "11"})
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if result.PaidDraws != 10 || result.ActualDraws != 11 || result.Cost != 1000 {
		t.Fatalf("draw counts/cost = %#v", result)
	}
	if len(result.Prizes) != 11 || result.Prizes[0].Name != "大獎" || result.Prizes[0].Code != "code-1" {
		t.Fatalf("prizes = %#v", result.Prizes)
	}
	if result.BalanceBefore != 6000 || result.BalanceAfter != 5033 {
		t.Fatalf("balances = before %d after %d", result.BalanceBefore, result.BalanceAfter)
	}
	if repo.Balances["guild-1/user-1"].Coins != 5033 {
		t.Fatalf("stored balance = %#v", repo.Balances["guild-1/user-1"])
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Count != 1 {
		t.Fatalf("remaining inventory = %#v", repo.Prizes["guild-1"])
	}
}

func TestDrawServiceReturnsAirAndKeepsInventory(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Balances["guild-1/user-1"] = domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1000}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{GuildID: "guild-1", Name: "大獎", Chance: 10, AutoDelete: true, Count: 1}}
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1}}
	service := DrawService{Repository: repo, Random: func() float64 { return 0.99 }}

	result, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1"})
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if len(result.Prizes) != 1 || !result.Prizes[0].Air || result.Prizes[0].Name != domain.GachaAirPrizeName {
		t.Fatalf("prizes = %#v", result.Prizes)
	}
	if repo.Balances["guild-1/user-1"].Coins != 500 {
		t.Fatalf("stored balance = %#v", repo.Balances["guild-1/user-1"])
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Count != 1 {
		t.Fatalf("inventory should be unchanged for air = %#v", repo.Prizes["guild-1"])
	}
}

func TestDrawServiceReloadsPoolAfterEachPrizeDepletion(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Balances["guild-1/user-1"] = domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1000}
	repo.Configs["guild-1"] = domain.EconomyConfig{GuildID: "guild-1", GachaCost: 100}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "限量獎品",
		Chance:     100,
		AutoDelete: true,
		Count:      1,
		GiveCoin:   25,
	}}
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "限量獎品", Chance: 100, Count: 1}}
	service := DrawService{Repository: repo, Random: func() float64 { return 0 }}

	result, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1", Choice: "5"})
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if len(result.Prizes) != 5 || result.Prizes[0].Name != "限量獎品" || result.Prizes[0].Air {
		t.Fatalf("prizes = %#v", result.Prizes)
	}
	for _, prize := range result.Prizes[1:] {
		if !prize.Air {
			t.Fatalf("depleted prize was drawn again: %#v", result.Prizes)
		}
	}
	if len(repo.PrizeConfigs["guild-1"]) != 0 || len(repo.Prizes["guild-1"]) != 0 {
		t.Fatalf("depleted inventory remains: configs=%#v prizes=%#v", repo.PrizeConfigs["guild-1"], repo.Prizes["guild-1"])
	}
	if result.BalanceAfter != 525 || repo.Balances["guild-1/user-1"].Coins != 525 {
		t.Fatalf("coin reward should apply once: result=%#v stored=%#v", result, repo.Balances["guild-1/user-1"])
	}
}

func TestDrawServiceValidationAndRepositoryErrors(t *testing.T) {
	if _, err := (DrawService{Repository: fakemongo.NewGachaRepository()}).Draw(context.Background(), domain.GachaDrawCommand{}); !errors.Is(err, domain.ErrInvalidGachaDraw) {
		t.Fatalf("missing command error = %v", err)
	}
	repo := fakemongo.NewGachaRepository()
	service := DrawService{Repository: repo, Random: func() float64 { return 0 }}
	if _, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1"}); !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("missing balance error = %v", err)
	}
	repo.Balances["guild-1/user-1"] = domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1000}
	if _, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1"}); !errors.Is(err, ports.ErrGachaPrizePoolEmpty) {
		t.Fatalf("empty pool error = %v", err)
	}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{GuildID: "guild-1", Name: "大獎", Chance: 100, Count: 1}}
	repo.Configs["guild-1"] = domain.EconomyConfig{GuildID: "guild-1", GachaCost: 2000}
	if _, err := service.Draw(context.Background(), domain.GachaDrawCommand{GuildID: "guild-1", UserID: "user-1"}); !errors.Is(err, ports.ErrGachaInsufficientCoins) {
		t.Fatalf("insufficient coins error = %v", err)
	}
}
