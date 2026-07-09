package utility_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestUtilityHandlerPanicRecoveryStillWorks(t *testing.T) {
	router := interactions.NewRouter(interactions.Recover())
	if err := router.RegisterSlash("boom", func(context.Context, interactions.Interaction, responses.Responder) error {
		panic("boom")
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	err := router.Handle(context.Background(), fakediscord.SlashInteraction("boom"), fakediscord.NewResponder())
	if !errors.Is(err, interactions.ErrPanicRecovered) {
		t.Fatalf("expected panic recovery, got %v", err)
	}
}

func TestUtilityModuleRoutesThroughPermissionMiddleware(t *testing.T) {
	called := false
	checker := interactions.PermissionCheckerFunc(func(context.Context, interactions.Actor, interactions.Route) error {
		called = true
		return nil
	})
	router := interactions.NewRouter(interactions.Permission(checker))
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction("ping"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatal("permission checker was not invoked")
	}
}
