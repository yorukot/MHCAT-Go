package economy

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestCoinGameTimeoutManagerKeepsNewerGeneration(t *testing.T) {
	manager := newCoinGameTimeoutManager(ports.SystemClock{})
	t.Cleanup(func() { _ = manager.Stop(context.Background()) })
	fired := make(chan string, 2)
	now := time.Now()
	manager.Schedule("game-1", 2, now.Add(20*time.Millisecond), func(context.Context) {
		fired <- "new"
	})
	manager.Schedule("game-1", 1, now.Add(time.Millisecond), func(context.Context) {
		fired <- "old"
	})

	select {
	case got := <-fired:
		if got != "new" {
			t.Fatalf("fired generation = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("newer timeout did not fire")
	}
	select {
	case got := <-fired:
		t.Fatalf("unexpected second timeout = %q", got)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestCoinGameTimeoutManagerReplacesOlderGeneration(t *testing.T) {
	manager := newCoinGameTimeoutManager(ports.SystemClock{})
	t.Cleanup(func() { _ = manager.Stop(context.Background()) })
	fired := make(chan uint64, 2)
	now := time.Now()
	manager.Schedule("game-1", 1, now.Add(time.Second), func(context.Context) {
		fired <- 1
	})
	manager.Schedule("game-1", 2, now.Add(time.Millisecond), func(context.Context) {
		fired <- 2
	})

	select {
	case generation := <-fired:
		if generation != 2 {
			t.Fatalf("fired generation = %d", generation)
		}
	case <-time.After(time.Second):
		t.Fatal("replacement timeout did not fire")
	}
}

func TestCoinGameTimeoutManagerShutdownCancelsAndWaits(t *testing.T) {
	manager := newCoinGameTimeoutManager(ports.SystemClock{})
	started := make(chan struct{})
	finished := make(chan struct{})
	manager.Schedule("running", 1, time.Now(), func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		close(finished)
	})
	manager.Schedule("future", 1, time.Now().Add(time.Hour), func(context.Context) {
		t.Error("future timeout fired after shutdown")
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("running timeout did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := manager.Stop(ctx); err != nil {
		t.Fatalf("stop timeout manager: %v", err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("stop returned before running timeout finished")
	}
	if err := manager.Stop(context.Background()); err != nil {
		t.Fatalf("second stop: %v", err)
	}
}
