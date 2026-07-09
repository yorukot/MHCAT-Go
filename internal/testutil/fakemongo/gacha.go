package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type GachaRepository struct {
	Prizes       map[string][]domain.GachaPrize
	PrizeConfigs map[string][]domain.GachaPrizeConfig
	Configs      map[string]domain.EconomyConfig
	Err          error
}

func NewGachaRepository() *GachaRepository {
	return &GachaRepository{
		Prizes:       map[string][]domain.GachaPrize{},
		PrizeConfigs: map[string][]domain.GachaPrizeConfig{},
		Configs:      map[string]domain.EconomyConfig{},
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

func (r *GachaRepository) DeleteGachaPrize(ctx context.Context, guildID string, prizeName string) (domain.GachaPrize, error) {
	if err := r.ready(ctx); err != nil {
		return domain.GachaPrize{}, err
	}
	prizes := r.Prizes[guildID]
	for index, prize := range prizes {
		if prize.Name != prizeName {
			continue
		}
		r.Prizes[guildID] = append(append([]domain.GachaPrize(nil), prizes[:index]...), prizes[index+1:]...)
		configs := r.PrizeConfigs[guildID]
		for configIndex, config := range configs {
			if config.Name == prizeName {
				r.PrizeConfigs[guildID] = append(append([]domain.GachaPrizeConfig(nil), configs[:configIndex]...), configs[configIndex+1:]...)
				break
			}
		}
		return prize, nil
	}
	return domain.GachaPrize{}, ports.ErrGachaPrizeMissing
}

func (r *GachaRepository) CountGachaPrizes(ctx context.Context, guildID string) (int64, error) {
	if err := r.ready(ctx); err != nil {
		return 0, err
	}
	return int64(len(r.Prizes[guildID])), nil
}

func (r *GachaRepository) CreateGachaPrize(ctx context.Context, prize domain.GachaPrizeConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if prize.GuildID == "" || prize.Name == "" || prize.Count <= 0 {
		return domain.ErrInvalidGachaPrize
	}
	for _, existing := range r.Prizes[prize.GuildID] {
		if existing.Name == prize.Name {
			return ports.ErrGachaPrizeExists
		}
	}
	r.Prizes[prize.GuildID] = append(r.Prizes[prize.GuildID], domain.GachaPrize{
		GuildID: prize.GuildID,
		Name:    prize.Name,
		Chance:  prize.Chance,
		Count:   prize.Count,
	})
	r.PrizeConfigs[prize.GuildID] = append(r.PrizeConfigs[prize.GuildID], prize)
	return nil
}

func (r *GachaRepository) EditGachaPrize(ctx context.Context, edit domain.GachaPrizeEdit) (domain.GachaPrizeConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.GachaPrizeConfig{}, err
	}
	if edit.GuildID == "" || edit.Name == "" || edit.Count <= 0 {
		return domain.GachaPrizeConfig{}, domain.ErrInvalidGachaPrize
	}
	prizes := r.Prizes[edit.GuildID]
	for index, prize := range prizes {
		if prize.Name != edit.Name {
			continue
		}
		existing := domain.GachaPrizeConfig{
			GuildID:    prize.GuildID,
			Name:       prize.Name,
			Chance:     prize.Chance,
			Count:      prize.Count,
			AutoDelete: true,
		}
		configs := r.PrizeConfigs[edit.GuildID]
		configIndex := -1
		for index, config := range configs {
			if config.Name == edit.Name {
				existing = config
				configIndex = index
				break
			}
		}
		updated := mergeLegacyFakeGachaPrizeEdit(existing, edit)
		r.Prizes[edit.GuildID][index] = domain.GachaPrize{
			GuildID: updated.GuildID,
			Name:    updated.Name,
			Chance:  updated.Chance,
			Count:   updated.Count,
		}
		if configIndex >= 0 {
			r.PrizeConfigs[edit.GuildID][configIndex] = updated
		} else {
			r.PrizeConfigs[edit.GuildID] = append(r.PrizeConfigs[edit.GuildID], updated)
		}
		return updated, nil
	}
	return domain.GachaPrizeConfig{}, ports.ErrGachaPrizeMissing
}

func mergeLegacyFakeGachaPrizeEdit(existing domain.GachaPrizeConfig, edit domain.GachaPrizeEdit) domain.GachaPrizeConfig {
	updated := existing
	updated.GuildID = edit.GuildID
	updated.Name = edit.Name
	if edit.Code != "" {
		updated.Code = edit.Code
	}
	if edit.ChanceSet && edit.Chance != 0 {
		updated.Chance = edit.Chance
	}
	if edit.AutoDelete {
		updated.AutoDelete = true
	}
	updated.Count = edit.Count
	if edit.GiveCoin != 0 {
		updated.GiveCoin = edit.GiveCoin
	}
	return updated
}

func (r *GachaRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.GachaPrizePoolRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeDeleteRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeCreateRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeEditRepository = (*GachaRepository)(nil)
