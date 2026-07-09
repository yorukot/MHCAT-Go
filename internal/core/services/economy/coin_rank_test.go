package economy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinRankSortsPagesAndComputesLegacyTieRank(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "a", Coins: 10})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "viewer", Coins: 30})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "b", Coins: 30})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "other", Coins: 100})

	page, err := (CoinRankService{Repository: repo}).Query(context.Background(), CoinRankQuery{
		GuildID:  " guild ",
		ViewerID: " viewer ",
		Page:     0,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	var users []string
	for _, entry := range page.Entries {
		users = append(users, entry.Balance.UserID)
	}
	if want := []string{"other", "viewer", "b", "a"}; !reflect.DeepEqual(users, want) {
		t.Fatalf("users = %#v want %#v", users, want)
	}
	if page.ViewerRank != 3 || !page.ViewerHasBalance {
		t.Fatalf("viewer rank = %d has=%v", page.ViewerRank, page.ViewerHasBalance)
	}
	if page.TotalPages != 1 || page.TotalEntries != 4 {
		t.Fatalf("page totals = %#v", page)
	}
}

func TestCoinRankPaginatesTenRows(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	for i := int64(0); i < 12; i++ {
		repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: string(rune('a' + i)), Coins: i})
	}
	page, err := (CoinRankService{Repository: repo}).Query(context.Background(), CoinRankQuery{
		GuildID:  "guild",
		ViewerID: "a",
		Page:     1,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	if len(page.Entries) != 2 || page.Entries[0].Rank != 11 || page.Entries[0].Balance.Coins != 1 {
		t.Fatalf("page entries = %#v", page.Entries)
	}
	if page.TotalPages != 2 {
		t.Fatalf("total pages = %d", page.TotalPages)
	}
}

func TestCoinRankRejectsInvalidQuery(t *testing.T) {
	_, err := (CoinRankService{Repository: fakemongo.NewEconomyRepository()}).Query(context.Background(), CoinRankQuery{GuildID: "guild"})
	if !errors.Is(err, domain.ErrInvalidCoinRankQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
}

func TestLegacyCoinRankAmount(t *testing.T) {
	tests := map[int64]string{
		999:           "999",
		1000:          "1K",
		1250:          "1.2K",
		1_000_000:     "1M",
		2_500_000:     "2.5M",
		1_000_000_000: "1G",
	}
	for value, want := range tests {
		if got := LegacyCoinRankAmount(value); got != want {
			t.Fatalf("LegacyCoinRankAmount(%d) = %q want %q", value, got, want)
		}
	}
}
