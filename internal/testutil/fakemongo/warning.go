package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WarningHistoryRepository struct {
	Histories map[string]domain.WarningHistory
}

type WarningSettingsRepository struct {
	Settings map[string]domain.WarningSettings
	Saves    []domain.WarningSettings
}

func NewWarningHistoryRepository() *WarningHistoryRepository {
	return &WarningHistoryRepository{Histories: map[string]domain.WarningHistory{}}
}

func NewWarningSettingsRepository() *WarningSettingsRepository {
	return &WarningSettingsRepository{Settings: map[string]domain.WarningSettings{}}
}

func (r *WarningHistoryRepository) Put(history domain.WarningHistory) {
	if r.Histories == nil {
		r.Histories = map[string]domain.WarningHistory{}
	}
	r.Histories[warningHistoryKey(history.GuildID, history.UserID)] = history
}

func (r *WarningHistoryRepository) GetWarningHistory(_ context.Context, guildID string, userID string) (domain.WarningHistory, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WarningHistory{}, domain.ErrInvalidWarningQuery
	}
	history, ok := r.Histories[warningHistoryKey(guildID, userID)]
	if !ok || len(history.Entries) == 0 {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	return history, nil
}

func (r *WarningSettingsRepository) SaveWarningSettings(_ context.Context, settings domain.WarningSettings) error {
	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.Action = strings.TrimSpace(settings.Action)
	if err := settings.Validate(); err != nil {
		return err
	}
	if r.Settings == nil {
		r.Settings = map[string]domain.WarningSettings{}
	}
	r.Settings[settings.GuildID] = settings
	r.Saves = append(r.Saves, settings)
	return nil
}

func warningHistoryKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(userID)
}

var _ ports.WarningHistoryRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningSettingsRepository = (*WarningSettingsRepository)(nil)
