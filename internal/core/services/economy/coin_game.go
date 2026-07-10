package economy

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type CoinGameService struct {
	Repository ports.EconomyCoinGameRepository
}

func (s CoinGameService) CheckBalances(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	if s.Repository == nil {
		return domain.CoinGameBalanceResult{}, domain.ErrInvalidCoinGameCommand
	}
	return s.Repository.CheckCoinGameBalances(ctx, command)
}

func (s CoinGameService) Reserve(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	if s.Repository == nil {
		return domain.CoinGameBalanceResult{}, domain.ErrInvalidCoinGameCommand
	}
	return s.Repository.ReserveCoinGameWager(ctx, command)
}

func (s CoinGameService) Settle(ctx context.Context, command domain.CoinGameSettlementCommand) (domain.CoinGameSettlementResult, error) {
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	if s.Repository == nil {
		return domain.CoinGameSettlementResult{}, domain.ErrInvalidCoinGameCommand
	}
	return s.Repository.SettleCoinGameWager(ctx, command)
}
