package utility_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestPingHandlerRepliesOnceAndTracksUsage(t *testing.T) {
	tracker := &fakeusage.Tracker{}
	clock := fixedClock{now: time.Unix(100, 50_000_000)}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, clock, tracker)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionCreatedAt("ping", time.Unix(100, 0))
	if err := module.PingHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Content, "Pong!") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	if len(tracker.Events) != 1 || tracker.Events[0].CommandName != "ping" {
		t.Fatalf("usage events = %#v", tracker.Events)
	}
}

func TestPingHandlerRespectsContextCancellation(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := module.PingHandler()(ctx, fakediscord.SlashInteraction("ping"), fakediscord.NewResponder())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

var _ ports.Clock = fixedClock{}
