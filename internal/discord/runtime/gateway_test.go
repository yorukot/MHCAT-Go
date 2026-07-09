package runtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func TestWaitReadySuccess(t *testing.T) {
	gateway := &fakeGateway{ready: make(chan struct{})}
	close(gateway.ready)
	if err := discordruntime.WaitReady(context.Background(), gateway); err != nil {
		t.Fatalf("wait ready: %v", err)
	}
}

func TestWaitReadyTimeout(t *testing.T) {
	gateway := &fakeGateway{ready: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	err := discordruntime.WaitReady(ctx, gateway)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

type fakeGateway struct {
	ready chan struct{}
}

func (f *fakeGateway) Open() error  { return nil }
func (f *fakeGateway) Close() error { return nil }
func (f *fakeGateway) RegisterInteractionHandler(discordruntime.Handler) func() {
	return func() {}
}
func (f *fakeGateway) Ready() <-chan struct{} { return f.ready }
