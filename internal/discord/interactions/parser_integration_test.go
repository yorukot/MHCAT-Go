package interactions_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeinteractions"
)

func TestRouterComponentRouteUsesParsedRouteKey(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	key := interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "ticket", Action: "close"}
	var got interactions.RouteKey
	if err := router.RegisterRoute(key, func(_ context.Context, interaction interactions.Interaction, _ responses.Responder) error {
		got = interaction.RouteKey
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	err := router.Handle(context.Background(), fakeinteractions.Component("mhcat:v1:ticket:close:"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if got != key {
		t.Fatalf("route key = %#v, want %#v", got, key)
	}
}

func TestRouterModalRouteUsesParsedRouteKey(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	key := interactions.RouteKey{Kind: interactions.TypeModal, Version: "v1", Feature: "ticket", Action: "rename"}
	called := false
	if err := router.RegisterRoute(key, func(context.Context, interactions.Interaction, responses.Responder) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	err := router.Handle(context.Background(), fakeinteractions.Modal("mhcat:v1:ticket:rename:state=abc123"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatal("modal handler was not called")
	}
}

func TestRouterUnknownComponentParseErrorReturnsSafeError(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	responder := fakediscord.NewResponder()
	raw := "unknown:id"
	err := router.Handle(context.Background(), fakeinteractions.Component(raw), responder)
	if !errors.Is(err, interactions.ErrBadInteractionID) {
		t.Fatalf("expected ErrBadInteractionID, got %v", err)
	}
	if len(responder.Errors) != 1 {
		t.Fatalf("expected safe responder error, got %d", len(responder.Errors))
	}
	if strings.Contains(responder.Errors[0].Content, raw) {
		t.Fatalf("safe error leaked custom id: %q", responder.Errors[0].Content)
	}
}

func TestRouterUnknownModalParseErrorReturnsSafeError(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	responder := fakediscord.NewResponder()
	raw := "unknown-modal"
	err := router.Handle(context.Background(), fakeinteractions.Modal(raw, customid.ModalField{CustomID: "unknown"}), responder)
	if !errors.Is(err, interactions.ErrBadInteractionID) {
		t.Fatalf("expected ErrBadInteractionID, got %v", err)
	}
	if len(responder.Errors) != 1 {
		t.Fatalf("expected safe responder error, got %d", len(responder.Errors))
	}
	if strings.Contains(responder.Errors[0].Content, raw) {
		t.Fatalf("safe error leaked modal id: %q", responder.Errors[0].Content)
	}
}

func TestRouterUnknownParsedRouteReturnsNotFound(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	err := router.Handle(context.Background(), fakeinteractions.Component("mhcat:v1:ticket:close:"), fakediscord.NewResponder())
	if !errors.Is(err, interactions.ErrRouteNotFound) {
		t.Fatalf("expected ErrRouteNotFound, got %v", err)
	}
}

func TestRouterMiddlewareStillAppliesAfterParsing(t *testing.T) {
	var order []string
	router := interactions.NewRouter(
		func(next interactions.Handler) interactions.Handler {
			return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
				order = append(order, "first")
				return next(ctx, interaction, responder)
			}
		},
		interactions.Timeout(time.Second),
		interactions.Permission(interactions.PermissionCheckerFunc(func(_ context.Context, _ interactions.Actor, route interactions.Route) error {
			order = append(order, route.RouteKey.String())
			return nil
		})),
		interactions.Recover(),
	)
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	key := interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "ticket", Action: "close"}
	if err := router.RegisterRoute(key, func(context.Context, interactions.Interaction, responses.Responder) error {
		order = append(order, "handler")
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	err := router.Handle(context.Background(), fakeinteractions.Component("mhcat:v1:ticket:close:"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	want := []string{"first", "component:v1:ticket:close:false", "handler"}
	if strings.Join(order, "|") != strings.Join(want, "|") {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestRouterPanicRecoveryStillWorksAfterParsing(t *testing.T) {
	router := interactions.NewRouter(interactions.Recover())
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	key := interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "ticket", Action: "close"}
	if err := router.RegisterRoute(key, func(context.Context, interactions.Interaction, responses.Responder) error {
		panic("boom")
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	err := router.Handle(context.Background(), fakeinteractions.Component("mhcat:v1:ticket:close:"), fakediscord.NewResponder())
	if !errors.Is(err, interactions.ErrPanicRecovered) {
		t.Fatalf("expected ErrPanicRecovered, got %v", err)
	}
}
