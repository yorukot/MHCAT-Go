package redeem

import (
	"context"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const LegacyCodeTTL = 7 * 24 * time.Hour

type Service struct {
	Repository ports.RedeemRepository
	Clock      ports.Clock
}

func NewService(repo ports.RedeemRepository, clock ports.Clock) Service {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Service{Repository: repo, Clock: clock}
}

func (s Service) Redeem(ctx context.Context, guildID string, code string) error {
	if s.Repository == nil {
		return domain.ErrInvalidRedeemCode
	}
	command := domain.RedeemCommand{
		GuildID: strings.TrimSpace(guildID),
		Code:    code,
		NowMS:   s.now().UnixMilli(),
	}
	if err := command.Validate(); err != nil {
		return err
	}
	redeemCode, err := s.Repository.GetRedeemCode(ctx, command.Code)
	if err != nil {
		return err
	}
	if err := redeemCode.Validate(); err != nil {
		return err
	}
	if float64(command.NowMS)-redeemCode.CreatedAtMillis > float64(LegacyCodeTTL.Milliseconds()) {
		return ports.ErrRedeemCodeExpired
	}
	return s.Repository.ConsumeRedeemCode(ctx, command, redeemCode)
}

func (s Service) now() time.Time {
	if s.Clock == nil {
		return time.Now()
	}
	return s.Clock.Now()
}
