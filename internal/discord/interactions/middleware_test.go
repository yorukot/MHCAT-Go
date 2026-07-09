package interactions_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestMiddlewareOrderDeterministic(t *testing.T) {
	var order []string
	handler := func(context.Context, interactions.Interaction, responses.Responder) error {
		order = append(order, "handler")
		return nil
	}
	first := func(next interactions.Handler) interactions.Handler {
		return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
			order = append(order, "first-before")
			err := next(ctx, interaction, responder)
			order = append(order, "first-after")
			return err
		}
	}
	second := func(next interactions.Handler) interactions.Handler {
		return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
			order = append(order, "second-before")
			err := next(ctx, interaction, responder)
			order = append(order, "second-after")
			return err
		}
	}

	err := interactions.Chain(handler, first, second)(context.Background(), fakediscord.SlashInteraction("ping"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("chain: %v", err)
	}
	want := []string{"first-before", "second-before", "handler", "second-after", "first-after"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	handler := func(context.Context, interactions.Interaction, responses.Responder) error {
		panic("boom")
	}
	responder := fakediscord.NewResponder()
	err := interactions.Chain(handler, interactions.Recover())(context.Background(), fakediscord.SlashInteraction("panic"), responder)
	if !errors.Is(err, interactions.ErrPanicRecovered) {
		t.Fatalf("expected ErrPanicRecovered, got %v", err)
	}
	if len(responder.Errors) != 1 {
		t.Fatalf("expected safe error response, got %d", len(responder.Errors))
	}
}

func TestTimeoutMiddlewareAppliesContext(t *testing.T) {
	handler := func(ctx context.Context, _ interactions.Interaction, _ responses.Responder) error {
		<-ctx.Done()
		return ctx.Err()
	}
	err := interactions.Chain(handler, interactions.Timeout(time.Nanosecond))(context.Background(), fakediscord.SlashInteraction("slow"), fakediscord.NewResponder())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestPermissionCheckerInvoked(t *testing.T) {
	called := false
	checker := interactions.PermissionCheckerFunc(func(context.Context, interactions.Actor, interactions.Route) error {
		called = true
		return nil
	})
	handler := func(context.Context, interactions.Interaction, responses.Responder) error {
		return nil
	}
	err := interactions.Chain(handler, interactions.Permission(checker))(context.Background(), fakediscord.SlashInteraction("secure"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("permission middleware: %v", err)
	}
	if !called {
		t.Fatal("permission checker was not called")
	}
}

func TestUsageTrackerInvokedOnSuccess(t *testing.T) {
	tracker := &fakeusage.Tracker{}
	handler := func(context.Context, interactions.Interaction, responses.Responder) error {
		return nil
	}
	err := interactions.Chain(handler, interactions.Usage(tracker))(context.Background(), fakediscord.SlashInteraction("ping"), fakediscord.NewResponder())
	if err != nil {
		t.Fatalf("usage middleware: %v", err)
	}
	if len(tracker.Events) != 1 || tracker.Events[0].CommandName != "ping" {
		t.Fatalf("events = %#v", tracker.Events)
	}
}

func TestUsageTrackerNotInvokedOnFailure(t *testing.T) {
	tracker := &fakeusage.Tracker{}
	handler := func(context.Context, interactions.Interaction, responses.Responder) error {
		return errors.New("handler failed")
	}
	_ = interactions.Chain(handler, interactions.Usage(tracker))(context.Background(), fakediscord.SlashInteraction("ping"), fakediscord.NewResponder())
	if len(tracker.Events) != 0 {
		t.Fatalf("usage tracked failed command: %#v", tracker.Events)
	}
}
