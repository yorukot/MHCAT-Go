package xp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

type orderedRankRepository struct {
	text  []domain.XPProfile
	voice []domain.XPProfile
}

func (r orderedRankRepository) ListTextXPProfiles(context.Context, string) ([]domain.XPProfile, error) {
	return append([]domain.XPProfile(nil), r.text...), nil
}

func (r orderedRankRepository) ListVoiceXPProfiles(context.Context, string) ([]domain.XPProfile, error) {
	return append([]domain.XPProfile(nil), r.voice...), nil
}

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

func TestRankServiceReversesLegacySourceOrderForTies(t *testing.T) {
	repo := orderedRankRepository{text: []domain.XPProfile{
		{GuildID: "guild", UserID: "first", Level: 1, XP: 10},
		{GuildID: "guild", UserID: "second", Level: 1, XP: 10},
		{GuildID: "guild", UserID: "leader", Level: 2, XP: 500},
	}}
	page, err := (RankService{Repository: repo}).Query(context.Background(), RankQuery{
		GuildID:  "guild",
		ViewerID: "first",
		Kind:     RankKindText,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	users := make([]string, 0, len(page.Entries))
	for _, entry := range page.Entries {
		users = append(users, entry.Profile.UserID)
	}
	if want := []string{"leader", "second", "first"}; !reflect.DeepEqual(users, want) {
		t.Fatalf("users = %#v, want %#v", users, want)
	}
}

func TestRankServiceKeepsHugeOutOfRangePageEmpty(t *testing.T) {
	repo := orderedRankRepository{text: []domain.XPProfile{{GuildID: "guild", UserID: "viewer", XP: 1}}}
	pageIndex := int(^uint(0) >> 1)
	page, err := (RankService{Repository: repo}).Query(context.Background(), RankQuery{
		GuildID:  "guild",
		ViewerID: "viewer",
		Kind:     RankKindText,
		Page:     pageIndex,
	})
	if err != nil {
		t.Fatalf("rank query: %v", err)
	}
	if page.Page != pageIndex || len(page.Entries) != 0 || page.TotalPages != 1 {
		t.Fatalf("page = %#v", page)
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
		1150:          "1.1K",
		1250:          "1.3K",
		1_000_000:     "1M",
		1_250_000:     "1.3M",
		2_500_000:     "2.5M",
		1_000_000_000: "1G",
		1_250_000_000: "1.3G",
	}
	for value, want := range tests {
		if got := LegacyRankAmount(value); got != want {
			t.Fatalf("LegacyRankAmount(%d) = %q want %q", value, got, want)
		}
	}
}
