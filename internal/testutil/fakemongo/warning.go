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

type WarningRemovalRepository struct {
	Histories  map[string]domain.WarningHistory
	RemoveOnes []domain.WarningRemoval
	RemoveAlls []domain.WarningRemoval
}

func NewWarningHistoryRepository() *WarningHistoryRepository {
	return &WarningHistoryRepository{Histories: map[string]domain.WarningHistory{}}
}

func NewWarningSettingsRepository() *WarningSettingsRepository {
	return &WarningSettingsRepository{Settings: map[string]domain.WarningSettings{}}
}

func NewWarningRemovalRepository() *WarningRemovalRepository {
	return &WarningRemovalRepository{Histories: map[string]domain.WarningHistory{}}
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

func (r *WarningRemovalRepository) Put(history domain.WarningHistory) {
	if r.Histories == nil {
		r.Histories = map[string]domain.WarningHistory{}
	}
	r.Histories[warningHistoryKey(history.GuildID, history.UserID)] = history
}

func (r *WarningRemovalRepository) RemoveWarning(_ context.Context, removal domain.WarningRemoval) error {
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateSingle(); err != nil {
		return err
	}
	key := warningHistoryKey(removal.GuildID, removal.UserID)
	history, ok := r.Histories[key]
	if !ok || len(history.Entries) == 0 {
		return ports.ErrWarningsNotFound
	}
	index := int(removal.Index - 1)
	if index < 0 || index >= len(history.Entries) {
		return ports.ErrWarningsNotFound
	}
	next := append([]domain.WarningEntry(nil), history.Entries[:index]...)
	next = append(next, history.Entries[index+1:]...)
	history.Entries = next
	r.Histories[key] = history
	r.RemoveOnes = append(r.RemoveOnes, removal)
	return nil
}

func (r *WarningRemovalRepository) RemoveAllWarnings(_ context.Context, removal domain.WarningRemoval) error {
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateAll(); err != nil {
		return err
	}
	key := warningHistoryKey(removal.GuildID, removal.UserID)
	history, ok := r.Histories[key]
	if !ok || len(history.Entries) == 0 {
		return ports.ErrWarningsNotFound
	}
	delete(r.Histories, key)
	r.RemoveAlls = append(r.RemoveAlls, removal)
	return nil
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
var _ ports.WarningRemovalRepository = (*WarningRemovalRepository)(nil)
