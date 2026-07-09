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

func (s ProfileService) PrepareAdd(ctx context.Context, request domain.BirthdayAddRequest) (domain.BirthdayProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.BirthdayProfile{}, err
	}
	if s.repo == nil {
		return domain.BirthdayProfile{}, domain.ErrInvalidBirthdayProfile
	}
	request.GuildID = strings.TrimSpace(request.GuildID)
	request.ActorUserID = strings.TrimSpace(request.ActorUserID)
	request.TargetUserID = strings.TrimSpace(request.TargetUserID)
	if request.TargetUserID == "" {
		request.TargetUserID = request.ActorUserID
	}
	if request.GuildID == "" || request.ActorUserID == "" || request.TargetUserID == "" {
		return domain.BirthdayProfile{}, domain.ErrInvalidBirthdayProfile
	}
	config, err := s.repo.FindBirthdayConfig(ctx, request.GuildID)
	if err != nil {
		return domain.BirthdayProfile{}, err
	}
	if !config.EveryoneCanSetBirthdayDate && !request.ActorCanManageMessages {
		return domain.BirthdayProfile{}, domain.ErrBirthdayManageMessagesRequired
	}
	if request.TargetUserID != request.ActorUserID && !request.ActorCanManageMessages {
		return domain.BirthdayProfile{}, domain.ErrBirthdaySelfOnly
	}
	if request.TargetUserID != request.ActorUserID && request.ActorCanManageMessages {
		profile, err := s.repo.FindBirthdayProfile(ctx, request.GuildID, request.TargetUserID)
		if err != nil && !errors.Is(err, ports.ErrBirthdayProfileMissing) {
			return domain.BirthdayProfile{}, err
		}
		if err == nil && !profile.AllowAdmin {
			return domain.BirthdayProfile{}, domain.ErrBirthdayAdminNotAllowed
		}
	}
	if err := domain.ValidateBirthdayDate(request.BirthdayYear, request.BirthdayMonth, request.BirthdayDay, request.CurrentYear); err != nil {
		return domain.BirthdayProfile{}, err
	}
	month := request.BirthdayMonth
	day := request.BirthdayDay
	return domain.BirthdayProfile{
		GuildID:       request.GuildID,
		UserID:        request.TargetUserID,
		BirthdayYear:  intPointerOrNil(request.BirthdayYear),
		BirthdayMonth: &month,
		BirthdayDay:   &day,
		AllowAdmin:    true,
	}, nil
}

func (s ProfileService) SaveDateTime(ctx context.Context, profile domain.BirthdayProfile, hour int, minute int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidBirthdayProfile
	}
	profile.GuildID = strings.TrimSpace(profile.GuildID)
	profile.UserID = strings.TrimSpace(profile.UserID)
	profile.SendHour = &hour
	profile.SendMinute = &minute
	profile.AllowAdmin = true
	if err := profile.ValidateDateTime(); err != nil {
		return err
	}
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

func intPointerOrNil(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
