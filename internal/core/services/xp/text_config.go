package xp

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TextConfigService struct {
	Repository ports.TextXPConfigRepository
}

type VoiceConfigService struct {
	Repository ports.VoiceXPConfigRepository
}

func (s TextConfigService) Save(ctx context.Context, config domain.TextXPConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidTextXPConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	config.Color = strings.TrimSpace(config.Color)
	if err := config.Validate(); err != nil {
		return err
	}
	return s.Repository.SaveTextXPConfig(ctx, config)
}

func (s TextConfigService) Delete(ctx context.Context, guildID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidTextXPConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidTextXPConfig
	}
	return s.Repository.DeleteTextXPConfig(ctx, guildID)
}

func (s VoiceConfigService) Save(ctx context.Context, config domain.VoiceXPConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceXPConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	config.Color = strings.TrimSpace(config.Color)
	if err := config.Validate(); err != nil {
		return err
	}
	return s.Repository.SaveVoiceXPConfig(ctx, config)
}

func (s VoiceConfigService) Delete(ctx context.Context, guildID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceXPConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidVoiceXPConfig
	}
	return s.Repository.DeleteVoiceXPConfig(ctx, guildID)
}
