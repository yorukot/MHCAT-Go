package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrLotteryNotFound      = errors.New("lottery not found")
	ErrLotteryEnded         = errors.New("lottery ended")
	ErrLotteryAlreadyJoined = errors.New("lottery participant already joined")
	ErrLotteryFull          = errors.New("lottery is full")
	ErrLotteryRoleDenied    = errors.New("lottery role requirement denied")
	ErrLotteryManagerOnly   = errors.New("lottery manager only")
)

type LotteryRepository interface {
	GetLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error)
	JoinLottery(ctx context.Context, request domain.LotteryJoinRequest) (domain.Lottery, error)
	EndLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error)
}
