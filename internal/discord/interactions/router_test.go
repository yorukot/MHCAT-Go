package interactions_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestRouterSlashCommandRoutesByExactName(t *testing.T) {
	router := interactions.NewRouter()
	called := false
	if err := router.RegisterSlash("ping", func(context.Context, interactions.Interaction, responses.Responder) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register slash: %v", err)
	}
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction("ping"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRouterUnknownSlashCommandReturnsNotFound(t *testing.T) {
	router := interactions.NewRouter()
	err := router.Handle(context.Background(), fakediscord.SlashInteraction("missing"), fakediscord.NewResponder())
	if !errors.Is(err, interactions.ErrRouteNotFound) {
		t.Fatalf("expected ErrRouteNotFound, got %v", err)
	}
}

func TestRouterComponentUsesParsedKey(t *testing.T) {
	router := interactions.NewRouter()
	key := interactions.ComponentKey{Version: "v1", Feature: "ticket", Action: "close"}
	called := false
	if err := router.RegisterComponent(key, func(context.Context, interactions.Interaction, responses.Responder) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register component: %v", err)
	}
	if err := router.Handle(context.Background(), fakediscord.ComponentInteraction(key), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatal("component handler was not called")
	}
}

func TestRouterModalUsesParsedKey(t *testing.T) {
	router := interactions.NewRouter()
	key := interactions.ModalKey{Version: "v1", Feature: "ticket", Action: "rename"}
	called := false
	if err := router.RegisterModal(key, func(context.Context, interactions.Interaction, responses.Responder) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register modal: %v", err)
	}
	if err := router.Handle(context.Background(), fakediscord.ModalInteraction(key), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatal("modal handler was not called")
	}
}
