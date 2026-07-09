package xp

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AdminService struct {
	Repository ports.XPAdminRepository
}

func (s AdminService) AddTextXP(ctx context.Context, adjustment domain.XPAdjustment) (domain.XPProfile, error) {
	if s.Repository == nil {
		return domain.XPProfile{}, domain.ErrInvalidXPAdjustment
	}
	return s.adjust(ctx, adjustment, s.Repository.GetTextXPProfile, s.Repository.SaveTextXPProfile, ports.ErrTextXPProfileMissing, domain.ApplyTextXPAdjustment)
}

func (s AdminService) AddVoiceXP(ctx context.Context, adjustment domain.XPAdjustment) (domain.XPProfile, error) {
	if s.Repository == nil {
		return domain.XPProfile{}, domain.ErrInvalidXPAdjustment
	}
	return s.adjust(ctx, adjustment, s.Repository.GetVoiceXPProfile, s.Repository.SaveVoiceXPProfile, ports.ErrVoiceXPProfileMissing, domain.ApplyVoiceXPAdjustment)
}

func (s AdminService) adjust(
	ctx context.Context,
	adjustment domain.XPAdjustment,
	get func(context.Context, string, string) (domain.XPProfile, error),
	save func(context.Context, domain.XPProfile) error,
	missing error,
	apply func(domain.XPProfile, int64) domain.XPProfile,
) (domain.XPProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.XPProfile{}, err
	}
	adjustment = adjustment.Normalize()
	if err := adjustment.Validate(); err != nil {
		return domain.XPProfile{}, err
	}
	profile, err := get(ctx, adjustment.GuildID, adjustment.UserID)
	if err != nil {
		if !errors.Is(err, missing) {
			return domain.XPProfile{}, err
		}
		profile = domain.XPProfile{GuildID: adjustment.GuildID, UserID: adjustment.UserID}
	}
	profile = apply(profile, adjustment.Delta)
	if err := profile.Validate(); err != nil {
		return domain.XPProfile{}, err
	}
	if err := save(ctx, profile); err != nil {
		return domain.XPProfile{}, err
	}
	return profile, ctx.Err()
}
