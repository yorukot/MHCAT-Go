package economy

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type CoinAdminService struct {
	Repository ports.EconomyCoinAdminRepository
}

func (s CoinAdminService) Adjust(ctx context.Context, command domain.CoinAdminCommand) (domain.CoinAdminResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinAdminResult{}, err
	}
	if s.Repository == nil {
		return domain.CoinAdminResult{}, domain.ErrInvalidCoinAdminCommand
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinAdminResult{}, err
	}
	return s.Repository.AdjustCoinBalance(ctx, command)
}
