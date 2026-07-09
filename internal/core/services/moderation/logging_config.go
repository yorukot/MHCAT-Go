package moderation

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type LoggingConfigService struct {
	repo ports.LoggingConfigRepository
}

func NewLoggingConfigService(repo ports.LoggingConfigRepository) LoggingConfigService {
	return LoggingConfigService{repo: repo}
}

func (s LoggingConfigService) Save(ctx context.Context, config domain.LoggingConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidLoggingConfig
	}
	if err := config.Validate(); err != nil {
		return err
	}
	return s.repo.SaveLoggingConfig(ctx, config)
}
