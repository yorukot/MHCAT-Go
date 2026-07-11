package economy

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SettingsService struct {
	Repository ports.EconomySettingsRepository
}

const maxCooldownHours = int64(1<<63-1) / 3600

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
	if command.SignCooldownHours < 0 || command.SignCooldownHours > maxCooldownHours {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	config := domain.EconomyConfig{
		GuildID:     command.GuildID,
		GachaCost:   command.GachaCost,
		SignCoins:   command.SignCoins,
		ChannelID:   command.NotificationID,
		XPMultiple:  command.XPMultiple,
		ResetMarker: command.SignCooldownHours * 60 * 60,
	}
	return s.Repository.SaveEconomyConfig(ctx, config)
}
