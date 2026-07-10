package lottery

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
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

func TestComponentModuleRegistersAllLegacyLotteryRoutes(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	now := time.Unix(1_700_000_000, 0)
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{GuildID: "guild-1", ID: lotteryTestID, EndsAtUnix: now.Add(time.Hour).Unix()}
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	module := NewComponentModule(repo, nil, fakediscord.NewSideEffects(), fakediscord.NewSideEffects(), lotteryFixedClock{now: now})
	if commands := module.Commands(); len(commands) != 0 {
		t.Fatalf("component-only module commands = %#v", commands)
	}
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	for _, customID := range []string{lotteryTestID, lotteryTestID + "search"} {
		responder := fakediscord.NewResponder()
		if err := router.Handle(context.Background(), fakediscord.ComponentInteractionFromID(customID), responder); err != nil {
			t.Fatalf("dispatch %s: %v", customID, err)
		}
		if len(responder.Edits) != 1 {
			t.Fatalf("dispatch %s edits = %#v", customID, responder.Edits)
		}
	}
}
