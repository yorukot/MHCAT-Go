package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrStatsConfigMissing = errors.New("stats config is missing")

type StatsConfigRepository interface {
	DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error)
}
