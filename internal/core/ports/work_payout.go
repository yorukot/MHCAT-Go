package ports

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type WorkPayoutRepository interface {
	PreviewWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error)
	RunWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error)
}
