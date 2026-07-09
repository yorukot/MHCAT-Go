package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrDeleteDataTargetMissing = errors.New("delete data target is missing")

type DeleteDataRepository interface {
	DeleteGuildConfig(ctx context.Context, request domain.DeleteDataRequest) error
}
