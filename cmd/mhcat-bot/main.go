package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/app"
)

func main() {
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	os.Exit(run(ctx, os.Stdout, os.Stderr, app.Run))
}

type appRunner func(context.Context, io.Writer) error

func run(ctx context.Context, stdout io.Writer, stderr io.Writer, runApp appRunner) int {
	if runApp == nil {
		fmt.Fprintln(stderr, "mhcat-bot: app runner is required")
		return 1
	}
	if err := runApp(ctx, stdout); err != nil {
		fmt.Fprintf(stderr, "mhcat-bot: %v\n", err)
		return 1
	}
	return 0
}
