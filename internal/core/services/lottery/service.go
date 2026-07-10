package lottery

import (
	"context"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type Service struct {
	Repository ports.LotteryRepository
}

func NewService(repository ports.LotteryRepository) Service {
	return Service{Repository: repository}
}

func (s Service) Get(ctx context.Context, guildID string, id string) (domain.Lottery, error) {
	if err := ctx.Err(); err != nil {
		return domain.Lottery{}, err
	}
	if s.Repository == nil {
		return domain.Lottery{}, domain.ErrInvalidLottery
	}
	if err := domain.ValidateLotteryKey(guildID, id); err != nil {
		return domain.Lottery{}, err
	}
	return s.Repository.GetLottery(ctx, strings.TrimSpace(guildID), strings.TrimSpace(id))
}

func (s Service) Join(ctx context.Context, request domain.LotteryJoinRequest, actorRoleIDs []string, now time.Time) (domain.Lottery, error) {
	if err := ctx.Err(); err != nil {
		return domain.Lottery{}, err
	}
	if s.Repository == nil {
		return domain.Lottery{}, domain.ErrInvalidLottery
	}
	if now.IsZero() {
		now = time.Now()
	}
	request = request.Normalized()
	request.NowUnix = now.Unix()
	request.JoinedAtMillis = now.UnixMilli()
	if err := request.Validate(); err != nil {
		return domain.Lottery{}, err
	}
	lottery, err := s.Repository.GetLottery(ctx, request.GuildID, request.ID)
	if err != nil {
		return domain.Lottery{}, err
	}
	if lottery.HasParticipant(request.UserID) {
		if lottery.Ended {
			return domain.Lottery{}, ports.ErrLotteryEnded
		}
		return domain.Lottery{}, ports.ErrLotteryAlreadyJoined
	}
	if lottery.AtCapacity() {
		return domain.Lottery{}, ports.ErrLotteryFull
	}
	if lottery.IsExpired(now) {
		return domain.Lottery{}, ports.ErrLotteryEnded
	}
	if !lotteryActorHasRequiredRoles(lottery, actorRoleIDs) {
		return domain.Lottery{}, ports.ErrLotteryRoleDenied
	}
	return s.Repository.JoinLottery(ctx, request)
}

func (s Service) GetManaged(ctx context.Context, guildID string, id string, actorUserID string, guildOwnerID string, actorCanManageMessages bool) (domain.Lottery, error) {
	lottery, err := s.Get(ctx, guildID, id)
	if err != nil {
		return domain.Lottery{}, err
	}
	if !s.CanManage(lottery, actorUserID, guildOwnerID, actorCanManageMessages) {
		return domain.Lottery{}, ports.ErrLotteryManagerOnly
	}
	return lottery, nil
}

func (s Service) CanManage(lottery domain.Lottery, actorUserID string, guildOwnerID string, actorCanManageMessages bool) bool {
	return lotteryActorCanManage(lottery, actorUserID, guildOwnerID, actorCanManageMessages)
}

func (s Service) EndManaged(ctx context.Context, guildID string, id string, actorUserID string, guildOwnerID string, actorCanManageMessages bool) (domain.Lottery, error) {
	if _, err := s.GetManaged(ctx, guildID, id, actorUserID, guildOwnerID, actorCanManageMessages); err != nil {
		return domain.Lottery{}, err
	}
	return s.Repository.EndLottery(ctx, strings.TrimSpace(guildID), strings.TrimSpace(id))
}

func lotteryActorHasRequiredRoles(lottery domain.Lottery, roleIDs []string) bool {
	hasRequired := lottery.RequiredRoleID == ""
	hasForbidden := false
	for _, roleID := range roleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID == lottery.RequiredRoleID {
			hasRequired = true
		}
		if roleID == lottery.ForbiddenRoleID && lottery.ForbiddenRoleID != "" {
			hasForbidden = true
		}
	}
	return hasRequired && !hasForbidden
}

func lotteryActorCanManage(lottery domain.Lottery, actorUserID string, guildOwnerID string, actorCanManageMessages bool) bool {
	actorUserID = strings.TrimSpace(actorUserID)
	if actorUserID == "" {
		return false
	}
	if actorUserID == strings.TrimSpace(guildOwnerID) && strings.TrimSpace(guildOwnerID) != "" {
		return true
	}
	if lottery.OwnerID != "" {
		return actorUserID == lottery.OwnerID
	}
	return actorCanManageMessages
}
