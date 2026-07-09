package birthday

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	repo ports.BirthdayConfigRepository
}

func NewConfigService(repo ports.BirthdayConfigRepository) ConfigService {
	return ConfigService{repo: repo}
}

func (s ConfigService) Save(ctx context.Context, config domain.BirthdayConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidBirthdayConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.Message = strings.TrimSpace(config.Message)
	config.UTCOffset = strings.TrimSpace(config.UTCOffset)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	config.RoleID = strings.TrimSpace(config.RoleID)
	if err := config.Validate(); err != nil {
		return err
	}
	return s.repo.SaveBirthdayConfig(ctx, config)
}
