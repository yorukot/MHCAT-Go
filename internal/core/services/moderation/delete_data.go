package moderation

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type DeleteDataService struct {
	Repository ports.DeleteDataRepository
}

func (s DeleteDataService) Delete(ctx context.Context, request domain.DeleteDataRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidDeleteDataRequest
	}
	request = request.Normalize()
	if err := request.Validate(); err != nil {
		return err
	}
	return s.Repository.DeleteGuildConfig(ctx, request)
}
