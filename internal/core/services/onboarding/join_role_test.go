package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestJoinRoleServiceCreateNormalizesAndChecksRole(t *testing.T) {
	repo := &fakeJoinRoleRepo{}
	inspector := fakeRoleInspector{assignable: true}
	service := JoinRoleService{Repository: repo, RoleInspector: inspector}
	err := service.Create(context.Background(), domain.JoinRoleConfig{
		GuildID: " guild ",
		RoleID:  " role ",
		GiveTo:  "",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if repo.created.GuildID != "guild" || repo.created.RoleID != "role" || repo.created.GiveTo != domain.JoinRoleGiveAllUsers {
		t.Fatalf("created = %#v", repo.created)
	}
}

func TestJoinRoleServiceCreateRejectsUnassignableRole(t *testing.T) {
	service := JoinRoleService{Repository: &fakeJoinRoleRepo{}, RoleInspector: fakeRoleInspector{}}
	err := service.Create(context.Background(), domain.JoinRoleConfig{GuildID: "guild", RoleID: "role"})
	if !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("expected unassignable role, got %v", err)
	}
}

func TestJoinRoleServiceDelete(t *testing.T) {
	repo := &fakeJoinRoleRepo{}
	service := JoinRoleService{Repository: repo}
	if err := service.Delete(context.Background(), " guild ", " role "); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if repo.deletedGuild != "guild" || repo.deletedRole != "role" {
		t.Fatalf("delete fields = %q/%q", repo.deletedGuild, repo.deletedRole)
	}
}

type fakeJoinRoleRepo struct {
	created      domain.JoinRoleConfig
	deletedGuild string
	deletedRole  string
}

func (r *fakeJoinRoleRepo) CreateJoinRoleConfig(ctx context.Context, config domain.JoinRoleConfig) error {
	r.created = config
	return ctx.Err()
}

func (r *fakeJoinRoleRepo) ListJoinRoleConfigs(ctx context.Context, guildID string) ([]domain.JoinRoleConfig, error) {
	return nil, ctx.Err()
}

func (r *fakeJoinRoleRepo) DeleteJoinRoleConfig(ctx context.Context, guildID string, roleID string) error {
	r.deletedGuild = guildID
	r.deletedRole = roleID
	return ctx.Err()
}

type fakeRoleInspector struct {
	assignable bool
	err        error
}

func (i fakeRoleInspector) CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	if i.err != nil {
		return false, i.err
	}
	return i.assignable, ctx.Err()
}
