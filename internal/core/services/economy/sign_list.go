package economy

import (
	"context"
	"errors"
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
	rollingWindow, cooldown := legacySignListWindow(config, configFound)
	if rollingWindow {
		result.RollingWindow = true
		nowUnix := now.Unix()
		for _, balance := range balances {
			delta := nowUnix - balance.Today
			if float64(delta) < cooldown && delta > 0 {
				result.Entries = append(result.Entries, domain.SignInListEntry{
					UserID:       balance.UserID,
					SignedAtUnix: balance.Today,
					ShowSignedAt: true,
				})
			}
		}
		return result, ctx.Err()
	}
	for _, balance := range balances {
		if balance.Today == 1 {
			result.Entries = append(result.Entries, domain.SignInListEntry{UserID: balance.UserID})
		}
	}
	return result, ctx.Err()
}

func legacySignListWindow(config domain.EconomyConfig, configFound bool) (bool, float64) {
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
