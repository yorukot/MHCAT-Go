package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrBirthdayProfileMissing = errors.New("birthday profile is missing")

type BirthdayConfigRepository interface {
	SaveBirthdayConfig(ctx context.Context, config domain.BirthdayConfig) error
	FindBirthdayProfile(ctx context.Context, guildID string, userID string) (domain.BirthdayProfile, error)
	SaveBirthdayProfile(ctx context.Context, profile domain.BirthdayProfile) error
	DeleteBirthdayProfile(ctx context.Context, guildID string, userID string) error
	ListBirthdayProfiles(ctx context.Context, guildID string) ([]domain.BirthdayProfile, error)
}
