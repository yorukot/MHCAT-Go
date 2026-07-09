package safety

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	repo ports.AntiScamConfigRepository
}

func NewConfigService(repo ports.AntiScamConfigRepository) ConfigService {
	return ConfigService{repo: repo}
}

func (s ConfigService) Toggle(ctx context.Context, guildID string) (domain.AntiScamConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AntiScamConfig{}, err
	}
	if s.repo == nil {
		return domain.AntiScamConfig{}, domain.ErrInvalidAntiScamConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.AntiScamConfig{}, domain.ErrInvalidAntiScamConfig
	}
	current, err := s.repo.FindAntiScamConfig(ctx, guildID)
	if err != nil && !errors.Is(err, ports.ErrAntiScamConfigMissing) {
		return domain.AntiScamConfig{}, err
	}
	config := domain.AntiScamConfig{
		GuildID: guildID,
		Open:    true,
	}
	if err == nil {
		config.Open = !current.Open
	}
	if err := config.Validate(); err != nil {
		return domain.AntiScamConfig{}, err
	}
	if err := s.repo.SaveAntiScamConfig(ctx, config); err != nil {
		return domain.AntiScamConfig{}, err
	}
	return config, ctx.Err()
}
