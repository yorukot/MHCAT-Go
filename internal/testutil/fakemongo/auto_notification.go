package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AutoNotificationScheduleRepository struct {
	Schedules            map[string][]domain.AutoNotificationSchedule
	Deleted              []AutoNotificationScheduleDelete
	Completed            []domain.AutoNotificationSetup
	PendingCleanupGuilds []string
	Err                  error
}

type AutoNotificationScheduleDelete struct {
	GuildID string
	ID      string
}

func NewAutoNotificationScheduleRepository() *AutoNotificationScheduleRepository {
	return &AutoNotificationScheduleRepository{Schedules: map[string][]domain.AutoNotificationSchedule{}}
}

func (r *AutoNotificationScheduleRepository) ListAutoNotificationSchedules(ctx context.Context, guildID string) ([]domain.AutoNotificationSchedule, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return append([]domain.AutoNotificationSchedule(nil), r.Schedules[guildID]...), nil
}

func (r *AutoNotificationScheduleRepository) DeleteAutoNotificationSchedule(ctx context.Context, guildID string, id string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	schedules := r.Schedules[guildID]
	for index, schedule := range schedules {
		if schedule.ID == id {
			r.Schedules[guildID] = append(append([]domain.AutoNotificationSchedule(nil), schedules[:index]...), schedules[index+1:]...)
			r.Deleted = append(r.Deleted, AutoNotificationScheduleDelete{GuildID: guildID, ID: id})
			return nil
		}
	}
	return ports.ErrAutoNotificationScheduleMissing
}

func (r *AutoNotificationScheduleRepository) CreatePendingAutoNotificationSchedule(ctx context.Context, draft domain.AutoNotificationSetupDraft) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	draft = draft.Normalized()
	if err := domain.ValidateAutoNotificationSetupDraft(draft); err != nil {
		return err
	}
	for _, schedule := range r.Schedules[draft.GuildID] {
		if schedule.ID == draft.ID {
			return ports.ErrAutoNotificationScheduleExists
		}
	}
	r.Schedules[draft.GuildID] = append(r.Schedules[draft.GuildID], domain.AutoNotificationSchedule{
		GuildID:   draft.GuildID,
		ID:        draft.ID,
		ChannelID: draft.ChannelID,
		Pending:   true,
	})
	return nil
}

func (r *AutoNotificationScheduleRepository) CompleteAutoNotificationSchedule(ctx context.Context, setup domain.AutoNotificationSetup) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	setup = setup.Normalized()
	if err := domain.ValidateAutoNotificationSetup(setup); err != nil {
		return err
	}
	schedules := r.Schedules[setup.GuildID]
	for index, schedule := range schedules {
		if schedule.ID == setup.ID {
			schedules[index].Cron = setup.Cron
			schedules[index].Pending = false
			r.Schedules[setup.GuildID] = schedules
			r.Completed = append(r.Completed, setup)
			return nil
		}
	}
	return ports.ErrAutoNotificationScheduleMissing
}

func (r *AutoNotificationScheduleRepository) DeletePendingAutoNotificationSchedules(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.PendingCleanupGuilds = append(r.PendingCleanupGuilds, guildID)
	schedules := r.Schedules[guildID]
	active := make([]domain.AutoNotificationSchedule, 0, len(schedules))
	for _, schedule := range schedules {
		if !schedule.Pending {
			active = append(active, schedule)
		}
	}
	r.Schedules[guildID] = active
	return nil
}

func (r *AutoNotificationScheduleRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	return nil
}

var _ ports.AutoNotificationScheduleRepository = (*AutoNotificationScheduleRepository)(nil)
