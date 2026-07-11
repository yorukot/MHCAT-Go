package stats

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	Repository ports.StatsConfigRepository
}

func (s ConfigService) Delete(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || s.Repository == nil {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	config, err := s.Repository.DeleteStatsConfig(ctx, guildID)
	if err != nil {
		return domain.StatsConfig{}, err
	}
	return config, nil
}
