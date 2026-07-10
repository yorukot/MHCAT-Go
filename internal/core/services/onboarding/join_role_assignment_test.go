package onboarding

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestJoinRoleAssignmentServiceAppliesLegacyAudienceRules(t *testing.T) {
	repo := fakeJoinRoleAssignmentRepo{configs: []domain.JoinRoleConfig{
		{GuildID: "guild", RoleID: "all", GiveTo: domain.JoinRoleGiveAllUsers},
		{GuildID: "guild", RoleID: "default"},
		{GuildID: "guild", RoleID: "bot", GiveTo: domain.JoinRoleGiveBots},
		{GuildID: "guild", RoleID: "member", GiveTo: domain.JoinRoleGiveMembers},
		{GuildID: "guild", RoleID: "unknown", GiveTo: "unknown"},
	}}
	roles := &fakeRolePort{}
	service := JoinRoleAssignmentService{Repository: repo, Roles: roles}

	if err := service.AssignOnJoin(context.Background(), "guild", "user", false); err != nil {
		t.Fatalf("assign member: %v", err)
	}
	if got := roles.roleIDs(); len(got) != 3 || got[0] != "all" || got[1] != "default" || got[2] != "member" {
		t.Fatalf("member roles = %#v", got)
	}

	roles.added = nil
	if err := service.AssignOnJoin(context.Background(), "guild", "bot-user", true); err != nil {
		t.Fatalf("assign bot: %v", err)
	}
	if got := roles.roleIDs(); len(got) != 3 || got[0] != "all" || got[1] != "default" || got[2] != "bot" {
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

func TestJoinRoleAssignmentServicePreservesLegacyRoleChecksAndOwnerWarning(t *testing.T) {
	repo := fakeJoinRoleAssignmentRepo{configs: []domain.JoinRoleConfig{
		{GuildID: "guild", RoleID: "missing", GiveTo: domain.JoinRoleGiveAllUsers},
		{GuildID: "guild", RoleID: "too-high", GiveTo: domain.JoinRoleGiveAllUsers},
		{GuildID: "guild", RoleID: "good", GiveTo: domain.JoinRoleGiveAllUsers},
	}}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.MissingRoles["guild/missing"] = true
	sideEffects.AssignableRoles["guild/good"] = true
	guilds := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner"}}
	service := JoinRoleAssignmentService{
		Repository:     repo,
		Roles:          sideEffects,
		RoleInspector:  sideEffects,
		Guilds:         guilds,
		DirectMessages: sideEffects,
	}

	if err := service.AssignOnJoin(context.Background(), "guild", "user", false); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "good" {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "owner" {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	want := "很抱歉，我沒有權限給他加入的成員身分組\n麻煩請將我的身份組位階調高!\n身分組:<@too-high>"
	if message := sideEffects.DirectMessages[0].Message; message.Content != want || !reflect.DeepEqual(message.AllowedMentions, ports.AllowedMentions{}) {
		t.Fatalf("owner warning = %#v", message)
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
