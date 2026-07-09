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

func TestPrizeCreateStoresPrize(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	prize := domain.GachaPrizeConfig{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "code-1",
		Chance:     12.5,
		AutoDelete: true,
		Count:      3,
		GiveCoin:   7,
	}
	if err := (PrizeCreateService{Repository: repo}).Create(context.Background(), prize); err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Name != "大獎" {
		t.Fatalf("prizes = %#v", repo.Prizes["guild-1"])
	}
	if len(repo.PrizeConfigs["guild-1"]) != 1 || repo.PrizeConfigs["guild-1"][0].Code != "code-1" || !repo.PrizeConfigs["guild-1"][0].AutoDelete {
		t.Fatalf("prize configs = %#v", repo.PrizeConfigs["guild-1"])
	}
}

func TestPrizeCreateRejectsFullPool(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	for i := 0; i < 25; i++ {
		repo.Prizes["guild-1"] = append(repo.Prizes["guild-1"], domain.GachaPrize{GuildID: "guild-1", Name: "獎品"})
	}
	err := (PrizeCreateService{Repository: repo}).Create(context.Background(), domain.GachaPrizeConfig{GuildID: "guild-1", Name: "新獎品", Count: 1})
	if !errors.Is(err, ports.ErrGachaPrizePoolFull) {
		t.Fatalf("expected ErrGachaPrizePoolFull, got %v", err)
	}
}

func TestPrizeCreateRejectsDuplicatePrize(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Count: 1}}
	err := (PrizeCreateService{Repository: repo}).Create(context.Background(), domain.GachaPrizeConfig{GuildID: "guild-1", Name: "大獎", Count: 1})
	if !errors.Is(err, ports.ErrGachaPrizeExists) {
		t.Fatalf("expected ErrGachaPrizeExists, got %v", err)
	}
}

func TestPrizeCreateRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		prize domain.GachaPrizeConfig
		repo  ports.GachaPrizeCreateRepository
	}{
		{prize: domain.GachaPrizeConfig{GuildID: "", Name: "大獎", Count: 1}, repo: fakemongo.NewGachaRepository()},
		{prize: domain.GachaPrizeConfig{GuildID: "guild-1", Name: "", Count: 1}, repo: fakemongo.NewGachaRepository()},
		{prize: domain.GachaPrizeConfig{GuildID: "guild-1", Name: "大獎", Count: 0}, repo: fakemongo.NewGachaRepository()},
		{prize: domain.GachaPrizeConfig{GuildID: "guild-1", Name: "大獎", Count: 1}, repo: nil},
	} {
		err := (PrizeCreateService{Repository: tc.repo}).Create(context.Background(), tc.prize)
		if !errors.Is(err, domain.ErrInvalidGachaPrize) {
			t.Fatalf("expected ErrInvalidGachaPrize, got %v", err)
		}
	}
}

func TestPrizeEditUpdatesPrizeWithLegacyMerge(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "old-code",
		Chance:     10,
		AutoDelete: false,
		Count:      2,
		GiveCoin:   5,
	}}
	updated, err := (PrizeEditService{Repository: repo}).Edit(context.Background(), domain.GachaPrizeEdit{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "new-code",
		Chance:     12.5,
		ChanceSet:  true,
		AutoDelete: true,
		Count:      3,
		GiveCoin:   7,
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if updated.Code != "new-code" || updated.Chance != 12.5 || !updated.AutoDelete || updated.Count != 3 || updated.GiveCoin != 7 {
		t.Fatalf("updated = %#v", updated)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Chance != 12.5 || repo.Prizes["guild-1"][0].Count != 3 {
		t.Fatalf("prizes = %#v", repo.Prizes["guild-1"])
	}
}

func TestPrizeEditPreservesLegacyFalseyOldFields(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "old-code",
		Chance:     10,
		AutoDelete: false,
		Count:      2,
		GiveCoin:   5,
	}}
	updated, err := (PrizeEditService{Repository: repo}).Edit(context.Background(), domain.GachaPrizeEdit{
		GuildID:    "guild-1",
		Name:       "大獎",
		Chance:     0,
		ChanceSet:  true,
		AutoDelete: false,
		Count:      1,
		GiveCoin:   0,
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if updated.Code != "old-code" || updated.Chance != 10 || updated.AutoDelete || updated.Count != 1 || updated.GiveCoin != 5 {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestPrizeEditMissingPrize(t *testing.T) {
	_, err := (PrizeEditService{Repository: fakemongo.NewGachaRepository()}).Edit(context.Background(), domain.GachaPrizeEdit{GuildID: "guild-1", Name: "不存在", Count: 1})
	if !errors.Is(err, ports.ErrGachaPrizeMissing) {
		t.Fatalf("expected ErrGachaPrizeMissing, got %v", err)
	}
}

func TestPrizeEditRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		edit domain.GachaPrizeEdit
		repo ports.GachaPrizeEditRepository
	}{
		{edit: domain.GachaPrizeEdit{GuildID: "", Name: "大獎", Count: 1}, repo: fakemongo.NewGachaRepository()},
		{edit: domain.GachaPrizeEdit{GuildID: "guild-1", Name: "", Count: 1}, repo: fakemongo.NewGachaRepository()},
		{edit: domain.GachaPrizeEdit{GuildID: "guild-1", Name: "大獎", Count: 0}, repo: fakemongo.NewGachaRepository()},
		{edit: domain.GachaPrizeEdit{GuildID: "guild-1", Name: "大獎", Count: 1}, repo: nil},
	} {
		_, err := (PrizeEditService{Repository: tc.repo}).Edit(context.Background(), tc.edit)
		if !errors.Is(err, domain.ErrInvalidGachaPrize) {
			t.Fatalf("expected ErrInvalidGachaPrize, got %v", err)
		}
	}
}
