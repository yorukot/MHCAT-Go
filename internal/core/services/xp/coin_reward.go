package xp

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TextCoinRewardService struct {
	Repository ports.TextXPCoinRewardRepository
}

type VoiceCoinRewardService struct {
	Repository ports.VoiceXPCoinRewardRepository
}

func (s TextCoinRewardService) ApplyLevelUp(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	if s.Repository == nil {
		return domain.CoinBalance{}, domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" || level < 0 {
		return domain.CoinBalance{}, domain.ErrInvalidXPAdjustment
	}
	return s.Repository.ApplyTextXPCoinReward(ctx, guildID, userID, level)
}

func (s VoiceCoinRewardService) ApplyLevelUp(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	if s.Repository == nil {
		return domain.CoinBalance{}, domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" || level < 0 {
		return domain.CoinBalance{}, domain.ErrInvalidXPAdjustment
	}
	return s.Repository.ApplyVoiceXPCoinReward(ctx, guildID, userID, level)
}
