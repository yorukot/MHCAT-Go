package onboarding

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestJoinRoleAssignmentHandlerAssignsMatchingRoles(t *testing.T) {
	repo := fakemongo.NewJoinRoleConfigRepository()
	repo.Configs["guild-1/all"] = domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "all", GiveTo: domain.JoinRoleGiveAllUsers}
	repo.Configs["guild-1/member"] = domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "member", GiveTo: domain.JoinRoleGiveMembers}
	repo.Configs["guild-1/bot"] = domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "bot", GiveTo: domain.JoinRoleGiveBots}
	sideEffects := fakediscord.NewSideEffects()
	module := NewJoinRoleAssignmentModule(repo, sideEffects, nil, nil, nil)

	err := module.JoinRoleAssignmentHandler()(context.Background(), events.Event{
		Type:    events.TypeMemberAdd,
		GuildID: "guild-1",
		UserID:  "user-1",
		Member:  &events.Member{UserID: "user-1", IsBot: false},
	})
	if err != nil {
		t.Fatalf("assign member roles: %v", err)
	}
	if !hasRole(sideEffects.AddedRoles, "all") || !hasRole(sideEffects.AddedRoles, "member") || hasRole(sideEffects.AddedRoles, "bot") {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
}

func TestJoinRoleAssignmentHandlerIgnoresNonMemberAddAndIncompleteEvents(t *testing.T) {
	repo := fakemongo.NewJoinRoleConfigRepository()
	repo.Configs["guild-1/all"] = domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "all", GiveTo: domain.JoinRoleGiveAllUsers}
	sideEffects := fakediscord.NewSideEffects()
	module := NewJoinRoleAssignmentModule(repo, sideEffects, nil, nil, nil)

	for _, event := range []events.Event{
		{Type: events.TypeMemberRemove, GuildID: "guild-1", UserID: "user-1"},
		{Type: events.TypeMemberAdd, UserID: "user-1"},
		{Type: events.TypeMemberAdd, GuildID: "guild-1"},
	} {
		if err := module.JoinRoleAssignmentHandler()(context.Background(), event); err != nil {
			t.Fatalf("event %#v: %v", event, err)
		}
	}
	if len(sideEffects.AddedRoles) != 0 {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
}

func TestJoinRoleAssignmentEventRouteRegistration(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewJoinRoleAssignmentModule(fakemongo.NewJoinRoleConfigRepository(), fakediscord.NewSideEffects(), nil, nil, nil).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMemberAdd) {
		t.Fatal("expected member-add handler")
	}

	empty := events.NewDispatcher(nil)
	Module{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMemberAdd) {
		t.Fatal("unexpected member-add handler for empty module")
	}
}

func hasRole(changes []fakediscord.RoleChange, roleID string) bool {
	for _, change := range changes {
		if change.RoleID == roleID {
			return true
		}
	}
	return false
}
