package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func TestOpenWithTimeout(t *testing.T) {
	session := newBlockingDiscord()
	err := openWithTimeout(context.Background(), session, time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if session.closes != 1 {
		t.Fatalf("session closes = %d", session.closes)
	}
}

func TestRunCanceledDuringGatewayOpenShutsDownCleanly(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordEnableGateway = true
	mongo := &fakeMongo{}
	discord := newBlockingDiscord()
	application, err := New(
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithMongoFactory(func(config.Config) (MongoClient, error) { return mongo, nil }),
		WithDiscordFactory(func(config.Config) (DiscordSession, error) { return discord, nil }),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	select {
	case <-discord.opened:
	case <-time.After(time.Second):
		t.Fatal("gateway did not start")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run app: %v", err)
	}
	if discord.closes != 1 || mongo.disconnects != 1 {
		t.Fatalf("closes=%d disconnects=%d", discord.closes, mongo.disconnects)
	}
}

type blockingDiscord struct {
	fakeDiscord
	release chan struct{}
	once    sync.Once
}

func (b *blockingDiscord) Open() error {
	b.opens++
	close(b.opened)
	<-b.release
	return nil
}

func (b *blockingDiscord) Close() error {
	b.closes++
	b.once.Do(func() { close(b.release) })
	return nil
}

func newBlockingDiscord() *blockingDiscord {
	return &blockingDiscord{
		fakeDiscord: fakeDiscord{ready: make(chan struct{}), opened: make(chan struct{})},
		release:     make(chan struct{}),
	}
}
