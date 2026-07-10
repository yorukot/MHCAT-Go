package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

type Handler func(context.Context, interactions.Interaction, responses.Responder) error
type ShutdownFunc func(context.Context) error

type Dispatcher struct {
	router       *interactions.Router
	logger       *slog.Logger
	mu           sync.RWMutex
	shutdowns    []ShutdownFunc
	shutdownOnce sync.Once
	shutdownErr  error
}

func NewDispatcher(router *interactions.Router, logger *slog.Logger) (*Dispatcher, error) {
	if router == nil {
		return nil, fmt.Errorf("%w: router is required", ErrRuntimeNotConfigured)
	}
	return &Dispatcher{router: router, logger: logger}, nil
}

func (d *Dispatcher) Handler() Handler {
	return d.Dispatch
}

func (d *Dispatcher) RegisterShutdown(fn ShutdownFunc) {
	if d == nil || fn == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.shutdowns = append(d.shutdowns, fn)
}

func (d *Dispatcher) Shutdown(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.shutdownOnce.Do(func() {
		d.mu.RLock()
		shutdowns := append([]ShutdownFunc(nil), d.shutdowns...)
		d.mu.RUnlock()
		var errs []error
		for index := len(shutdowns) - 1; index >= 0; index-- {
			if shutdowns[index] == nil {
				continue
			}
			if err := shutdowns[index](ctx); err != nil {
				errs = append(errs, err)
			}
		}
		d.shutdownErr = errors.Join(errs...)
	})
	return d.shutdownErr
}

func (d *Dispatcher) Dispatch(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if d == nil || d.router == nil {
		return fmt.Errorf("%w: dispatcher is nil", ErrRuntimeNotConfigured)
	}
	if responder == nil {
		return fmt.Errorf("%w: responder is required", ErrRuntimeNotConfigured)
	}
	tracked := &trackingResponder{next: responder}
	if d.logger != nil {
		d.logger.InfoContext(ctx, "runtime interaction dispatch", "route", interaction.Route().String(), "command", interaction.CommandName)
	}
	err := d.router.Handle(ctx, interaction, tracked)
	if err != nil {
		if !tracked.responded {
			_ = tracked.Error(ctx, err)
		}
		if d.logger != nil {
			d.logger.WarnContext(ctx, "runtime interaction failed", "route", interaction.Route().String(), "command", interaction.CommandName, "error", err.Error())
		}
		return err
	}
	if d.logger != nil {
		d.logger.InfoContext(ctx, "runtime interaction finished", "route", interaction.Route().String(), "command", interaction.CommandName)
	}
	return nil
}

type trackingResponder struct {
	next      responses.Responder
	responded bool
}

func (r *trackingResponder) Reply(ctx context.Context, msg responses.Message) error {
	err := r.next.Reply(ctx, msg)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) Defer(ctx context.Context, opts responses.DeferOptions) error {
	err := r.next.Defer(ctx, opts)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) DeferUpdate(ctx context.Context) error {
	err := r.next.DeferUpdate(ctx)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) ShowModal(ctx context.Context, modal responses.Modal) error {
	err := r.next.ShowModal(ctx, modal)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) UpdateMessage(ctx context.Context, msg responses.Message) error {
	err := r.next.UpdateMessage(ctx, msg)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) EditOriginal(ctx context.Context, msg responses.Message) error {
	err := r.next.EditOriginal(ctx, msg)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) FollowUp(ctx context.Context, msg responses.Message) error {
	err := r.next.FollowUp(ctx, msg)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) CreateFollowUp(ctx context.Context, msg responses.Message) (string, error) {
	messageID, err := r.next.CreateFollowUp(ctx, msg)
	if err == nil {
		r.responded = true
	}
	return messageID, err
}

func (r *trackingResponder) EditFollowUp(ctx context.Context, messageID string, msg responses.Message) error {
	err := r.next.EditFollowUp(ctx, messageID, msg)
	if err == nil {
		r.responded = true
	}
	return err
}

func (r *trackingResponder) DeleteFollowUp(ctx context.Context, messageID string) error {
	return r.next.DeleteFollowUp(ctx, messageID)
}

func (r *trackingResponder) Error(ctx context.Context, err error) error {
	responseErr := r.next.Error(ctx, err)
	if responseErr == nil {
		r.responded = true
	}
	return responseErr
}
