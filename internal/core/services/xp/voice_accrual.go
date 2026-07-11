package xp

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const LegacyVoiceXPTickAmount int64 = 5

type VoiceAccrualService struct {
	Repository ports.VoiceXPAccrualRepository
}

type VoiceAccrualResult struct {
	Profile domain.XPProfile
	Gained  int64
	Leveled bool
	Active  bool
}

func (s VoiceAccrualService) Tick(ctx context.Context, guildID string, userID string) (VoiceAccrualResult, error) {
	if s.Repository == nil {
		return VoiceAccrualResult{}, domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return VoiceAccrualResult{}, domain.ErrInvalidXPAdjustment
	}
	if repository, ok := s.Repository.(ports.AtomicVoiceXPAccrualRepository); ok {
		profile, active, leveled, err := repository.AccrueVoiceXP(ctx, guildID, userID, LegacyVoiceXPTickAmount)
		if err != nil {
			return VoiceAccrualResult{}, err
		}
		gained := int64(0)
		if active {
			gained = LegacyVoiceXPTickAmount
		}
		return VoiceAccrualResult{Profile: profile, Gained: gained, Leveled: leveled, Active: active}, ctx.Err()
	}
	profile, err := s.Repository.GetVoiceXPProfile(ctx, guildID, userID)
	if err != nil {
		return VoiceAccrualResult{}, err
	}
	profile = profile.Normalize()
	if profile.LeaveJoin != domain.VoiceXPSessionJoined {
		return VoiceAccrualResult{Profile: profile}, ctx.Err()
	}
	profile, leveled := domain.ApplyVoiceXPTick(profile, LegacyVoiceXPTickAmount)
	if err := profile.Validate(); err != nil {
		return VoiceAccrualResult{}, err
	}
	if err := s.Repository.SaveVoiceXPProfile(ctx, profile); err != nil {
		return VoiceAccrualResult{}, err
	}
	return VoiceAccrualResult{Profile: profile, Gained: LegacyVoiceXPTickAmount, Leveled: leveled, Active: true}, ctx.Err()
}
