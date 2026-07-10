package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/app"
)

func main() {
	ctx, stop := app.SignalContext(context.Background())
	defer stop()

	if err := app.Run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "mhcat-bot: %v\n", err)
		os.Exit(1)
	}
}
