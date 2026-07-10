package fakemongo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TicketConfigRepository struct {
	mu          sync.Mutex
	configs     map[string]domain.TicketConfig
	creationIDs map[string]string
	nextID      uint64
	Err         error
}

func NewTicketConfigRepository() *TicketConfigRepository {
	return &TicketConfigRepository{
		configs:     map[string]domain.TicketConfig{},
		creationIDs: map[string]string{},
	}
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

func (r *TicketConfigRepository) CreateTicketConfig(ctx context.Context, config domain.TicketConfig) (ports.TicketConfigCreation, error) {
	if err := ctx.Err(); err != nil {
		return ports.TicketConfigCreation{}, err
	}
	if r.Err != nil {
		return ports.TicketConfigCreation{}, r.Err
	}
	if err := config.ValidateForWrite(); err != nil {
		return ports.TicketConfigCreation{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.configs[config.GuildID]; ok {
		return ports.TicketConfigCreation{}, ports.ErrTicketConfigExists
	}
	r.nextID++
	creation := ports.TicketConfigCreation{
		GuildID: config.GuildID,
		ID:      fmt.Sprintf("ticket-config-%d", r.nextID),
	}
	r.configs[config.GuildID] = config
	r.creationIDs[config.GuildID] = creation.ID
	return creation, nil
}

func (r *TicketConfigRepository) RollbackTicketConfigCreation(ctx context.Context, creation ports.TicketConfigCreation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	if strings.TrimSpace(creation.GuildID) == "" || strings.TrimSpace(creation.ID) == "" {
		return domain.ErrInvalidTicketConfig
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.creationIDs[creation.GuildID] != creation.ID {
		return nil
	}
	delete(r.configs, creation.GuildID)
	delete(r.creationIDs, creation.GuildID)
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
	delete(r.creationIDs, guildID)
	return nil
}
