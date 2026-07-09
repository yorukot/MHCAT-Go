package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type WorkPayoutRepository struct {
	PreviewResult domain.WorkPayoutResult
	RunResult     domain.WorkPayoutResult
	PreviewErr    error
	RunErr        error
	PreviewCalls  int
	RunCalls      int
	PreviewNow    int64
	RunNow        int64
}

func (r *WorkPayoutRepository) PreviewWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkPayoutResult{}, err
	}
	r.PreviewCalls++
	r.PreviewNow = nowUnix
	if r.PreviewErr != nil {
		return domain.WorkPayoutResult{}, r.PreviewErr
	}
	return r.PreviewResult, nil
}

func (r *WorkPayoutRepository) RunWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkPayoutResult{}, err
	}
	r.RunCalls++
	r.RunNow = nowUnix
	if r.RunErr != nil {
		return domain.WorkPayoutResult{}, r.RunErr
	}
	return r.RunResult, nil
}
