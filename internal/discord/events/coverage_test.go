package events

import (
	"context"
	"errors"
	"testing"
)

func TestContinueOnErrorStringAndDispatchSafe(t *testing.T) {
	continued := ContinueOnError(errors.New("continued"))
	if continued == nil || continued.Error() != "continued" {
		t.Fatalf("continued error = %v", continued)
	}
	dispatcher := NewDispatcher(nil)
	dispatcher.Register(TypeReady, func(context.Context, Event) error { return errors.New("failure") })
	dispatcher.DispatchSafe(context.Background(), Event{Type: TypeReady})
	dispatcher.DispatchSafe(context.Background(), Event{Type: TypeMessageCreate})
}
