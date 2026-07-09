package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrWarningsNotFound = errors.New("warnings not found")

type WarningHistoryRepository interface {
	GetWarningHistory(ctx context.Context, guildID string, userID string) (domain.WarningHistory, error)
}
