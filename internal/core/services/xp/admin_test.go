package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAdminServiceAddsTextXPAndCreatesMissingProfile(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	service := AdminService{Repository: repo}

	profile, err := service.AddTextXP(context.Background(), domain.XPAdjustment{GuildID: " guild-1 ", UserID: " user-1 ", Delta: 150})
	if err != nil {
		t.Fatalf("add text xp: %v", err)
	}
	if profile.GuildID != "guild-1" || profile.UserID != "user-1" || profile.Level != 1 || profile.XP != 50 {
		t.Fatalf("profile = %#v", profile)
	}
	saved, ok := repo.TextProfiles["guild-1/user-1"]
	if !ok || saved != profile {
		t.Fatalf("saved profile = %#v ok=%v", saved, ok)
	}
}

func TestAdminServiceAddsVoiceXPToExistingProfile(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", Level: 2, XP: 50}
	service := AdminService{Repository: repo}

	profile, err := service.AddVoiceXP(context.Background(), domain.XPAdjustment{GuildID: "guild-1", UserID: "user-1", Delta: 500})
	if err != nil {
		t.Fatalf("add voice xp: %v", err)
	}
	if profile.Level != 3 || profile.XP != 250 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestAdminServiceRejectsInvalidAdjustment(t *testing.T) {
	service := AdminService{Repository: fakemongo.NewXPAdminRepository()}
	_, err := service.AddTextXP(context.Background(), domain.XPAdjustment{GuildID: "guild-1", Delta: 1})
	if !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid adjustment, got %v", err)
	}
}
