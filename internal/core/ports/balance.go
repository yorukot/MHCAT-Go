package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrBalanceMissing = errors.New("balance is missing")

type BalanceRepository interface {
	GetBalance(ctx context.Context, guildID string) (domain.Balance, error)
}
