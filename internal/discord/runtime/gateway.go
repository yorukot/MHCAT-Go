package runtime

import "context"

type Gateway interface {
	Open() error
	Close() error
	RegisterInteractionHandler(Handler) func()
	Ready() <-chan struct{}
}

func WaitReady(ctx context.Context, gateway Gateway) error {
	if gateway == nil {
		return ErrRuntimeNotConfigured
	}
	select {
	case <-gateway.Ready():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
