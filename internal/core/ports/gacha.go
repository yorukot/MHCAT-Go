package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrGachaPrizePoolEmpty = errors.New("gacha prize pool is empty")

type GachaPrizePoolRepository interface {
	ListGachaPrizes(ctx context.Context, guildID string) ([]domain.GachaPrize, error)
	GetGachaConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error)
}
