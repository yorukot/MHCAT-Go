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

func (r *StatsConfigRepository) GetStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
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
	return config.Normalize(), nil
}

func (r *StatsConfigRepository) SaveStatsConfig(ctx context.Context, config domain.StatsConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	r.Put(config)
	return nil
}

func (r *StatsConfigRepository) AddStatsConfigChannel(ctx context.Context, guildID string, option string, channelID string, currentValue int) (domain.StatsConfig, error) {
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
	next, err := config.WithOptionalChannel(option, channelID, currentValue)
	if err != nil {
		return domain.StatsConfig{}, err
	}
	r.Configs[guildID] = next
	return next, nil
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
