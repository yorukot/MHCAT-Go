package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrTicketConfigNotFound = errors.New("ticket config not found")

type TicketConfigRepository interface {
	GetTicketConfig(ctx context.Context, guildID string) (domain.TicketConfig, error)
	SaveTicketConfig(ctx context.Context, config domain.TicketConfig) error
	DeleteTicketConfig(ctx context.Context, guildID string) error
}
