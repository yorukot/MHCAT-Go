package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WarningHistoryRepository struct {
	Histories map[string]domain.WarningHistory
	Err       error
}

type WarningSettingsRepository struct {
	Settings map[string]domain.WarningSettings
	Saves    []domain.WarningSettings
	Err      error
}

type WarningRemovalRepository struct {
	Histories  map[string]domain.WarningHistory
	RemoveOnes []domain.WarningRemoval
	RemoveAlls []domain.WarningRemoval
	Err        error
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
	if r.Err != nil {
		return domain.WarningHistory{}, r.Err
	}
	history, ok := r.Histories[warningHistoryKey(guildID, userID)]
	if !ok || len(history.Entries) == 0 {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	return history, nil
}

func (r *WarningHistoryRepository) AddWarning(_ context.Context, issue domain.WarningIssue) (domain.WarningIssueResult, error) {
	issue.GuildID = strings.TrimSpace(issue.GuildID)
	issue.UserID = strings.TrimSpace(issue.UserID)
	issue.ModeratorID = strings.TrimSpace(issue.ModeratorID)
	issue.Time = strings.TrimSpace(issue.Time)
	if err := issue.Validate(); err != nil {
		return domain.WarningIssueResult{}, err
	}
	if r.Err != nil {
		return domain.WarningIssueResult{}, r.Err
	}
	if r.Histories == nil {
		r.Histories = map[string]domain.WarningHistory{}
	}
	key := warningHistoryKey(issue.GuildID, issue.UserID)
	history, ok := r.Histories[key]
	created := !ok
	if created {
		history = domain.WarningHistory{GuildID: issue.GuildID, UserID: issue.UserID}
	}
	history.Entries = append(history.Entries, domain.WarningEntry{
		ModeratorID: issue.ModeratorID,
		Reason:      issue.Reason,
		Time:        issue.Time,
	})
	r.Histories[key] = history
	return domain.WarningIssueResult{History: history, Created: created}, nil
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
	if r.Err != nil {
		return r.Err
	}
	key := warningHistoryKey(removal.GuildID, removal.UserID)
	history, ok := r.Histories[key]
	if !ok {
		return ports.ErrWarningsNotFound
	}
	index, remove := fakeWarningSpliceIndex(removal.Index, len(history.Entries))
	next := append([]domain.WarningEntry(nil), history.Entries...)
	if remove {
		next = append([]domain.WarningEntry(nil), history.Entries[:index]...)
		next = append(next, history.Entries[index+1:]...)
	}
	history.Entries = next
	r.Histories[key] = history
	r.RemoveOnes = append(r.RemoveOnes, removal)
	return nil
}

func fakeWarningSpliceIndex(optionIndex int64, length int) (int, bool) {
	if length == 0 {
		return 0, false
	}
	start := optionIndex
	if start > -1<<63 {
		start--
	}
	if start < 0 {
		if start <= -int64(length) {
			start = 0
		} else {
			start = int64(length) + start
		}
	}
	if start >= int64(length) {
		return length, false
	}
	return int(start), true
}

func (r *WarningRemovalRepository) RemoveAllWarnings(_ context.Context, removal domain.WarningRemoval) error {
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateAll(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
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

func (r *WarningSettingsRepository) GetWarningSettings(_ context.Context, guildID string) (domain.WarningSettings, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.WarningSettings{}, domain.ErrInvalidWarningSettings
	}
	if r.Err != nil {
		return domain.WarningSettings{}, r.Err
	}
	settings, ok := r.Settings[guildID]
	if !ok {
		return domain.WarningSettings{}, ports.ErrWarningSettingsNotFound
	}
	return settings, nil
}

func (r *WarningSettingsRepository) SaveWarningSettings(_ context.Context, settings domain.WarningSettings) error {
	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.Action = strings.TrimSpace(settings.Action)
	if err := settings.ValidateWrite(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
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
var _ ports.WarningIssueRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningSettingsRepository = (*WarningSettingsRepository)(nil)
var _ ports.WarningRemovalRepository = (*WarningRemovalRepository)(nil)
