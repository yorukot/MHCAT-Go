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
