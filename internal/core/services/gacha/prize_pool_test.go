package gacha

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestPrizePoolQueryUsesLegacyDefaultsWhenConfigMissing(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "A", Chance: 10, Count: 2}}
	result, err := (PrizePoolService{Repository: repo}).Query(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if result.ConfigFound {
		t.Fatal("expected missing config")
	}
	if result.Config.GachaCost != DefaultGachaCost || result.Config.SignCoins != DefaultSignCoins || result.Config.XPMultiple != DefaultXPMultiple {
		t.Fatalf("config = %#v", result.Config)
	}
}

func TestPrizePoolQueryReturnsEmptyError(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	_, err := (PrizePoolService{Repository: repo}).Query(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrGachaPrizePoolEmpty) {
		t.Fatalf("expected ErrGachaPrizePoolEmpty, got %v", err)
	}
}

func TestPrizePoolQueryRejectsInvalidInput(t *testing.T) {
	_, err := (PrizePoolService{Repository: fakemongo.NewGachaRepository()}).Query(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidGachaQuery) {
		t.Fatalf("expected ErrInvalidGachaQuery, got %v", err)
	}
}

func TestPrizeDeleteRemovesPrize(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{
		{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1},
		{GuildID: "guild-1", Name: "小獎", Chance: 90, Count: 9},
	}
	prize, err := (PrizeDeleteService{Repository: repo}).Delete(context.Background(), "guild-1", "大獎")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if prize.Name != "大獎" {
		t.Fatalf("deleted prize = %#v", prize)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Name != "小獎" {
		t.Fatalf("remaining prizes = %#v", repo.Prizes["guild-1"])
	}
}

func TestPrizeDeleteMissingPrize(t *testing.T) {
	_, err := (PrizeDeleteService{Repository: fakemongo.NewGachaRepository()}).Delete(context.Background(), "guild-1", "不存在")
	if !errors.Is(err, ports.ErrGachaPrizeMissing) {
		t.Fatalf("expected ErrGachaPrizeMissing, got %v", err)
	}
}

func TestPrizeDeleteRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		guildID   string
		prizeName string
		repo      ports.GachaPrizeDeleteRepository
	}{
		{guildID: "", prizeName: "大獎", repo: fakemongo.NewGachaRepository()},
		{guildID: "guild-1", prizeName: "", repo: fakemongo.NewGachaRepository()},
		{guildID: "guild-1", prizeName: "大獎", repo: nil},
	} {
		_, err := (PrizeDeleteService{Repository: tc.repo}).Delete(context.Background(), tc.guildID, tc.prizeName)
		if !errors.Is(err, domain.ErrInvalidGachaQuery) {
			t.Fatalf("expected ErrInvalidGachaQuery, got %v", err)
		}
	}
}
