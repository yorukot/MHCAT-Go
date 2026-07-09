package interactions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var (
	ErrPanicRecovered   = errors.New("interaction handler panic recovered")
	ErrPermissionDenied = errors.New("interaction permission denied")
)

type PermissionChecker interface {
	Check(ctx context.Context, actor Actor, route Route) error
}

type PermissionCheckerFunc func(ctx context.Context, actor Actor, route Route) error

func (f PermissionCheckerFunc) Check(ctx context.Context, actor Actor, route Route) error {
	return f(ctx, actor, route)
}

func AllowAllPermissions() PermissionChecker {
	return PermissionCheckerFunc(func(context.Context, Actor, Route) error {
		return nil
	})
}

func Timeout(duration time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, interaction Interaction, responder responses.Responder) error {
			if duration <= 0 {
				return next(ctx, interaction, responder)
			}
			timedCtx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			return next(timedCtx, interaction, responder)
		}
	}
}

func Recover() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, interaction Interaction, responder responses.Responder) (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					if responder != nil {
						_ = responder.Error(ctx, fmt.Errorf("%w: %v", ErrPanicRecovered, recovered))
					}
					err = fmt.Errorf("%w: %v", ErrPanicRecovered, recovered)
				}
			}()
			return next(ctx, interaction, responder)
		}
	}
}

func Logging(logger *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, interaction Interaction, responder responses.Responder) error {
			if logger != nil {
				logger.InfoContext(ctx, "interaction route start", "route", interaction.Route().String())
			}
			err := next(ctx, interaction, responder)
			if logger != nil {
				if err != nil {
					logger.WarnContext(ctx, "interaction route failed", "route", interaction.Route().String(), "error", err.Error())
				} else {
					logger.InfoContext(ctx, "interaction route finished", "route", interaction.Route().String())
				}
			}
			return err
		}
	}
}

func Permission(checker PermissionChecker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, interaction Interaction, responder responses.Responder) error {
			if checker == nil {
				return fmt.Errorf("%w: checker is not configured", ErrPermissionDenied)
			}
			if err := checker.Check(ctx, interaction.Actor, interaction.Route()); err != nil {
				return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
			}
			return next(ctx, interaction, responder)
		}
	}
}

func Usage(tracker ports.UsageTracker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, interaction Interaction, responder responses.Responder) error {
			err := next(ctx, interaction, responder)
			if err != nil || tracker == nil || interaction.CommandName == "" {
				return err
			}
			if trackErr := tracker.TrackCommand(ctx, ports.UsageEvent{
				CommandName: interaction.CommandName,
				UserID:      interaction.Actor.UserID,
				GuildID:     interaction.Actor.GuildID,
				Feature:     routeFeature(interaction),
			}); trackErr != nil {
				return trackErr
			}
			return nil
		}
	}
}

func routeFeature(interaction Interaction) string {
	if !interaction.RouteKey.IsZero() {
		return interaction.RouteKey.Feature
	}
	if interaction.CommandName == "help" || interaction.CommandName == "info" || interaction.CommandName == "ping" || interaction.CommandName == "翻譯" {
		return "utility"
	}
	return ""
}
