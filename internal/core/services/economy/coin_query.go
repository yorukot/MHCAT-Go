package economy

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const DefaultGachaCost int64 = 500

type CoinQueryService struct {
	Repository ports.EconomyQueryRepository
}

type CoinQueryResult struct {
	Balance      domain.CoinBalance
	Config       domain.EconomyConfig
	ConfigFound  bool
	GachaCost    int64
	MissingCoins int64
}

func (s CoinQueryService) Query(ctx context.Context, guildID string, userID string) (CoinQueryResult, error) {
	if s.Repository == nil {
		return CoinQueryResult{}, domain.ErrInvalidEconomyQuery
	}
	balance, err := s.Repository.GetCoinBalance(ctx, guildID, userID)
	if err != nil {
		return CoinQueryResult{}, err
	}
	config, err := s.Repository.GetEconomyConfig(ctx, guildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return CoinQueryResult{}, err
		}
		config = domain.EconomyConfig{GuildID: guildID, GachaCost: DefaultGachaCost}
		configFound = false
	}
	gachaCost := config.GachaCost
	if !configFound {
		gachaCost = DefaultGachaCost
	}
	missing := gachaCost - balance.Coins
	if missing < 0 {
		missing = 0
	}
	return CoinQueryResult{
		Balance:      balance,
		Config:       config,
		ConfigFound:  configFound,
		GachaCost:    gachaCost,
		MissingCoins: missing,
	}, nil
}
