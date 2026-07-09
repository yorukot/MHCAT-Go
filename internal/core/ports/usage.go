package ports

import "context"

type UsageEvent struct {
	CommandName string
	UserID      string
	GuildID     string
	Feature     string
}

type UsageTracker interface {
	TrackCommand(ctx context.Context, event UsageEvent) error
}
