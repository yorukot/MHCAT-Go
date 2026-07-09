package ports

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type LoggingConfigRepository interface {
	SaveLoggingConfig(ctx context.Context, config domain.LoggingConfig) error
}
