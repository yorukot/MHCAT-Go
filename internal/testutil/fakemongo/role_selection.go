package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type RoleSelectionRepository struct {
	mu        sync.Mutex
	Reactions map[string]domain.RoleReactionConfig
	Buttons   map[string]domain.RoleButtonConfig
	Err       error
}

func NewRoleSelectionRepository() *RoleSelectionRepository {
	return &RoleSelectionRepository{
		Reactions: map[string]domain.RoleReactionConfig{},
		Buttons:   map[string]domain.RoleButtonConfig{},
	}
}

func (r *RoleSelectionRepository) SaveRoleReactionConfig(ctx context.Context, config domain.RoleReactionConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Reactions[roleReactionKey(config.GuildID, config.MessageID, config.React)] = config
	return nil
}

func (r *RoleSelectionRepository) DeleteRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := roleReactionKey(guildID, messageID, react)
	if _, ok := r.Reactions[key]; !ok {
		return ports.ErrRoleReactionConfigMissing
	}
	delete(r.Reactions, key)
	return nil
}

func (r *RoleSelectionRepository) GetRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) (domain.RoleReactionConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Reactions[roleReactionKey(guildID, messageID, react)]
	if !ok {
		return domain.RoleReactionConfig{}, ports.ErrRoleReactionConfigMissing
	}
	return config, nil
}

func (r *RoleSelectionRepository) SaveRoleButtonConfigs(ctx context.Context, configs ...domain.RoleButtonConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, config := range configs {
		config = config.Normalize()
		if err := config.Validate(); err != nil {
			return err
		}
		r.Buttons[roleButtonKey(config.GuildID, config.Number)] = config
	}
	return nil
}

func (r *RoleSelectionRepository) GetRoleButtonConfig(ctx context.Context, guildID string, number string) (domain.RoleButtonConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.RoleButtonConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Buttons[roleButtonKey(guildID, number)]
	if !ok {
		return domain.RoleButtonConfig{}, ports.ErrRoleButtonConfigMissing
	}
	return config, nil
}

func (r *RoleSelectionRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	return nil
}

func roleReactionKey(guildID string, messageID string, react string) string {
	return guildID + "/" + messageID + "/" + react
}

func roleButtonKey(guildID string, number string) string {
	return guildID + "/" + number
}

var _ ports.RoleSelectionRepository = (*RoleSelectionRepository)(nil)
