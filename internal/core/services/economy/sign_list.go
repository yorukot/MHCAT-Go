package economy

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SignInListService struct {
	Repository ports.EconomySignInRepository
}

func (s SignInListService) List(ctx context.Context, guildID string, actorUserID string, now time.Time) (domain.SignInListResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.SignInListResult{}, err
	}
	if s.Repository == nil {
		return domain.SignInListResult{}, domain.ErrInvalidSignIn
	}
	guildID = strings.TrimSpace(guildID)
	actorUserID = strings.TrimSpace(actorUserID)
	if guildID == "" || actorUserID == "" || now.IsZero() {
		return domain.SignInListResult{}, domain.ErrInvalidSignIn
	}
	balances, err := s.Repository.ListCoinBalances(ctx, guildID)
	if err != nil {
		return domain.SignInListResult{}, err
	}
	config, err := s.Repository.GetEconomyConfig(ctx, guildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return domain.SignInListResult{}, err
		}
		configFound = false
	}
	result := domain.SignInListResult{
		GuildID:     guildID,
		ActorUserID: actorUserID,
		Entries:     []domain.SignInListEntry{},
	}
	rollingWindow, cooldown := LegacySignWindow(config, configFound)
	if rollingWindow {
		result.RollingWindow = true
		nowUnix := float64(LegacyRoundedUnixSeconds(now))
		for _, balance := range balances {
			today, numeric := legacySignListToday(balance)
			delta := nowUnix - today
			if numeric && delta < cooldown && delta > 0 {
				result.Entries = append(result.Entries, domain.SignInListEntry{
					UserID:       balance.UserID,
					SignedAtUnix: today,
					ShowSignedAt: true,
				})
			}
		}
		return result, ctx.Err()
	}
	for _, balance := range balances {
		today, numeric := legacySignListToday(balance)
		if numeric && today == 1 {
			result.Entries = append(result.Entries, domain.SignInListEntry{UserID: balance.UserID})
		}
	}
	return result, ctx.Err()
}

// LegacyRoundedUnixSeconds matches Math.round(Date.now() / 1000) for current positive epochs.
func LegacyRoundedUnixSeconds(now time.Time) int64 {
	return int64(math.Round(float64(now.UnixNano()) / float64(time.Second)))
}

func legacySignListToday(balance domain.CoinBalance) (float64, bool) {
	todayText := strings.TrimSpace(balance.TodayText)
	if todayText == "" {
		return float64(balance.Today), true
	}
	return legacyDisplayedNumber(todayText)
}

// LegacySignWindow preserves the JavaScript mode and fallback semantics of gift_changes.time.
func LegacySignWindow(config domain.EconomyConfig, configFound bool) (bool, float64) {
	if !configFound {
		return false, 0
	}
	markerText := strings.TrimSpace(config.ResetMarkerText)
	if markerText == "" {
		if config.ResetMarker == 0 {
			return false, 0
		}
		return true, float64(config.ResetMarker)
	}
	if markerText == "undefined" || markerText == "null" || markerText == "NaN" {
		return true, float64(DefaultSignCooldownSec)
	}
	marker, numeric := legacyDisplayedNumber(markerText)
	if !numeric {
		return true, float64(DefaultSignCooldownSec)
	}
	if marker == 0 {
		return false, 0
	}
	return true, marker
}

// LegacySignReward preserves numeric sign_coin values while rejecting undefined and NaN writes.
func LegacySignReward(config domain.EconomyConfig, configFound bool) (float64, bool) {
	if !configFound {
		return float64(DefaultSignCoins), true
	}
	rewardText := strings.TrimSpace(config.SignCoinsText)
	if rewardText == "" {
		return float64(config.SignCoins), true
	}
	return LegacyEconomyNumber(rewardText)
}
