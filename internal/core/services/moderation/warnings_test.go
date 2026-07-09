package moderation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/moderation"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestWarningHistoryServiceReturnsHistory(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-1",
		Entries: []domain.WarningEntry{{
			ModeratorID: "mod-1",
			Reason:      "reason",
			Time:        "time",
		}},
	})
	service := moderation.WarningHistoryService{Repository: repo}
	got, err := service.History(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(got.Entries) != 1 || got.Entries[0].ModeratorID != "mod-1" {
		t.Fatalf("history = %#v", got)
	}
}

func TestWarningHistoryServiceMissingAndEmpty(t *testing.T) {
	service := moderation.WarningHistoryService{Repository: fakemongo.NewWarningHistoryRepository()}
	if _, err := service.History(context.Background(), "guild-1", "user-1"); !errors.Is(err, ports.ErrWarningsNotFound) {
		t.Fatalf("missing err = %v", err)
	}
	if _, err := service.History(context.Background(), "", "user-1"); !errors.Is(err, domain.ErrInvalidWarningQuery) {
		t.Fatalf("invalid err = %v", err)
	}
}

func TestWarningSettingsServiceConfiguresSettings(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	service := moderation.WarningSettingsService{Repository: repo}
	err := service.Configure(context.Background(), domain.WarningSettings{
		GuildID:   " guild-1 ",
		Threshold: 4,
		Action:    " 踢出 ",
	})
	if err != nil {
		t.Fatalf("configure settings: %v", err)
	}
	got := repo.Settings["guild-1"]
	if got.Threshold != 4 || got.Action != domain.WarningSettingsActionKick {
		t.Fatalf("saved settings = %#v", got)
	}
}

func TestWarningSettingsServiceReadsSettings(t *testing.T) {
	repo := fakemongo.NewWarningSettingsRepository()
	repo.Settings["guild-1"] = domain.WarningSettings{GuildID: "guild-1", Threshold: 2, Action: domain.WarningSettingsActionBan}
	service := moderation.WarningSettingsService{Repository: repo}
	got, err := service.Settings(context.Background(), " guild-1 ")
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if got.Threshold != 2 || got.Action != domain.WarningSettingsActionBan {
		t.Fatalf("settings = %#v", got)
	}
}

func TestWarningSettingsServiceRejectsInvalidAndMissingRepository(t *testing.T) {
	service := moderation.WarningSettingsService{}
	err := service.Configure(context.Background(), domain.WarningSettings{GuildID: "guild-1", Threshold: 1, Action: domain.WarningSettingsActionBan})
	if !errors.Is(err, ports.ErrWarningSettingsUnavailable) {
		t.Fatalf("missing repository err = %v", err)
	}
	service.Repository = fakemongo.NewWarningSettingsRepository()
	err = service.Configure(context.Background(), domain.WarningSettings{GuildID: "guild-1", Threshold: 0, Action: domain.WarningSettingsActionBan})
	if !errors.Is(err, domain.ErrInvalidWarningSettings) {
		t.Fatalf("invalid err = %v", err)
	}
}

func TestWarningIssueServiceAppendsWarnings(t *testing.T) {
	repo := fakemongo.NewWarningHistoryRepository()
	service := moderation.WarningIssueService{Repository: repo}
	first, err := service.Issue(context.Background(), domain.WarningIssue{
		GuildID:     " guild-1 ",
		UserID:      " user-1 ",
		ModeratorID: " mod-1 ",
		Reason:      " spam ",
		Time:        " 2026年07月04日 18點30分 ",
	})
	if err != nil {
		t.Fatalf("first issue: %v", err)
	}
	if !first.Created || len(first.History.Entries) != 1 {
		t.Fatalf("first issue result = %#v", first)
	}
	second, err := service.Issue(context.Background(), domain.WarningIssue{
		GuildID:     "guild-1",
		UserID:      "user-1",
		ModeratorID: "mod-2",
		Reason:      "again",
		Time:        "2026年07月04日 18點31分",
	})
	if err != nil {
		t.Fatalf("second issue: %v", err)
	}
	if second.Created || len(second.History.Entries) != 2 || second.History.Entries[1].ModeratorID != "mod-2" {
		t.Fatalf("second issue result = %#v", second)
	}
}

func TestWarningIssueServiceRejectsInvalidAndMissingRepository(t *testing.T) {
	service := moderation.WarningIssueService{}
	issue := domain.WarningIssue{GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-1", Reason: "spam", Time: "time"}
	if _, err := service.Issue(context.Background(), issue); !errors.Is(err, ports.ErrWarningIssueUnavailable) {
		t.Fatalf("missing repository err = %v", err)
	}
	service.Repository = fakemongo.NewWarningHistoryRepository()
	issue.Reason = ""
	if _, err := service.Issue(context.Background(), issue); !errors.Is(err, domain.ErrInvalidWarningIssue) {
		t.Fatalf("invalid err = %v", err)
	}
}

func TestWarningRemovalServiceRemovesOneAndAll(t *testing.T) {
	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-1",
		Entries: []domain.WarningEntry{
			{ModeratorID: "mod-1", Reason: "one"},
			{ModeratorID: "mod-2", Reason: "two"},
		},
	})
	service := moderation.WarningRemovalService{Repository: repo}
	err := service.RemoveOne(context.Background(), domain.WarningRemoval{GuildID: " guild-1 ", UserID: " user-1 ", Index: 1})
	if err != nil {
		t.Fatalf("remove one: %v", err)
	}
	if got := repo.Histories["guild-1\x00user-1"].Entries; len(got) != 1 || got[0].Reason != "two" {
		t.Fatalf("remaining entries = %#v", got)
	}
	err = service.RemoveAll(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1"})
	if err != nil {
		t.Fatalf("remove all: %v", err)
	}
	if _, ok := repo.Histories["guild-1\x00user-1"]; ok {
		t.Fatalf("expected all warnings removed, histories = %#v", repo.Histories)
	}
}

func TestWarningRemovalServiceRejectsInvalidAndMissingRepository(t *testing.T) {
	service := moderation.WarningRemovalService{}
	err := service.RemoveAll(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1"})
	if !errors.Is(err, ports.ErrWarningRemovalUnavailable) {
		t.Fatalf("missing repository err = %v", err)
	}
	service.Repository = fakemongo.NewWarningRemovalRepository()
	err = service.RemoveOne(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1", Index: 0})
	if !errors.Is(err, domain.ErrInvalidWarningRemoval) {
		t.Fatalf("invalid err = %v", err)
	}
}
