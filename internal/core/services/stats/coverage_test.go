package stats

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRenameWorkerRunOnceSafeHandlesServiceError(t *testing.T) {
	worker := NewRenameWorker(RenameService{}, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	worker.runOnceSafe(context.Background())
}
