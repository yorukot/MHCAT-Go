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

func (s VoiceSessionService) JoinedSessions(ctx context.Context) ([]domain.XPProfile, error) {
	if s.Repository == nil {
		return nil, domain.ErrInvalidXPAdjustment
	}
	profiles, err := s.Repository.ListJoinedVoiceXPSessions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.XPProfile, 0, len(profiles))
	for _, profile := range profiles {
		profile = profile.Normalize()
		if profile.GuildID == "" || profile.UserID == "" || profile.LeaveJoin != domain.VoiceXPSessionJoined {
			continue
		}
		out = append(out, profile)
	}
	return out, ctx.Err()
}
