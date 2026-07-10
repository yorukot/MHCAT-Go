package xp

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type VoiceSessionService struct {
	Repository ports.VoiceXPSessionRepository
}

func (s VoiceSessionService) Join(ctx context.Context, guildID string, userID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	return s.Repository.MarkVoiceXPJoined(ctx, guildID, userID)
}

func (s VoiceSessionService) Leave(ctx context.Context, guildID string, userID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	return s.Repository.MarkVoiceXPLeft(ctx, guildID, userID)
}
