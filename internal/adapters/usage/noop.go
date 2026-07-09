package usage

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type NoopTracker struct{}

func (NoopTracker) TrackCommand(context.Context, ports.UsageEvent) error {
	return nil
}
