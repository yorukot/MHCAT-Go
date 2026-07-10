package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrAutoNotificationScheduleMissing = errors.New("auto-notification schedule is missing")
var ErrAutoNotificationScheduleExists = errors.New("auto-notification schedule already exists")
var ErrAutoNotificationScheduleLimit = errors.New("auto-notification schedule limit reached")

type AutoNotificationScheduleRepository interface {
	ListAutoNotificationSchedules(ctx context.Context, guildID string) ([]domain.AutoNotificationSchedule, error)
	CreatePendingAutoNotificationSchedule(ctx context.Context, draft domain.AutoNotificationSetupDraft) error
	CompleteAutoNotificationSchedule(ctx context.Context, setup domain.AutoNotificationSetup) error
	DeleteAutoNotificationSchedule(ctx context.Context, guildID string, id string) error
	DeletePendingAutoNotificationSchedules(ctx context.Context, guildID string) error
}

type AutoNotificationDeliveryRepository interface {
	ListAutoNotificationDeliveries(ctx context.Context) ([]domain.AutoNotificationSchedule, error)
	GetAutoNotificationDelivery(ctx context.Context, guildID string, id string) (domain.AutoNotificationSchedule, error)
}
