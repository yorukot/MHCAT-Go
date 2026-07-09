package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type DailyResetRepository struct {
	PreviewResult domain.DailyResetResult
	RunResult     domain.DailyResetResult
	PreviewErr    error
	RunErr        error
	PreviewCalls  int
	RunCalls      int
}

func (r *DailyResetRepository) PreviewDailyReset(ctx context.Context) (domain.DailyResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.DailyResetResult{}, err
	}
	r.PreviewCalls++
	if r.PreviewErr != nil {
		return domain.DailyResetResult{}, r.PreviewErr
	}
	return r.PreviewResult, nil
}

func (r *DailyResetRepository) RunDailyReset(ctx context.Context) (domain.DailyResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.DailyResetResult{}, err
	}
	r.RunCalls++
	if r.RunErr != nil {
		return domain.DailyResetResult{}, r.RunErr
	}
	return r.RunResult, nil
}
