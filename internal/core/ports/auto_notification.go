package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrAutoNotificationScheduleMissing = errors.New("auto-notification schedule is missing")

type AutoNotificationScheduleRepository interface {
	ListAutoNotificationSchedules(ctx context.Context, guildID string) ([]domain.AutoNotificationSchedule, error)
	DeleteAutoNotificationSchedule(ctx context.Context, guildID string, id string) error
	DeletePendingAutoNotificationSchedules(ctx context.Context, guildID string) error
}
