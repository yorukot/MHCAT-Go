package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
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

func TestResetServiceDeletesIndividualProfiles(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 10, Level: 1}
	repo.VoiceProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 20, Level: 2}
	service := ResetService{Repository: repo}

	if err := service.ResetTextProfile(context.Background(), " guild-1 ", " user-1 "); err != nil {
		t.Fatalf("reset text profile: %v", err)
	}
	if _, ok := repo.TextProfiles["guild-1/user-1"]; ok {
		t.Fatal("text profile was not deleted")
	}
	if err := service.ResetVoiceProfile(context.Background(), "guild-1", "user-2"); err != nil {
		t.Fatalf("reset voice profile: %v", err)
	}
	if _, ok := repo.VoiceProfiles["guild-1/user-2"]; ok {
		t.Fatal("voice profile was not deleted")
	}
}

func TestResetServiceReturnsMissingForAbsentProfile(t *testing.T) {
	service := ResetService{Repository: fakemongo.NewXPAdminRepository()}

	err := service.ResetTextProfile(context.Background(), "guild-1", "user-1")
	if !errors.Is(err, ports.ErrTextXPProfileMissing) {
		t.Fatalf("expected missing text profile, got %v", err)
	}
	err = service.ResetVoiceProfile(context.Background(), "guild-1", "user-1")
	if !errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		t.Fatalf("expected missing voice profile, got %v", err)
	}
}

func TestResetServiceDeletesGuildProfiles(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1"}
	repo.TextProfiles["guild-2/user-2"] = domain.XPProfile{GuildID: "guild-2", UserID: "user-2"}
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1"}
	service := ResetService{Repository: repo}

	if err := service.ResetTextGuild(context.Background(), " guild-1 "); err != nil {
		t.Fatalf("reset text guild: %v", err)
	}
	if _, ok := repo.TextProfiles["guild-1/user-1"]; ok {
		t.Fatal("guild text profile was not deleted")
	}
	if _, ok := repo.TextProfiles["guild-2/user-2"]; !ok {
		t.Fatal("other guild text profile was deleted")
	}
	if err := service.ResetVoiceGuild(context.Background(), "guild-1"); err != nil {
		t.Fatalf("reset voice guild: %v", err)
	}
	if _, ok := repo.VoiceProfiles["guild-1/user-1"]; ok {
		t.Fatal("guild voice profile was not deleted")
	}
}

func TestResetServicePreservesLegacyEmptyGuildAsymmetry(t *testing.T) {
	service := ResetService{Repository: fakemongo.NewXPAdminRepository()}

	if err := service.ResetTextGuild(context.Background(), "guild-1"); err != nil {
		t.Fatalf("empty text reset: %v", err)
	}
	if err := service.ResetVoiceGuild(context.Background(), "guild-1"); !errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		t.Fatalf("empty voice reset error = %v", err)
	}
}
