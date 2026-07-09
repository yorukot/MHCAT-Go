package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestJoinRoleAssignmentServiceAppliesLegacyAudienceRules(t *testing.T) {
	repo := fakeJoinRoleAssignmentRepo{configs: []domain.JoinRoleConfig{
		{GuildID: "guild", RoleID: "all", GiveTo: domain.JoinRoleGiveAllUsers},
		{GuildID: "guild", RoleID: "bot", GiveTo: domain.JoinRoleGiveBots},
		{GuildID: "guild", RoleID: "member", GiveTo: domain.JoinRoleGiveMembers},
	}}
	roles := &fakeRolePort{}
	service := JoinRoleAssignmentService{Repository: repo, Roles: roles}

	if err := service.AssignOnJoin(context.Background(), "guild", "user", false); err != nil {
		t.Fatalf("assign member: %v", err)
	}
	if got := roles.roleIDs(); len(got) != 2 || got[0] != "all" || got[1] != "member" {
		t.Fatalf("member roles = %#v", got)
	}

	roles.added = nil
	if err := service.AssignOnJoin(context.Background(), "guild", "bot-user", true); err != nil {
		t.Fatalf("assign bot: %v", err)
	}
	if got := roles.roleIDs(); len(got) != 2 || got[0] != "all" || got[1] != "bot" {
		t.Fatalf("bot roles = %#v", got)
	}
}

func TestJoinRoleAssignmentServiceContinuesAfterRoleError(t *testing.T) {
	repo := fakeJoinRoleAssignmentRepo{configs: []domain.JoinRoleConfig{
		{GuildID: "guild", RoleID: "bad", GiveTo: domain.JoinRoleGiveAllUsers},
		{GuildID: "guild", RoleID: "good", GiveTo: domain.JoinRoleGiveAllUsers},
	}}
	roles := &fakeRolePort{failRole: "bad"}
	service := JoinRoleAssignmentService{Repository: repo, Roles: roles}

	err := service.AssignOnJoin(context.Background(), "guild", "user", false)
	if err == nil {
		t.Fatal("expected joined assignment error")
	}
	if got := roles.roleIDs(); len(got) != 2 || got[0] != "bad" || got[1] != "good" {
		t.Fatalf("roles = %#v", got)
	}
}

type fakeJoinRoleAssignmentRepo struct {
	configs []domain.JoinRoleConfig
	err     error
}

func (r fakeJoinRoleAssignmentRepo) ListJoinRoleConfigs(ctx context.Context, guildID string) ([]domain.JoinRoleConfig, error) {
	if r.err != nil {
		return nil, r.err
	}
	return append([]domain.JoinRoleConfig(nil), r.configs...), ctx.Err()
}

type fakeRolePort struct {
	added    []domain.JoinRoleConfig
	failRole string
}

func (p *fakeRolePort) AddRole(ctx context.Context, guildID string, userID string, roleID string) error {
	p.added = append(p.added, domain.JoinRoleConfig{GuildID: guildID, RoleID: roleID})
	if p.failRole == roleID {
		return errors.New("role add failed")
	}
	return ctx.Err()
}

func (p *fakeRolePort) RemoveRole(ctx context.Context, guildID string, userID string, roleID string) error {
	return ctx.Err()
}

func (p *fakeRolePort) roleIDs() []string {
	ids := make([]string, 0, len(p.added))
	for _, role := range p.added {
		ids = append(ids, role.RoleID)
	}
	return ids
}
