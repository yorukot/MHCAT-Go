package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TextXPConfigRepository struct {
	Configs map[string]domain.TextXPConfig
	Err     error
}

type VoiceXPConfigRepository struct {
	Configs map[string]domain.VoiceXPConfig
	Err     error
}

func NewTextXPConfigRepository() *TextXPConfigRepository {
	return &TextXPConfigRepository{Configs: map[string]domain.TextXPConfig{}}
}

func NewVoiceXPConfigRepository() *VoiceXPConfigRepository {
	return &VoiceXPConfigRepository{Configs: map[string]domain.VoiceXPConfig{}}
}

func (r *TextXPConfigRepository) SaveTextXPConfig(ctx context.Context, config domain.TextXPConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.Configs[strings.TrimSpace(config.GuildID)] = config
	return nil
}

func (r *TextXPConfigRepository) DeleteTextXPConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrTextXPConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *VoiceXPConfigRepository) SaveVoiceXPConfig(ctx context.Context, config domain.VoiceXPConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.Configs[strings.TrimSpace(config.GuildID)] = config
	return nil
}

func (r *VoiceXPConfigRepository) DeleteVoiceXPConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrVoiceXPConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *TextXPConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func (r *VoiceXPConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.TextXPConfigRepository = (*TextXPConfigRepository)(nil)
var _ ports.VoiceXPConfigRepository = (*VoiceXPConfigRepository)(nil)
