package gacha

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	DefaultGachaCost  int64 = 500
	DefaultSignCoins  int64 = 25
	DefaultXPMultiple       = 0
)

type PrizePoolService struct {
	Repository ports.GachaPrizePoolRepository
}

func (s PrizePoolService) Query(ctx context.Context, guildID string) (domain.GachaPrizePool, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaPrizePool{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || s.Repository == nil {
		return domain.GachaPrizePool{}, domain.ErrInvalidGachaQuery
	}
	prizes, err := s.Repository.ListGachaPrizes(ctx, guildID)
	if err != nil {
		return domain.GachaPrizePool{}, err
	}
	if len(prizes) == 0 {
		return domain.GachaPrizePool{}, ports.ErrGachaPrizePoolEmpty
	}
	config, err := s.Repository.GetGachaConfig(ctx, guildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return domain.GachaPrizePool{}, err
		}
		configFound = false
		config = domain.EconomyConfig{
			GuildID:    guildID,
			GachaCost:  DefaultGachaCost,
			SignCoins:  DefaultSignCoins,
			XPMultiple: DefaultXPMultiple,
		}
	}
	return domain.GachaPrizePool{
		GuildID:     guildID,
		Prizes:      append([]domain.GachaPrize(nil), prizes...),
		Config:      config,
		ConfigFound: configFound,
	}, ctx.Err()
}
