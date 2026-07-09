package birthday

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ProfileService struct {
	repo ports.BirthdayConfigRepository
}

func NewProfileService(repo ports.BirthdayConfigRepository) ProfileService {
	return ProfileService{repo: repo}
}

func (s ProfileService) SetAllowAdmin(ctx context.Context, guildID string, userID string, allow bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidBirthdayProfile
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidBirthdayProfile
	}
	profile, err := s.repo.FindBirthdayProfile(ctx, guildID, userID)
	if err != nil {
		if !errors.Is(err, ports.ErrBirthdayProfileMissing) {
			return err
		}
		profile = domain.BirthdayProfile{GuildID: guildID, UserID: userID}
	}
	profile.GuildID = guildID
	profile.UserID = userID
	profile.AllowAdmin = allow
	return s.repo.SaveBirthdayProfile(ctx, profile)
}

func (s ProfileService) Delete(ctx context.Context, guildID string, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidBirthdayProfile
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidBirthdayProfile
	}
	return s.repo.DeleteBirthdayProfile(ctx, guildID, userID)
}

func (s ProfileService) List(ctx context.Context, guildID string) ([]domain.BirthdayProfile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.repo == nil {
		return nil, domain.ErrInvalidBirthdayProfile
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidBirthdayProfile
	}
	return s.repo.ListBirthdayProfiles(ctx, guildID)
}
