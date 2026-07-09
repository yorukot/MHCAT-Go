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

func TestProfileServicePrepareAddRequiresConfigBeforeDateValidation(t *testing.T) {
	_, err := NewProfileService(&fakemongo.BirthdayConfigRepository{}).PrepareAdd(context.Background(), domain.BirthdayAddRequest{
		GuildID:       "guild-1",
		ActorUserID:   "user-1",
		BirthdayMonth: 13,
		BirthdayDay:   1,
		CurrentYear:   2026,
	})
	if !errors.Is(err, ports.ErrBirthdayConfigMissing) {
		t.Fatalf("expected config missing before invalid month, got %v", err)
	}
}

func TestProfileServicePrepareAddAllowsSelfWhenConfigAllowsEveryone(t *testing.T) {
	repo := birthdayRepoWithConfig(true)
	profile, err := NewProfileService(repo).PrepareAdd(context.Background(), domain.BirthdayAddRequest{
		GuildID:       " guild-1 ",
		ActorUserID:   " user-1 ",
		BirthdayYear:  birthdayInt(2000),
		BirthdayMonth: 7,
		BirthdayDay:   9,
		CurrentYear:   2026,
	})
	if err != nil {
		t.Fatalf("prepare add: %v", err)
	}
	if profile.GuildID != "guild-1" || profile.UserID != "user-1" || profile.BirthdayMonth == nil || *profile.BirthdayMonth != 7 || !profile.AllowAdmin {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestProfileServicePrepareAddRequiresManageMessagesWhenConfigDisallowsEveryone(t *testing.T) {
	repo := birthdayRepoWithConfig(false)
	_, err := NewProfileService(repo).PrepareAdd(context.Background(), domain.BirthdayAddRequest{
		GuildID:       "guild-1",
		ActorUserID:   "user-1",
		TargetUserID:  "user-2",
		BirthdayMonth: 7,
		BirthdayDay:   9,
		CurrentYear:   2026,
	})
	if !errors.Is(err, domain.ErrBirthdayManageMessagesRequired) {
		t.Fatalf("expected manage messages error, got %v", err)
	}
}

func TestProfileServicePrepareAddRejectsNonAdminTargetWhenEveryoneCanSet(t *testing.T) {
	repo := birthdayRepoWithConfig(true)
	_, err := NewProfileService(repo).PrepareAdd(context.Background(), domain.BirthdayAddRequest{
		GuildID:       "guild-1",
		ActorUserID:   "user-1",
		TargetUserID:  "user-2",
		BirthdayMonth: 7,
		BirthdayDay:   9,
		CurrentYear:   2026,
	})
	if !errors.Is(err, domain.ErrBirthdaySelfOnly) {
		t.Fatalf("expected self-only error, got %v", err)
	}
}

func TestProfileServicePrepareAddRejectsAdminWhenTargetDisallows(t *testing.T) {
	repo := birthdayRepoWithConfig(true)
	repo.Profiles = map[string]domain.BirthdayProfile{
		"guild-1/user-2": {GuildID: "guild-1", UserID: "user-2", AllowAdmin: false},
	}
	_, err := NewProfileService(repo).PrepareAdd(context.Background(), domain.BirthdayAddRequest{
		GuildID:                "guild-1",
		ActorUserID:            "user-1",
		TargetUserID:           "user-2",
		ActorCanManageMessages: true,
		BirthdayYear:           birthdayInt(2027),
		BirthdayMonth:          7,
		BirthdayDay:            9,
		CurrentYear:            2026,
	})
	if !errors.Is(err, domain.ErrBirthdayAdminNotAllowed) {
		t.Fatalf("expected admin-not-allowed before future year, got %v", err)
	}
}

func TestProfileServiceSaveDateTimeWritesLegacyFieldsAndAllowTrue(t *testing.T) {
	repo := birthdayRepoWithConfig(true)
	year, month, day := 2000, 7, 9
	profile := domain.BirthdayProfile{
		GuildID:       "guild-1",
		UserID:        "user-1",
		BirthdayYear:  &year,
		BirthdayMonth: &month,
		BirthdayDay:   &day,
		AllowAdmin:    false,
	}
	if err := NewProfileService(repo).SaveDateTime(context.Background(), profile, 8, 30); err != nil {
		t.Fatalf("save date/time: %v", err)
	}
	saved := repo.Profiles["guild-1/user-1"]
	if saved.SendHour == nil || *saved.SendHour != 8 || saved.SendMinute == nil || *saved.SendMinute != 30 || !saved.AllowAdmin {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestProfileServiceSaveDateTimeRejectsInvalidMinute(t *testing.T) {
	month, day := 7, 9
	profile := domain.BirthdayProfile{GuildID: "guild-1", UserID: "user-1", BirthdayMonth: &month, BirthdayDay: &day}
	err := NewProfileService(birthdayRepoWithConfig(true)).SaveDateTime(context.Background(), profile, 8, 31)
	if !errors.Is(err, domain.ErrInvalidBirthdayTime) {
		t.Fatalf("expected invalid time, got %v", err)
	}
}

func birthdayRepoWithConfig(everyoneCanSet bool) *fakemongo.BirthdayConfigRepository {
	return &fakemongo.BirthdayConfigRepository{Configs: map[string]domain.BirthdayConfig{
		"guild-1": {
			GuildID:                    "guild-1",
			Message:                    "{user} 生日快樂",
			UTCOffset:                  "+08:00",
			ChannelID:                  "channel-1",
			EveryoneCanSetBirthdayDate: everyoneCanSet,
		},
	}}
}

func birthdayInt(value int) *int {
	return &value
}
