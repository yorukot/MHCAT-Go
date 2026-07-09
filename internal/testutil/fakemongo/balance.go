package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type BalanceRepository struct {
	Balances map[string]domain.Balance
	Err      error
}

func NewBalanceRepository() *BalanceRepository {
	return &BalanceRepository{Balances: map[string]domain.Balance{}}
}

func (r *BalanceRepository) GetBalance(ctx context.Context, guildID string) (domain.Balance, error) {
	if err := ctx.Err(); err != nil {
		return domain.Balance{}, err
	}
	if r.Err != nil {
		return domain.Balance{}, r.Err
	}
	guildID = strings.TrimSpace(guildID)
	balance, ok := r.Balances[guildID]
	if !ok {
		return domain.Balance{}, ports.ErrBalanceMissing
	}
	return balance, nil
}

var _ ports.BalanceRepository = (*BalanceRepository)(nil)
