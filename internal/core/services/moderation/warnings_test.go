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
