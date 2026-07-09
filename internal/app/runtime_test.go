package app

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOpenWithTimeout(t *testing.T) {
	session := &blockingDiscord{fakeDiscord: fakeDiscord{ready: make(chan struct{})}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	err := openWithTimeout(ctx, session, time.Nanosecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

type blockingDiscord struct {
	fakeDiscord
}

func (b *blockingDiscord) Open() error {
	select {}
}
