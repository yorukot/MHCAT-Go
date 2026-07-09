package utility

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type BalanceService struct {
	Repository ports.BalanceRepository
}

func (s BalanceService) Get(ctx context.Context, guildID string) (domain.Balance, error) {
	if err := ctx.Err(); err != nil {
		return domain.Balance{}, err
	}
	if s.Repository == nil {
		return domain.Balance{}, domain.ErrInvalidBalanceQuery
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.Balance{}, domain.ErrInvalidBalanceQuery
	}
	balance, err := s.Repository.GetBalance(ctx, guildID)
	if errors.Is(err, ports.ErrBalanceMissing) {
		return domain.Balance{GuildID: guildID, Amount: "0"}, nil
	}
	if err != nil {
		return domain.Balance{}, err
	}
	balance.GuildID = strings.TrimSpace(balance.GuildID)
	balance.Amount = strings.TrimSpace(balance.Amount)
	if balance.Amount == "" {
		balance.Amount = "0"
	}
	if err := balance.Validate(); err != nil {
		return domain.Balance{}, err
	}
	return balance, ctx.Err()
}
