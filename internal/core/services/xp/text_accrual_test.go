package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextAccrualServiceCreatesAndUpdatesProfiles(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	service := TextAccrualService{Repository: repo, RandomMultiplier: fixedMultiplier(500)}

	result, err := service.AccrueMessage(context.Background(), " guild-1 ", " user-1 ", "hello")
	if err != nil {
		t.Fatalf("accrue create: %v", err)
	}
	if result.Gained != 5 || result.Leveled || result.Profile.XP != 5 || result.Profile.Level != 0 {
		t.Fatalf("create result = %#v", result)
	}

	result, err = service.AccrueMessage(context.Background(), "guild-1", "user-1", "abc")
	if err != nil {
		t.Fatalf("accrue update: %v", err)
	}
	if result.Gained != 3 || result.Leveled || result.Profile.XP != 8 || result.Profile.Level != 0 {
		t.Fatalf("update result = %#v", result)
	}
}

func TestTextAccrualServiceResetsXPOnLegacyLevelUp(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	service := TextAccrualService{Repository: repo, RandomMultiplier: fixedMultiplier(500)}

	result, err := service.AccrueMessage(context.Background(), "guild-1", "user-1", "hello")
	if err != nil {
		t.Fatalf("accrue: %v", err)
	}
	if !result.Leveled || result.Gained != 5 || result.Profile.Level != 1 || result.Profile.XP != 0 {
		t.Fatalf("level result = %#v", result)
	}
	saved := repo.TextProfiles["guild-1/user-1"]
	if saved.Level != 1 || saved.XP != 0 {
		t.Fatalf("saved profile = %#v", saved)
	}
}

func TestTextAccrualServiceRejectsInvalidInput(t *testing.T) {
	service := TextAccrualService{Repository: fakemongo.NewXPAdminRepository()}
	_, err := service.AccrueMessage(context.Background(), "", "user-1", "hello")
	if !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid adjustment, got %v", err)
	}
	_, err = (TextAccrualService{}).AccrueMessage(context.Background(), "guild-1", "user-1", "hello")
	if !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid service, got %v", err)
	}
}

func TestTextAccrualServiceReturnsRepositoryErrors(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.Err = ports.ErrTextXPProfileMissing
	service := TextAccrualService{Repository: repo, RandomMultiplier: fixedMultiplier(500)}

	_, err := service.AccrueMessage(context.Background(), "guild-1", "user-1", "hello")
	if !errors.Is(err, ports.ErrTextXPProfileMissing) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestTextAccrualServiceUsesAtomicRepositoryHotPath(t *testing.T) {
	repo := &atomicTextXPRepository{
		profile: domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 5},
	}
	service := TextAccrualService{Repository: repo, RandomMultiplier: fixedMultiplier(500)}
	result, err := service.AccrueMessage(context.Background(), "guild-1", "user-1", "hello")
	if err != nil {
		t.Fatalf("accrue: %v", err)
	}
	if repo.atomicCalls != 1 || repo.getCalls != 0 || repo.saveCalls != 0 {
		t.Fatalf("atomic=%d get=%d save=%d", repo.atomicCalls, repo.getCalls, repo.saveCalls)
	}
	if result.Gained != 5 || result.Profile.XP != 5 {
		t.Fatalf("result = %#v", result)
	}
}

func TestLegacyTextXPForMessageMatchesLegacyLengthAndMultiplier(t *testing.T) {
	if got := LegacyTextXPContentLength("ab你🙂"); got != 8 {
		t.Fatalf("legacy length = %d", got)
	}
	if got := LegacyTextXPForMessage("hello", 500); got != 5 {
		t.Fatalf("xp = %d", got)
	}
	if got := LegacyTextXPForMessage("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", 800); got != 80 {
		t.Fatalf("capped xp = %d", got)
	}
}

func TestLegacyTextXPLevelUpAnnouncementUsesDefaultAndFirstReplacements(t *testing.T) {
	if got := LegacyTextXPLevelUpAnnouncement("", 2, " user-1 "); got != "🆙恭喜<@user-1> 的聊天等級成功升級到 2" {
		t.Fatalf("default announcement = %q", got)
	}
	got := LegacyTextXPLevelUpAnnouncement("(user) {user} (leavel) {level} (user) {level}", 7, "user-1")
	want := "<@user-1> <@user-1> 7 7 (user) {level}"
	if got != want {
		t.Fatalf("custom announcement = %q, want %q", got, want)
	}
}

func fixedMultiplier(value int64) func() int64 {
	return func() int64 { return value }
}

type atomicTextXPRepository struct {
	profile     domain.XPProfile
	err         error
	atomicCalls int
	getCalls    int
	saveCalls   int
}

func (r *atomicTextXPRepository) AccrueTextXP(context.Context, string, string, int64) (domain.XPProfile, bool, error) {
	r.atomicCalls++
	return r.profile, false, r.err
}

func (r *atomicTextXPRepository) GetTextXPProfile(context.Context, string, string) (domain.XPProfile, error) {
	r.getCalls++
	return domain.XPProfile{}, errors.New("legacy get should not run")
}

func (r *atomicTextXPRepository) SaveTextXPProfile(context.Context, domain.XPProfile) error {
	r.saveCalls++
	return errors.New("legacy save should not run")
}
