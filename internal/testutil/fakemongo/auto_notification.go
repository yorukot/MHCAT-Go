package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AutoNotificationScheduleRepository struct {
	Schedules            map[string][]domain.AutoNotificationSchedule
	Deleted              []AutoNotificationScheduleDelete
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
