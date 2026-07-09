package xp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRankServiceSortsPagesAndComputesLegacyViewerRank(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild", UserID: "a", Level: 1, XP: 50})
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild", UserID: "viewer", Level: 2, XP: 120})
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild", UserID: "b", Level: 2, XP: 10})
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild", UserID: "other", Level: 3, XP: 200})

	page, err := (RankService{Repository: repo}).Query(context.Background(), RankQuery{
		GuildID:  " guild ",
		ViewerID: " viewer ",
		Kind:     RankKindText,
		Page:     0,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	var users []string
	for _, entry := range page.Entries {
		users = append(users, entry.Profile.UserID)
	}
	if want := []string{"other", "viewer", "a", "b"}; !reflect.DeepEqual(users, want) {
		t.Fatalf("users = %#v want %#v", users, want)
	}
	if page.ViewerRank != 2 || !page.ViewerHasProfile {
		t.Fatalf("viewer rank = %d has=%v", page.ViewerRank, page.ViewerHasProfile)
	}
	if page.TotalPages != 1 || page.TotalEntries != 4 {
		t.Fatalf("page totals = %#v", page)
	}
}

func TestRankServicePaginatesTenRows(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	for i := int64(0); i < 12; i++ {
		_ = repo.SaveVoiceXPProfile(context.Background(), domain.XPProfile{GuildID: "guild", UserID: string(rune('a' + i)), Level: 1, XP: i})
	}
	page, err := (RankService{Repository: repo}).Query(context.Background(), RankQuery{
		GuildID:  "guild",
		ViewerID: "a",
		Kind:     RankKindVoice,
		Page:     1,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	if len(page.Entries) != 2 || page.Entries[0].Rank != 11 || page.Entries[0].Profile.XP != 1 {
		t.Fatalf("page entries = %#v", page.Entries)
	}
	if page.TotalPages != 2 {
		t.Fatalf("total pages = %d", page.TotalPages)
	}
}

func TestRankServiceRejectsInvalidQuery(t *testing.T) {
	_, err := (RankService{Repository: fakemongo.NewXPAdminRepository()}).Query(context.Background(), RankQuery{GuildID: "guild", Kind: RankKindText})
	if !errors.Is(err, domain.ErrInvalidXPRankQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
	_, err = (RankService{}).Query(context.Background(), RankQuery{GuildID: "guild", ViewerID: "user", Kind: RankKindText})
	if !errors.Is(err, domain.ErrInvalidXPRankQuery) {
		t.Fatalf("expected invalid query for nil repo, got %v", err)
	}
}

func TestLegacyRankAmount(t *testing.T) {
	tests := map[int64]string{
		999:           "999",
		1000:          "1K",
		1250:          "1.2K",
		1_000_000:     "1M",
		2_500_000:     "2.5M",
		1_000_000_000: "1G",
	}
	for value, want := range tests {
		if got := LegacyRankAmount(value); got != want {
			t.Fatalf("LegacyRankAmount(%d) = %q want %q", value, got, want)
		}
	}
}
