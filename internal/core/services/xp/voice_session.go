package xp

import (
	"context"
	"sort"
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

func (s VoiceSessionService) Reconcile(ctx context.Context, guildID string, activeUserIDs []string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	unique := make(map[string]struct{}, len(activeUserIDs))
	for _, userID := range activeUserIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		unique[userID] = struct{}{}
	}
	users := make([]string, 0, len(unique))
	for userID := range unique {
		users = append(users, userID)
	}
	sort.Strings(users)
	return s.Repository.ReconcileVoiceXPSessions(ctx, guildID, users)
}
