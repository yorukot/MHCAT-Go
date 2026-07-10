package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestVerificationConfigServiceSavesAssignableRole(t *testing.T) {
	repo := &fakeVerificationConfigRepo{}
	roles := &fakeVerificationRoleInspector{assignable: true}
	service := VerificationConfigService{Repository: repo, RoleInspector: roles}
	err := service.Save(context.Background(), domain.VerificationConfig{
		GuildID:        " guild ",
		RoleID:         " role ",
		RenameTemplate: " {name} | MHCAT ",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if repo.saved.GuildID != "guild" || repo.saved.RoleID != "role" || repo.saved.RenameTemplate != " {name} | MHCAT " {
		t.Fatalf("saved = %#v", repo.saved)
	}
}

func TestVerificationConfigServiceRejectsUnassignableRole(t *testing.T) {
	repo := &fakeVerificationConfigRepo{}
	roles := &fakeVerificationRoleInspector{}
	service := VerificationConfigService{Repository: repo, RoleInspector: roles}
	err := service.Save(context.Background(), domain.VerificationConfig{GuildID: "guild", RoleID: "role"})
	if !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("expected unassignable, got %v", err)
	}
	if repo.saved.GuildID != "" {
		t.Fatalf("unexpected save: %#v", repo.saved)
	}
}

type fakeVerificationConfigRepo struct {
	saved domain.VerificationConfig
}

func (r *fakeVerificationConfigRepo) SaveVerificationConfig(ctx context.Context, config domain.VerificationConfig) error {
	r.saved = config
	return ctx.Err()
}

func (r *fakeVerificationConfigRepo) GetVerificationConfig(ctx context.Context, guildID string) (domain.VerificationConfig, error) {
	return r.saved, ctx.Err()
}

type fakeVerificationRoleInspector struct {
	assignable bool
}

func (r *fakeVerificationRoleInspector) CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	return r.assignable, ctx.Err()
}
