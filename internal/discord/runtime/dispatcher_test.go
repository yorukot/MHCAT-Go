package runtime_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestDispatcherRoutesHelp(t *testing.T) {
	router := interactions.NewRouter()
	if err := router.RegisterSlash("help", func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		return responder.Reply(ctx, responses.Message{Content: "help ok"})
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	dispatcher, err := discordruntime.NewDispatcher(router, nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("help"), responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Content != "help ok" {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestDispatcherUnknownCommandSafeError(t *testing.T) {
	dispatcher, err := discordruntime.NewDispatcher(interactions.NewRouter(), nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	responder := fakediscord.NewResponder()
	err = dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("missing"), responder)
	if !errors.Is(err, interactions.ErrRouteNotFound) {
		t.Fatalf("expected ErrRouteNotFound, got %v", err)
	}
	if len(responder.Errors) != 1 || strings.Contains(responder.Errors[0].Content, "missing") {
		t.Fatalf("safe error = %#v", responder.Errors)
	}
}

func TestDispatcherContextCancellationPropagates(t *testing.T) {
	router := interactions.NewRouter()
	if err := router.RegisterSlash("slow", func(ctx context.Context, _ interactions.Interaction, _ responses.Responder) error {
		<-ctx.Done()
		return ctx.Err()
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	dispatcher, err := discordruntime.NewDispatcher(router, nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	err = dispatcher.Dispatch(ctx, fakediscord.SlashInteraction("slow"), fakediscord.NewResponder())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
