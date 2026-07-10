package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrTicketConfigNotFound = errors.New("ticket config not found")
var ErrTicketConfigExists = errors.New("ticket config already exists")

type TicketConfigCreation struct {
	GuildID string
	ID      string
}

type TicketConfigRepository interface {
	GetTicketConfig(ctx context.Context, guildID string) (domain.TicketConfig, error)
	CreateTicketConfig(ctx context.Context, config domain.TicketConfig) (TicketConfigCreation, error)
	RollbackTicketConfigCreation(ctx context.Context, creation TicketConfigCreation) error
	DeleteTicketConfig(ctx context.Context, guildID string) error
}
