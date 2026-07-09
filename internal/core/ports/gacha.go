package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrGachaPrizePoolEmpty = errors.New("gacha prize pool is empty")
	ErrGachaPrizeMissing   = errors.New("gacha prize is missing")
)

type GachaPrizePoolRepository interface {
	ListGachaPrizes(ctx context.Context, guildID string) ([]domain.GachaPrize, error)
	GetGachaConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error)
}

type GachaPrizeDeleteRepository interface {
	DeleteGachaPrize(ctx context.Context, guildID string, prizeName string) (domain.GachaPrize, error)
}
