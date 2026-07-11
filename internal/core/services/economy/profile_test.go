package economy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestProfileServiceQueriesLegacyReadModels(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 500, Today: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 600})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-3", Coins: 500})
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", GachaCost: 1000, SignCoins: 25, XPMultiple: 1.5, ResetMarker: 3600})
	repo.PutWorkConfig(domain.WorkConfig{GuildID: "guild-1", DailyEnergy: 10, MaxEnergy: 100})
	repo.PutWorkUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", Energy: 55, EndTimeUnix: 2_000})
	repo.PutTextXP(domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 50, Level: 2})
	repo.PutTextXP(domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 10, Level: 3})
	repo.PutVoiceXP(domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 80, Level: 2})
	repo.PutVoiceXP(domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 90, Level: 3})

	result, err := (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{
		GuildID: " guild-1 ",
		UserID:  " user-1 ",
		Now:     time.Unix(1_500, 0),
	})
	if err != nil {
		t.Fatalf("query profile: %v", err)
	}
	if !result.CoinFound || result.CoinRank != 3 || result.SignStatus != "已簽到" {
		t.Fatalf("coin result = found %t rank %d sign %q", result.CoinFound, result.CoinRank, result.SignStatus)
	}
	if !result.TextXPFound || result.TextRank != 2 {
		t.Fatalf("text xp = found %t rank %d", result.TextXPFound, result.TextRank)
	}
	if !result.VoiceXPFound || result.VoiceRank != 2 {
		t.Fatalf("voice xp = found %t rank %d", result.VoiceXPFound, result.VoiceRank)
	}
	if !result.WorkConfigFound || !result.WorkUserFound || result.WorkUser.Energy != 55 {
		t.Fatalf("work data = %#v", result)
	}
}

func TestProfileServiceAllowsMissingOptionalRows(t *testing.T) {
	result, err := (ProfileService{Repository: fakemongo.NewEconomyProfileRepository()}).Query(context.Background(), ProfileQuery{
		GuildID: "guild-1",
		UserID:  "user-1",
		Now:     time.Unix(10, 0),
	})
	if err != nil {
		t.Fatalf("query missing profile: %v", err)
	}
	if result.CoinFound || result.TextXPFound || result.VoiceXPFound || result.WorkConfigFound || result.WorkUserFound {
		t.Fatalf("expected missing optional data, got %#v", result)
	}
	if result.SignStatus != "沒有資料" {
		t.Fatalf("sign status = %q", result.SignStatus)
	}
}

func TestProfileServiceSignStatusUsesLegacyCooldown(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1, Today: 1_000})
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", ResetMarker: 3600})

	result, err := (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "user-1", Now: time.Unix(4_500, 0)})
	if err != nil {
		t.Fatalf("query signed profile: %v", err)
	}
	if result.SignStatus != "已簽到" {
		t.Fatalf("within cooldown sign status = %q", result.SignStatus)
	}
	result, err = (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "user-1", Now: time.Unix(4_700, 0)})
	if err != nil {
		t.Fatalf("query unsigned profile: %v", err)
	}
	if result.SignStatus != "未簽到" {
		t.Fatalf("after cooldown sign status = %q", result.SignStatus)
	}
}

func TestProfileServiceRejectsInvalidQuery(t *testing.T) {
	_, err := (ProfileService{Repository: fakemongo.NewEconomyProfileRepository()}).Query(context.Background(), ProfileQuery{GuildID: "guild-1"})
	if !errors.Is(err, domain.ErrInvalidEconomyProfileQuery) {
		t.Fatalf("expected ErrInvalidEconomyProfileQuery, got %v", err)
	}
	_, err = (ProfileService{}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "user-1"})
	if !errors.Is(err, domain.ErrInvalidEconomyProfileQuery) {
		t.Fatalf("expected ErrInvalidEconomyProfileQuery for nil repo, got %v", err)
	}
}

func TestProfileServicePropagatesUnexpectedRepositoryErrors(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	repo.Err = ports.ErrCoinLimitExceeded
	_, err := (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "user-1"})
	if !errors.Is(err, ports.ErrCoinLimitExceeded) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestLegacyProfileFormatting(t *testing.T) {
	cases := map[float64]string{
		999:           "999",
		1_000:         "1K",
		1_200:         "1.2K",
		1_000_000:     "1M",
		1_500_000_000: "1.5G",
		1.5:           "1.5",
		1_250:         "1.3K",
	}
	for input, want := range cases {
		if got := LegacyProfileAmount(input); got != want {
			t.Fatalf("LegacyProfileAmount(%v) = %q, want %q", input, got, want)
		}
	}
	if got := LegacyProfileXPRequired(2, false); got != 233 {
		t.Fatalf("text required XP = %d, want 233", got)
	}
	if got := LegacyProfileXPRequired(2, true); got != 300 {
		t.Fatalf("voice required XP = %d, want 300", got)
	}
}

func TestProfileServicePreservesLegacyCoinScalarRankAndDisplay(t *testing.T) {
	tests := []struct {
		name        string
		viewer      string
		other       string
		wantDisplay string
		wantRank    int
	}{
		{name: "undefined", viewer: "undefined", other: "500", wantDisplay: "NaN", wantRank: 2},
		{name: "null", viewer: "null", other: "500", wantDisplay: "0", wantRank: 2},
		{name: "decimal", viewer: "125.5", other: "500", wantDisplay: "125.5", wantRank: 2},
		{name: "positive infinity", viewer: "Infinity", other: "500", wantDisplay: "InfinityG", wantRank: 1},
		{name: "negative infinity", viewer: "-Infinity", other: "500", wantDisplay: "-Infinity", wantRank: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyProfileRepository()
			repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "viewer", CoinsText: test.viewer})
			repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "other", CoinsText: test.other})
			result, err := (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "viewer"})
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if got := LegacyProfileCoinAmount(result.CoinBalance); got != test.wantDisplay {
				t.Fatalf("display = %q want %q", got, test.wantDisplay)
			}
			if result.CoinRank != test.wantRank {
				t.Fatalf("rank = %d want %d", result.CoinRank, test.wantRank)
			}
		})
	}
}

func TestLegacyProfileRawAmountPreservesConfigScalars(t *testing.T) {
	tests := map[string]string{
		"undefined": "undefined",
		"null":      "null",
		"125.5":     "125.5",
		"Infinity":  "InfinityG",
		"-Infinity": "-Infinity",
	}
	for value, want := range tests {
		if got := LegacyProfileRawAmount(value, 0); got != want {
			t.Fatalf("LegacyProfileRawAmount(%q) = %q want %q", value, got, want)
		}
	}
	if got := LegacyProfileRawAmount("", 1250); got != "1.3K" {
		t.Fatalf("typed fallback = %q", got)
	}
}

func TestLegacyProfileWorkStatePreservesEndTimeScalars(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "undefined", want: "待業中"},
		{text: "null", want: "待業中"},
		{text: "999.5", want: "待業中"},
		{text: "1000.5", want: "打工中"},
		{text: "Infinity", want: "打工中"},
		{text: "-Infinity", want: "待業中"},
	}
	for _, test := range tests {
		if got := LegacyProfileWorkState(test.text, 0, 1000); got != test.want {
			t.Fatalf("LegacyProfileWorkState(%q) = %q want %q", test.text, got, test.want)
		}
	}
	if got := LegacyProfileWorkState("", 1001, 1000); got != "打工中" {
		t.Fatalf("typed fallback = %q", got)
	}
}

func TestProfileXPScalarsPreserveLegacyDisplayProgressAndRank(t *testing.T) {
	decimal := domain.XPProfile{XPText: "125.5", LevelText: "2.5"}
	if got := LegacyProfileXPAmount(decimal); got != "125.5" {
		t.Fatalf("decimal XP = %q", got)
	}
	if got := LegacyProfileLevelText(decimal); got != "2.5" {
		t.Fatalf("decimal level = %q", got)
	}
	if got := LegacyProfileXPRequiredForProfile(decimal, false); got != 308 {
		t.Fatalf("decimal required XP = %v", got)
	}
	malformed := domain.XPProfile{XPText: "undefined", LevelText: "undefined"}
	if got := LegacyProfileXPAmount(malformed); got != "NaN" {
		t.Fatalf("malformed XP = %q", got)
	}
	if got := LegacyProfileLevelText(malformed); got != "undefined" {
		t.Fatalf("malformed level = %q", got)
	}

	repo := fakemongo.NewEconomyProfileRepository()
	repo.PutTextXP(domain.XPProfile{GuildID: "guild-1", UserID: "viewer", XPText: "undefined", LevelText: "2"})
	repo.PutTextXP(domain.XPProfile{GuildID: "guild-1", UserID: "other", XPText: "500", LevelText: "2"})
	result, err := (ProfileService{Repository: repo}).Query(context.Background(), ProfileQuery{GuildID: "guild-1", UserID: "viewer"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if result.TextRank != 2 {
		t.Fatalf("malformed viewer rank = %d", result.TextRank)
	}
}
