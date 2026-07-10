package discordgo

import (
	"context"
	"errors"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestCachedRoleInspectorUsesGuildRoleAndBotMemberState(t *testing.T) {
	state := dgo.NewState()
	state.User = &dgo.User{ID: "bot-1", Bot: true}
	if err := state.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Roles: []*dgo.Role{
			{ID: "guild-1", Position: 0},
			{ID: "role-low", Position: 1},
			{ID: "role-bot", Position: 2},
			{ID: "role-high", Position: 3},
		},
		Members: []*dgo.Member{{User: state.User, Roles: []string{"role-bot"}}},
	}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	inspector := NewCachedRoleInspector(SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}})

	assignable, err := inspector.CanAssignRole(context.Background(), "guild-1", "role-low")
	if err != nil || !assignable {
		t.Fatalf("low role: assignable=%t err=%v", assignable, err)
	}
	assignable, err = inspector.CanAssignRole(context.Background(), "guild-1", "role-high")
	if err != nil || assignable {
		t.Fatalf("high role: assignable=%t err=%v", assignable, err)
	}
	_, err = inspector.CanAssignRole(context.Background(), "guild-1", "missing")
	if !errors.Is(err, ports.ErrDiscordRoleMissing) {
		t.Fatalf("missing role error = %v", err)
	}
}

func TestCachedRoleInspectorRequiresCachedBotMember(t *testing.T) {
	state := dgo.NewState()
	state.User = &dgo.User{ID: "bot-1", Bot: true}
	if err := state.GuildAdd(&dgo.Guild{ID: "guild-1", Roles: []*dgo.Role{{ID: "role-1", Position: 1}}}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	inspector := NewCachedRoleInspector(SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}})
	if _, err := inspector.CanAssignRole(context.Background(), "guild-1", "role-1"); err == nil {
		t.Fatal("expected missing cached bot member error")
	}
}
