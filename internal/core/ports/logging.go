package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type LoggingConfigRepository interface {
	SaveLoggingConfig(ctx context.Context, config domain.LoggingConfig) error
}

var ErrLoggingConfigMissing = errors.New("logging config is missing")

type LoggingConfigReader interface {
	GetLoggingConfig(ctx context.Context, guildID string) (domain.LoggingConfig, error)
}
