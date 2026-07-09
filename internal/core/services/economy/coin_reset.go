package economy

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type CoinResetService struct {
	Repository ports.EconomyCoinResetRepository
}

func (s CoinResetService) Reset(ctx context.Context, command domain.CoinResetCommand) (domain.CoinResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinResetResult{}, err
	}
	if s.Repository == nil {
		return domain.CoinResetResult{}, domain.ErrInvalidCoinResetCommand
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinResetResult{}, err
	}
	return s.Repository.ResetCoinBalances(ctx, command)
}
