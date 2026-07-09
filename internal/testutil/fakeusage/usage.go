package fakeusage

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type Tracker struct {
	Events []ports.UsageEvent
	Err    error
}

func (t *Tracker) TrackCommand(ctx context.Context, event ports.UsageEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if t.Err != nil {
		return t.Err
	}
	t.Events = append(t.Events, event)
	return nil
}
