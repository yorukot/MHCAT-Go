package ports

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type BirthdayConfigRepository interface {
	SaveBirthdayConfig(ctx context.Context, config domain.BirthdayConfig) error
}
