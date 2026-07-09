package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AutoChatConfigRepository struct {
	Configs map[string]domain.AutoChatConfig
	Err     error
	Saved   []domain.AutoChatConfig
}

func NewAutoChatConfigRepository() *AutoChatConfigRepository {
	return &AutoChatConfigRepository{Configs: map[string]domain.AutoChatConfig{}}
}

func (r *AutoChatConfigRepository) SaveAutoChatConfig(ctx context.Context, config domain.AutoChatConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if r.Configs == nil {
		r.Configs = map[string]domain.AutoChatConfig{}
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	r.Configs[config.GuildID] = config
	r.Saved = append(r.Saved, config)
	return nil
}

func (r *AutoChatConfigRepository) DeleteAutoChatConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if r.Configs == nil {
		r.Configs = map[string]domain.AutoChatConfig{}
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrAutoChatConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *AutoChatConfigRepository) Last() (domain.AutoChatConfig, bool) {
	if len(r.Saved) == 0 {
		return domain.AutoChatConfig{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

func (r *AutoChatConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.AutoChatConfigRepository = (*AutoChatConfigRepository)(nil)
