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

func NewWarningHistoryRepository() *WarningHistoryRepository {
	return &WarningHistoryRepository{Histories: map[string]domain.WarningHistory{}}
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

func warningHistoryKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(userID)
}

var _ ports.WarningHistoryRepository = (*WarningHistoryRepository)(nil)
