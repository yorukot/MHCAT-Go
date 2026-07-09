package ports

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type DailyResetRepository interface {
	PreviewDailyReset(ctx context.Context) (domain.DailyResetResult, error)
	RunDailyReset(ctx context.Context) (domain.DailyResetResult, error)
}
