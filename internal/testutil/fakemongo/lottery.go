package fakemongo

import (
	"context"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type LotteryRepository struct {
	mu        sync.Mutex
	Lotteries map[string]domain.Lottery
	Ended     []string
	Err       error
}

func NewLotteryRepository() *LotteryRepository {
	return &LotteryRepository{Lotteries: map[string]domain.Lottery{}}
}

func (r *LotteryRepository) GetLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Lottery{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	lottery, ok := r.Lotteries[lotteryKey(guildID, id)]
	if !ok {
		return domain.Lottery{}, ports.ErrLotteryNotFound
	}
	return cloneLottery(lottery), nil
}

func (r *LotteryRepository) JoinLottery(ctx context.Context, request domain.LotteryJoinRequest) (domain.Lottery, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Lottery{}, err
	}
	request = request.Normalized()
	if err := request.Validate(); err != nil {
		return domain.Lottery{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := lotteryKey(request.GuildID, request.ID)
	lottery, ok := r.Lotteries[key]
	if !ok {
		return domain.Lottery{}, ports.ErrLotteryNotFound
	}
	if lottery.Ended || lottery.EndsAtUnix <= 0 || lottery.EndsAtUnix < request.NowUnix {
		return domain.Lottery{}, ports.ErrLotteryEnded
	}
	if lottery.HasParticipant(request.UserID) {
		return domain.Lottery{}, ports.ErrLotteryAlreadyJoined
	}
	if lottery.AtCapacity() {
		return domain.Lottery{}, ports.ErrLotteryFull
	}
	lottery.Participants = append(lottery.Participants, domain.LotteryParticipant{
		UserID:         request.UserID,
		JoinedAtMillis: request.JoinedAtMillis,
		JoinedAtRaw:    time.UnixMilli(request.JoinedAtMillis).Format(time.RFC3339),
	})
	r.Lotteries[key] = cloneLottery(lottery)
	return cloneLottery(lottery), nil
}

func (r *LotteryRepository) EndLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Lottery{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := lotteryKey(guildID, id)
	lottery, ok := r.Lotteries[key]
	if !ok {
		return domain.Lottery{}, ports.ErrLotteryNotFound
	}
	lottery.Ended = true
	r.Lotteries[key] = cloneLottery(lottery)
	r.Ended = append(r.Ended, key)
	return cloneLottery(lottery), nil
}

func (r *LotteryRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func lotteryKey(guildID string, id string) string {
	return guildID + ":" + id
}

func cloneLottery(lottery domain.Lottery) domain.Lottery {
	lottery.Participants = append([]domain.LotteryParticipant(nil), lottery.Participants...)
	return lottery
}

var _ ports.LotteryRepository = (*LotteryRepository)(nil)
