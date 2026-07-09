package moderation

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WarningHistoryService struct {
	Repository ports.WarningHistoryRepository
}

type WarningSettingsService struct {
	Repository ports.WarningSettingsRepository
}

func (s WarningHistoryService) History(ctx context.Context, guildID string, userID string) (domain.WarningHistory, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WarningHistory{}, domain.ErrInvalidWarningQuery
	}
	if s.Repository == nil {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	history, err := s.Repository.GetWarningHistory(ctx, guildID, userID)
	if err != nil {
		return domain.WarningHistory{}, err
	}
	if len(history.Entries) == 0 {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	return history, nil
}

func (s WarningSettingsService) Configure(ctx context.Context, settings domain.WarningSettings) error {
	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.Action = strings.TrimSpace(settings.Action)
	if err := settings.Validate(); err != nil {
		return err
	}
	if s.Repository == nil {
		return ports.ErrWarningSettingsUnavailable
	}
	return s.Repository.SaveWarningSettings(ctx, settings)
}
