package app

import (
	"context"
	"fmt"
	"io"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/observability"
)

func Run(ctx context.Context, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := observability.NewLogger(observability.LoggerOptions{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Writer: stdout,
	})

	application, err := New(cfg, logger)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	return application.Run(ctx)
}
