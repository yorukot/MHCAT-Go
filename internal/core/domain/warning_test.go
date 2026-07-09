package domain_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestWarningSettingsValidate(t *testing.T) {
	valid := domain.WarningSettings{GuildID: "guild-1", Threshold: 3, Action: domain.WarningSettingsActionBan}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid settings: %v", err)
	}
	for _, settings := range []domain.WarningSettings{
		{GuildID: "", Threshold: 3, Action: domain.WarningSettingsActionBan},
		{GuildID: "guild-1", Threshold: 0, Action: domain.WarningSettingsActionBan},
		{GuildID: "guild-1", Threshold: 3, Action: "mute"},
	} {
		if err := settings.Validate(); !errors.Is(err, domain.ErrInvalidWarningSettings) {
			t.Fatalf("settings %#v err = %v", settings, err)
		}
	}
}

func TestWarningIssueValidate(t *testing.T) {
	valid := domain.WarningIssue{GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-1", Reason: "spam", Time: "2026年07月04日 18點30分"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid issue: %v", err)
	}
	for _, issue := range []domain.WarningIssue{
		{GuildID: "", UserID: "user-1", ModeratorID: "mod-1", Reason: "spam", Time: "time"},
		{GuildID: "guild-1", UserID: "", ModeratorID: "mod-1", Reason: "spam", Time: "time"},
		{GuildID: "guild-1", UserID: "user-1", ModeratorID: "", Reason: "spam", Time: "time"},
		{GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-1", Reason: "", Time: "time"},
		{GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-1", Reason: "spam", Time: ""},
	} {
		if err := issue.Validate(); !errors.Is(err, domain.ErrInvalidWarningIssue) {
			t.Fatalf("issue %#v err = %v", issue, err)
		}
	}
}

func TestWarningRemovalValidate(t *testing.T) {
	valid := domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1", Index: 1}
	if err := valid.ValidateSingle(); err != nil {
		t.Fatalf("valid single removal: %v", err)
	}
	if err := (domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1"}).ValidateAll(); err != nil {
		t.Fatalf("valid all removal: %v", err)
	}
	for _, removal := range []domain.WarningRemoval{
		{GuildID: "", UserID: "user-1", Index: 1},
		{GuildID: "guild-1", UserID: "", Index: 1},
		{GuildID: "guild-1", UserID: "user-1", Index: 0},
	} {
		if err := removal.ValidateSingle(); !errors.Is(err, domain.ErrInvalidWarningRemoval) {
			t.Fatalf("single removal %#v err = %v", removal, err)
		}
	}
	if err := (domain.WarningRemoval{GuildID: "guild-1"}).ValidateAll(); !errors.Is(err, domain.ErrInvalidWarningRemoval) {
		t.Fatalf("all removal err = %v", err)
	}
}
