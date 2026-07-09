package birthday

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestProfileServiceSetAllowAdminCreatesPlaceholderProfile(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	service := NewProfileService(repo)

	if err := service.SetAllowAdmin(context.Background(), " guild-1 ", " user-1 ", false); err != nil {
		t.Fatalf("set allow admin: %v", err)
	}
	saved := repo.Profiles["guild-1/user-1"]
	if saved.GuildID != "guild-1" || saved.UserID != "user-1" || saved.AllowAdmin {
		t.Fatalf("saved = %#v", saved)
	}
	if saved.BirthdayYear != nil || saved.BirthdayMonth != nil || saved.BirthdayDay != nil || saved.SendHour != nil || saved.SendMinute != nil {
		t.Fatalf("placeholder should preserve null date/time fields: %#v", saved)
	}
}

func TestProfileServiceSetAllowAdminPreservesBirthdayFields(t *testing.T) {
	year, month, day, hour, minute := 2000, 12, 31, 8, 30
	repo := &fakemongo.BirthdayConfigRepository{Profiles: map[string]domain.BirthdayProfile{
		"guild-1/user-1": {
			GuildID: "guild-1", UserID: "user-1",
			BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day,
			SendHour: &hour, SendMinute: &minute,
			AllowAdmin: false,
		},
	}}

	if err := NewProfileService(repo).SetAllowAdmin(context.Background(), "guild-1", "user-1", true); err != nil {
		t.Fatalf("set allow admin: %v", err)
	}
	saved := repo.Profiles["guild-1/user-1"]
	if !saved.AllowAdmin || saved.BirthdayYear == nil || *saved.BirthdayYear != 2000 || saved.SendMinute == nil || *saved.SendMinute != 30 {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestProfileServiceDeleteMissingProfile(t *testing.T) {
	err := NewProfileService(&fakemongo.BirthdayConfigRepository{}).Delete(context.Background(), "guild-1", "user-1")
	if !errors.Is(err, ports.ErrBirthdayProfileMissing) {
		t.Fatalf("expected ErrBirthdayProfileMissing, got %v", err)
	}
}

func TestProfileServiceListRejectsEmptyGuild(t *testing.T) {
	_, err := NewProfileService(&fakemongo.BirthdayConfigRepository{}).List(context.Background(), " ")
	if !errors.Is(err, domain.ErrInvalidBirthdayProfile) {
		t.Fatalf("expected ErrInvalidBirthdayProfile, got %v", err)
	}
}
