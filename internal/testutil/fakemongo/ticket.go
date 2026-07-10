package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TicketConfigRepository struct {
	mu      sync.Mutex
	configs map[string]domain.TicketConfig
	Err     error
}

func NewTicketConfigRepository() *TicketConfigRepository {
	return &TicketConfigRepository{configs: map[string]domain.TicketConfig{}}
}

func (r *TicketConfigRepository) GetTicketConfig(ctx context.Context, guildID string) (domain.TicketConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.TicketConfig{}, err
	}
	if r.Err != nil {
		return domain.TicketConfig{}, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.configs[guildID]
	if !ok {
		return domain.TicketConfig{}, ports.ErrTicketConfigNotFound
	}
	return config, nil
}

func (r *TicketConfigRepository) CreateTicketConfig(ctx context.Context, config domain.TicketConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	if err := config.ValidateForWrite(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.configs[config.GuildID]; ok {
		return ports.ErrTicketConfigExists
	}
	r.configs[config.GuildID] = config
	return nil
}

func (r *TicketConfigRepository) DeleteTicketConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.configs[guildID]; !ok {
		return ports.ErrTicketConfigNotFound
	}
	delete(r.configs, guildID)
	return nil
}
