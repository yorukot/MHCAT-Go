package economy

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SettingsService struct {
	Repository ports.EconomySettingsRepository
}

func (s SettingsService) Save(ctx context.Context, command domain.EconomySettingsCommand) (domain.EconomyConfig, error) {
	if s.Repository == nil {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.NotificationID = strings.TrimSpace(command.NotificationID)
	if command.GuildID == "" || command.NotificationID == "" {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	if command.GachaCost > MaxLegacyCoinBalance {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	if command.SignCoins > MaxLegacyCoinBalance {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	if command.SignCooldownHours < 0 {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	resetMarkerNumber := float64(command.SignCooldownHours) * 60 * 60
	resetMarker := int64(resetMarkerNumber)
	if resetMarkerNumber > math.MaxInt64 {
		resetMarker = math.MaxInt64
	}
	config := domain.EconomyConfig{
		GuildID:         command.GuildID,
		GachaCost:       command.GachaCost,
		SignCoins:       command.SignCoins,
		ChannelID:       command.NotificationID,
		XPMultiple:      command.XPMultiple,
		ResetMarker:     resetMarker,
		ResetMarkerText: strconv.FormatFloat(resetMarkerNumber, 'f', -1, 64),
	}
	return s.Repository.SaveEconomyConfig(ctx, config)
}
