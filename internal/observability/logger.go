package observability

import (
	"context"
	"io"
	"log/slog"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

type LoggerOptions struct {
	Level  string
	Format string
	Writer io.Writer
}

func NewLogger(opts LoggerOptions) *slog.Logger {
	writer := opts.Writer
	if writer == nil {
		writer = io.Discard
	}

	level := slog.LevelInfo
	switch opts.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handlerOpts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if opts.Format == "json" {
		handler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		handler = slog.NewTextHandler(writer, handlerOpts)
	}
	return slog.New(redactingHandler{handler: handler})
}

type redactingHandler struct {
	handler slog.Handler
}

func (h redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	redacted := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		redacted.AddAttrs(redactAttr(attr))
		return true
	})
	return h.handler.Handle(ctx, redacted)
}

func (h redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		redacted = append(redacted, redactAttr(attr))
	}
	return redactingHandler{handler: h.handler.WithAttrs(redacted)}
}

func (h redactingHandler) WithGroup(name string) slog.Handler {
	return redactingHandler{handler: h.handler.WithGroup(name)}
}

func redactAttr(attr slog.Attr) slog.Attr {
	if attr.Value.Kind() == slog.KindGroup {
		group := attr.Value.Group()
		redacted := make([]slog.Attr, 0, len(group))
		for _, nested := range group {
			redacted = append(redacted, redactAttr(nested))
		}
		return slog.Group(attr.Key, attrsToAny(redacted)...)
	}
	if attr.Value.Kind() == slog.KindString {
		return slog.String(attr.Key, config.RedactValue(attr.Key, attr.Value.String()))
	}
	return attr
}

func attrsToAny(attrs []slog.Attr) []any {
	values := make([]any, len(attrs))
	for i, attr := range attrs {
		values[i] = attr
	}
	return values
}
