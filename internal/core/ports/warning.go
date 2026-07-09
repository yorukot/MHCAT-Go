package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrWarningsNotFound = errors.New("warnings not found")
var ErrWarningSettingsUnavailable = errors.New("warning settings repository unavailable")
var ErrWarningRemovalUnavailable = errors.New("warning removal repository unavailable")

type WarningHistoryRepository interface {
	GetWarningHistory(ctx context.Context, guildID string, userID string) (domain.WarningHistory, error)
}

type WarningSettingsRepository interface {
	SaveWarningSettings(ctx context.Context, settings domain.WarningSettings) error
}

type WarningRemovalRepository interface {
	RemoveWarning(ctx context.Context, removal domain.WarningRemoval) error
	RemoveAllWarnings(ctx context.Context, removal domain.WarningRemoval) error
}
