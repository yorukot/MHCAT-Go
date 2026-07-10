package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceSessionServiceMarksJoinAndLeave(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	service := VoiceSessionService{Repository: repo}

	if err := service.Join(context.Background(), " guild-1 ", " user-1 "); err != nil {
		t.Fatalf("join: %v", err)
	}
	profile := repo.VoiceProfiles["guild-1/user-1"]
	if profile.GuildID != "guild-1" || profile.UserID != "user-1" || profile.XP != 0 || profile.Level != 0 || profile.LeaveJoin != domain.VoiceXPSessionJoined {
		t.Fatalf("joined profile = %#v", profile)
	}

	if err := service.Leave(context.Background(), "guild-1", "user-1"); err != nil {
		t.Fatalf("leave: %v", err)
	}
	profile = repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionLeft {
		t.Fatalf("left profile = %#v", profile)
	}
}

func TestVoiceSessionServiceRejectsInvalidIDs(t *testing.T) {
	service := VoiceSessionService{Repository: fakemongo.NewXPAdminRepository()}

	if err := service.Join(context.Background(), "", "user-1"); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid join, got %v", err)
	}
	if err := (VoiceSessionService{}).Leave(context.Background(), "guild-1", "user-1"); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid leave, got %v", err)
	}
}
