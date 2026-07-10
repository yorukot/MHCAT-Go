package xp

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ResetService struct {
	Repository ports.XPResetRepository
}

func (s ResetService) ResetTextProfile(ctx context.Context, guildID string, userID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID, userID = strings.TrimSpace(guildID), strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.Repository.DeleteTextXPProfile(ctx, guildID, userID)
}

func (s ResetService) ResetVoiceProfile(ctx context.Context, guildID string, userID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID, userID = strings.TrimSpace(guildID), strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.Repository.DeleteVoiceXPProfile(ctx, guildID, userID)
}

func (s ResetService) ResetTextGuild(ctx context.Context, guildID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err := s.Repository.DeleteTextXPGuild(ctx, guildID)
	if errors.Is(err, ports.ErrTextXPProfileMissing) {
		return nil
	}
	return err
}

func (s ResetService) ResetVoiceGuild(ctx context.Context, guildID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.Repository.DeleteVoiceXPGuild(ctx, guildID)
}
