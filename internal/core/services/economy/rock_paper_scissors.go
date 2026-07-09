package economy

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type RockPaperScissorsService struct {
	Repository ports.EconomyRockPaperScissorsRepository
}

func (s RockPaperScissorsService) Play(ctx context.Context, command domain.RockPaperScissorsCommand) (domain.RockPaperScissorsResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	if s.Repository == nil {
		return domain.RockPaperScissorsResult{}, domain.ErrInvalidRockPaperScissorsCommand
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	if command.Wager > MaxLegacyCoinBalance {
		return domain.RockPaperScissorsResult{}, domain.ErrInvalidRockPaperScissorsCommand
	}
	return s.Repository.ApplyRockPaperScissors(ctx, command)
}
