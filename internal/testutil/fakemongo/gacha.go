package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type GachaRepository struct {
	Prizes  map[string][]domain.GachaPrize
	Configs map[string]domain.EconomyConfig
	Err     error
}

func NewGachaRepository() *GachaRepository {
	return &GachaRepository{
		Prizes:  map[string][]domain.GachaPrize{},
		Configs: map[string]domain.EconomyConfig{},
	}
}

func (r *GachaRepository) ListGachaPrizes(ctx context.Context, guildID string) ([]domain.GachaPrize, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return append([]domain.GachaPrize(nil), r.Prizes[guildID]...), nil
}

func (r *GachaRepository) GetGachaConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.EconomyConfig{}, err
	}
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.EconomyConfig{GuildID: guildID}, ports.ErrEconomyConfigMissing
	}
	return config, nil
}

func (r *GachaRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.GachaPrizePoolRepository = (*GachaRepository)(nil)
