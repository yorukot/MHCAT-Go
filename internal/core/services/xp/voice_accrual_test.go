package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceAccrualServiceTicksJoinedProfile(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 10, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	service := VoiceAccrualService{Repository: repo}

	result, err := service.Tick(context.Background(), " guild-1 ", " user-1 ")
	if err != nil {
		t.Fatalf("tick: %v", err)
	}
	if !result.Active || result.Leveled || result.Gained != LegacyVoiceXPTickAmount || result.Profile.XP != 15 || result.Profile.Level != 0 {
		t.Fatalf("result = %#v", result)
	}
	if saved := repo.VoiceProfiles["guild-1/user-1"]; saved.XP != 15 || saved.Level != 0 || saved.LeaveJoin != domain.VoiceXPSessionJoined {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestVoiceAccrualServicePreservesLegacyLevelUpXP(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	service := VoiceAccrualService{Repository: repo}

	result, err := service.Tick(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("tick: %v", err)
	}
	if !result.Active || !result.Leveled || result.Profile.Level != 1 || result.Profile.XP != LegacyVoiceXPTickAmount {
		t.Fatalf("result = %#v", result)
	}
	if saved := repo.VoiceProfiles["guild-1/user-1"]; saved.Level != 1 || saved.XP != LegacyVoiceXPTickAmount {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestVoiceAccrualServiceSkipsLeftProfile(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 20, Level: 1, LeaveJoin: domain.VoiceXPSessionLeft}
	service := VoiceAccrualService{Repository: repo}

	result, err := service.Tick(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("tick: %v", err)
	}
	if result.Active || result.Gained != 0 || result.Leveled || result.Profile.XP != 20 || result.Profile.Level != 1 {
		t.Fatalf("result = %#v", result)
	}
	if saved := repo.VoiceProfiles["guild-1/user-1"]; saved.XP != 20 || saved.Level != 1 {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestVoiceAccrualServiceRejectsInvalidInput(t *testing.T) {
	service := VoiceAccrualService{Repository: fakemongo.NewXPAdminRepository()}
	if _, err := service.Tick(context.Background(), "", "user-1"); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid adjustment, got %v", err)
	}
	if _, err := (VoiceAccrualService{}).Tick(context.Background(), "guild-1", "user-1"); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid service, got %v", err)
	}
}

func TestVoiceAccrualServiceReturnsMissingProfile(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	service := VoiceAccrualService{Repository: repo}

	if _, err := service.Tick(context.Background(), "guild-1", "user-1"); !errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		t.Fatalf("expected missing profile, got %v", err)
	}
}
