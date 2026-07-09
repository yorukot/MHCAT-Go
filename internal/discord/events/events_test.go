package events_test

import (
	"context"
	"errors"
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
