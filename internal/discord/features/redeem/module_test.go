package redeem

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestModuleRegistersRoute(t *testing.T) {
	router := interactions.NewRouter()
	if err := NewModule(nil, nil, nil).RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction(CommandName), responder); err != nil {
		t.Fatalf("dispatch redeem route: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("expected redeem route to handle interaction, defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}
