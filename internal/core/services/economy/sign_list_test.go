package economy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSignInListUsesDailyTodayMarkerWhenConfigMissing(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user-1", Today: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user-2", Today: 0})
	result, err := (SignInListService{Repository: repo}).List(context.Background(), "guild", "user-1", time.Unix(1000, 0))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.RollingWindow {
		t.Fatalf("expected daily marker mode: %#v", result)
	}
	if len(result.Entries) != 1 || result.Entries[0].UserID != "user-1" || result.Entries[0].ShowSignedAt {
		t.Fatalf("entries = %#v", result.Entries)
	}
}

func TestSignInListUsesRollingCooldownWindow(t *testing.T) {
	now := time.Unix(10_000, 0)
	repo := fakemongo.NewEconomyRepository()
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild", ResetMarker: 3600})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "recent", Today: now.Add(-30 * time.Minute).Unix()})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "old", Today: now.Add(-2 * time.Hour).Unix()})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "future", Today: now.Add(time.Minute).Unix()})
	result, err := (SignInListService{Repository: repo}).List(context.Background(), "guild", "recent", now)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !result.RollingWindow {
		t.Fatalf("expected rolling mode: %#v", result)
	}
	if len(result.Entries) != 1 || result.Entries[0].UserID != "recent" || !result.Entries[0].ShowSignedAt {
		t.Fatalf("entries = %#v", result.Entries)
	}
}

func TestLegacySignListWindowPreservesMongooseTimeSemantics(t *testing.T) {
	tests := []struct {
		name        string
		found       bool
		config      domain.EconomyConfig
		wantRolling bool
		wantWindow  float64
	}{
		{name: "missing config", found: false},
		{name: "numeric zero", found: true, config: domain.EconomyConfig{ResetMarkerText: "0"}},
		{name: "missing field", found: true, config: domain.EconomyConfig{ResetMarkerText: "undefined"}, wantRolling: true, wantWindow: 86400},
		{name: "null field", found: true, config: domain.EconomyConfig{ResetMarkerText: "null"}, wantRolling: true, wantWindow: 86400},
		{name: "nan field", found: true, config: domain.EconomyConfig{ResetMarkerText: "NaN"}, wantRolling: true, wantWindow: 86400},
		{name: "fractional field", found: true, config: domain.EconomyConfig{ResetMarkerText: "3600.5"}, wantRolling: true, wantWindow: 3600.5},
		{name: "negative field", found: true, config: domain.EconomyConfig{ResetMarkerText: "-1"}, wantRolling: true, wantWindow: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rolling, window := legacySignListWindow(test.config, test.found)
			if rolling != test.wantRolling || window != test.wantWindow {
				t.Fatalf("window = (%v, %v), want (%v, %v)", rolling, window, test.wantRolling, test.wantWindow)
			}
		})
	}
}

func TestSignInListAppliesLegacyMissingAndNegativeTimeWindows(t *testing.T) {
	now := time.Unix(100_000, 0)
	tests := []struct {
		name        string
		markerText  string
		wantEntries int
	}{
		{name: "missing field uses one day", markerText: "undefined", wantEntries: 1},
		{name: "negative field matches nobody", markerText: "-1", wantEntries: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyRepository()
			repo.PutConfig(domain.EconomyConfig{GuildID: "guild", ResetMarkerText: test.markerText})
			repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "recent", Today: now.Add(-time.Hour).Unix()})
			result, err := (SignInListService{Repository: repo}).List(context.Background(), "guild", "recent", now)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if !result.RollingWindow || len(result.Entries) != test.wantEntries {
				t.Fatalf("result = %#v", result)
			}
			if len(result.Entries) == 1 && !result.Entries[0].ShowSignedAt {
				t.Fatalf("rolling entry omitted timestamp: %#v", result.Entries[0])
			}
		})
	}
}

func TestSignInListRequiresRepositoryIDsAndNow(t *testing.T) {
	_, err := (SignInListService{}).List(context.Background(), "guild", "user", time.Unix(1, 0))
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("nil repo error = %v", err)
	}
	repo := fakemongo.NewEconomyRepository()
	_, err = (SignInListService{Repository: repo}).List(context.Background(), "", "user", time.Unix(1, 0))
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("missing guild error = %v", err)
	}
	_, err = (SignInListService{Repository: repo}).List(context.Background(), "guild", "", time.Unix(1, 0))
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("missing user error = %v", err)
	}
	_, err = (SignInListService{Repository: repo}).List(context.Background(), "guild", "user", time.Time{})
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("zero time error = %v", err)
	}
}
