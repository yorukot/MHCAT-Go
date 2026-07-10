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
	if schedules[0].ID != "active-1" || schedules[0].Cron != " */30 * * * * " || schedules[0].ChannelID != "channel-1" {
		t.Fatalf("first schedule normalization = %#v", schedules[0])
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

func TestStartSetupCreatesPendingDraft(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	service := NewScheduleService(repo)

	err := service.StartSetup(context.Background(), domain.AutoNotificationSetupDraft{
		GuildID:   " guild-1 ",
		ID:        " 1700000000000 ",
		ChannelID: " channel-1 ",
	})
	if err != nil {
		t.Fatalf("start setup: %v", err)
	}
	schedules := repo.Schedules["guild-1"]
	if len(schedules) != 1 || schedules[0].ID != "1700000000000" || schedules[0].ChannelID != "channel-1" || !schedules[0].Pending {
		t.Fatalf("schedules = %#v", schedules)
	}
}

func TestStartSetupRejectsScheduleLimit(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	for index := 0; index < 10; index++ {
		repo.Schedules["guild-1"] = append(repo.Schedules["guild-1"], domain.AutoNotificationSchedule{GuildID: "guild-1", ID: string(rune('a' + index)), Cron: "0 9 * * 1"})
	}
	service := NewScheduleService(repo)

	err := service.StartSetup(context.Background(), domain.AutoNotificationSetupDraft{GuildID: "guild-1", ID: "new", ChannelID: "channel-1"})
	if !errors.Is(err, ports.ErrAutoNotificationScheduleLimit) {
		t.Fatalf("expected schedule limit, got %v", err)
	}
	if len(repo.Schedules["guild-1"]) != 10 {
		t.Fatalf("schedule count changed: %#v", repo.Schedules["guild-1"])
	}
}

func TestCompleteSetupUpdatesPendingDraft(t *testing.T) {
	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "1700000000000", ChannelID: "channel-1", Pending: true}}
	service := NewScheduleService(repo)

	err := service.CompleteSetup(context.Background(), domain.AutoNotificationSetup{
		GuildID: "guild-1",
		ID:      "1700000000000",
		Cron:    "*/30 * * * *",
		Message: domain.AutoNotificationMessage{Content: "hello"},
	})
	if err != nil {
		t.Fatalf("complete setup: %v", err)
	}
	if len(repo.Completed) != 1 || repo.Completed[0].Cron != "*/30 * * * *" || repo.Schedules["guild-1"][0].Pending {
		t.Fatalf("completed=%#v schedules=%#v", repo.Completed, repo.Schedules["guild-1"])
	}
}
