package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type StatsConfigRepository struct {
	mu      sync.Mutex
	Configs map[string]domain.StatsConfig
	Err     error
}

func NewStatsConfigRepository() *StatsConfigRepository {
	return &StatsConfigRepository{Configs: map[string]domain.StatsConfig{}}
}

func (r *StatsConfigRepository) Put(config domain.StatsConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	config = config.Normalize()
	r.Configs[config.GuildID] = config
}

func (r *StatsConfigRepository) DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	if r.Err != nil {
		return domain.StatsConfig{}, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	delete(r.Configs, guildID)
	return config, nil
}

var _ ports.StatsConfigRepository = (*StatsConfigRepository)(nil)
