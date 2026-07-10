package ports

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidUsageEvent = errors.New("invalid usage event")

type UsageEvent struct {
	CommandName string
	UserID      string
	GuildID     string
	Feature     string
}

func (e UsageEvent) Validate() error {
	if strings.TrimSpace(e.CommandName) == "" {
		return ErrInvalidUsageEvent
	}
	return nil
}

type UsageTracker interface {
	TrackCommand(ctx context.Context, event UsageEvent) error
}
