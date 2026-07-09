package lottery

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestModuleRegistersCreateRoute(t *testing.T) {
	router := interactions.NewRouter()
	module := NewModule(nil)
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction(LotteryCreateCommandName), responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != legacyUnavailableTitle {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
