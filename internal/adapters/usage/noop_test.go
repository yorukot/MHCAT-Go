package usage_test

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/usage"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestNoopTrackerNeverFails(t *testing.T) {
	tracker := usage.NoopTracker{}
	if err := tracker.TrackCommand(context.Background(), ports.UsageEvent{CommandName: "ping"}); err != nil {
		t.Fatalf("track command: %v", err)
	}
}
