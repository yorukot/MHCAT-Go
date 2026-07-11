package notifications

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ScheduleService struct {
	Repository ports.AutoNotificationScheduleRepository
}

func NewScheduleService(repository ports.AutoNotificationScheduleRepository) ScheduleService {
	return ScheduleService{Repository: repository}
}

func (s ScheduleService) List(ctx context.Context, guildID string) ([]domain.AutoNotificationSchedule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Repository == nil {
		return nil, domain.ErrInvalidAutoNotificationSchedule
	}
	guildID = strings.TrimSpace(guildID)
	if err := domain.ValidateAutoNotificationGuildID(guildID); err != nil {
		return nil, err
	}
	schedules, err := s.Repository.ListAutoNotificationSchedules(ctx, guildID)
	if err != nil {
		return nil, err
	}
	active := make([]domain.AutoNotificationSchedule, 0, len(schedules))
	hasPending := false
	for _, schedule := range schedules {
		schedule = schedule.Normalized()
		if schedule.Pending {
			hasPending = true
			continue
		}
		active = append(active, schedule)
	}
	if hasPending {
		if err := s.Repository.DeletePendingAutoNotificationSchedules(ctx, guildID); err != nil {
			return nil, err
		}
	}
	return active, ctx.Err()
}

func (s ScheduleService) Delete(ctx context.Context, guildID string, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAutoNotificationSchedule
	}
	guildID = strings.TrimSpace(guildID)
	id = strings.TrimSpace(id)
	if err := domain.ValidateAutoNotificationDelete(guildID, id); err != nil {
		return err
	}
	return s.Repository.DeleteAutoNotificationSchedule(ctx, guildID, id)
}

func (s ScheduleService) StartSetup(ctx context.Context, draft domain.AutoNotificationSetupDraft) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAutoNotificationSchedule
	}
	draft = draft.Normalized()
	if err := domain.ValidateAutoNotificationSetupDraft(draft); err != nil {
		return err
	}
	schedules, err := s.Repository.ListAutoNotificationSchedules(ctx, draft.GuildID)
	if err != nil {
		return err
	}
	if len(schedules) >= 10 {
		return ports.ErrAutoNotificationScheduleLimit
	}
	return s.Repository.CreatePendingAutoNotificationSchedule(ctx, draft)
}

func (s ScheduleService) CompleteSetup(ctx context.Context, setup domain.AutoNotificationSetup) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAutoNotificationSchedule
	}
	setup = setup.Normalized()
	if err := domain.ValidateAutoNotificationSetup(setup); err != nil {
		return err
	}
	return s.Repository.CompleteAutoNotificationSchedule(ctx, setup)
}
