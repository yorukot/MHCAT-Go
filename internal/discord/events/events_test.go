package events_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func TestDispatcherRoutesRegisteredHandler(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	var got events.Event
	dispatcher.Register(events.TypeMessageCreate, func(ctx context.Context, event events.Event) error {
		got = event
		return nil
	})
	err := dispatcher.Dispatch(context.Background(), events.Event{Type: events.TypeMessageCreate, MessageID: "message-1"})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if got.MessageID != "message-1" {
		t.Fatalf("event = %#v", got)
	}
}

func TestDispatcherUnknownHandler(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	err := dispatcher.Dispatch(context.Background(), events.Event{Type: events.TypeVoiceState})
	if !errors.Is(err, events.ErrNoHandler) {
		t.Fatalf("expected ErrNoHandler, got %v", err)
	}
}

func TestDispatcherContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dispatcher := events.NewDispatcher(nil)
	err := dispatcher.Dispatch(ctx, events.Event{Type: events.TypeReady})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestDispatcherStopPropagationDoesNotRunLaterHandlers(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	calls := 0
	dispatcher.Register(events.TypeMemberAdd, func(ctx context.Context, event events.Event) error {
		calls++
		return events.ErrStopPropagation
	})
	dispatcher.Register(events.TypeMemberAdd, func(ctx context.Context, event events.Event) error {
		calls++
		return nil
	})
	if err := dispatcher.Dispatch(context.Background(), events.Event{Type: events.TypeMemberAdd}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDispatcherContinueOnErrorRunsLaterHandlersAndReturnsError(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	wantErr := errors.New("best-effort handler failed")
	calls := 0
	dispatcher.Register(events.TypeMemberAdd, func(ctx context.Context, event events.Event) error {
		calls++
		return events.ContinueOnError(wantErr)
	})
	dispatcher.Register(events.TypeMemberAdd, func(ctx context.Context, event events.Event) error {
		calls++
		return nil
	})
	err := dispatcher.Dispatch(context.Background(), events.Event{Type: events.TypeMemberAdd})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped failure", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestDispatcherShutdownRunsRegisteredCallbacksOnceInReverseOrder(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	var calls []string
	dispatcher.RegisterShutdown(func(context.Context) error {
		calls = append(calls, "first")
		return nil
	})
	dispatcher.RegisterShutdown(func(context.Context) error {
		calls = append(calls, "second")
		return errors.New("shutdown failed")
	})

	err := dispatcher.Shutdown(context.Background())
	if err == nil || !strings.Contains(err.Error(), "shutdown failed") {
		t.Fatalf("expected shutdown error, got %v", err)
	}
	if err := dispatcher.Shutdown(context.Background()); err == nil || !strings.Contains(err.Error(), "shutdown failed") {
		t.Fatalf("expected cached shutdown error, got %v", err)
	}
	if len(calls) != 2 || calls[0] != "second" || calls[1] != "first" {
		t.Fatalf("calls = %#v", calls)
	}
}
