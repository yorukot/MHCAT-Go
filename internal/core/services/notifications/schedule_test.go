package notifications

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestListReturnsActiveSchedulesAndCleansPendingDrafts(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{
		{GuildID: " guild-1 ", ID: " active-1 ", Cron: " */30 * * * * ", ChannelID: " channel-1 "},
		{GuildID: "guild-1", ID: "draft-1", Pending: true, ChannelID: "channel-2"},
		{GuildID: "guild-1", ID: "active-2", Cron: "0 9 * * 1", ChannelID: "channel-3"},
	}
	service := NewScheduleService(repo)

	schedules, err := service.List(context.Background(), " guild-1 ")
	if err != nil {
		t.Fatalf("list schedules: %v", err)
	}
	if len(schedules) != 2 {
		t.Fatalf("schedules = %#v", schedules)
	}
	if schedules[0].ID != "active-1" || schedules[0].Cron != "*/30 * * * *" || schedules[0].ChannelID != "channel-1" {
		t.Fatalf("first schedule not normalized: %#v", schedules[0])
	}
	if len(repo.PendingCleanupGuilds) != 1 || repo.PendingCleanupGuilds[0] != "guild-1" {
		t.Fatalf("cleanup guilds = %#v", repo.PendingCleanupGuilds)
	}
	for _, schedule := range repo.Schedules["guild-1"] {
		if schedule.Pending {
			t.Fatalf("pending schedule was not cleaned up: %#v", repo.Schedules["guild-1"])
		}
	}
}

func TestDeleteValidatesIdentityAndPropagatesMissing(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	service := NewScheduleService(repo)

	if err := service.Delete(context.Background(), "", "id-1"); !errors.Is(err, domain.ErrInvalidAutoNotificationSchedule) {
		t.Fatalf("expected invalid schedule error, got %v", err)
	}
	if err := service.Delete(context.Background(), "guild-1", "missing"); !errors.Is(err, ports.ErrAutoNotificationScheduleMissing) {
		t.Fatalf("expected missing schedule error, got %v", err)
	}
}
