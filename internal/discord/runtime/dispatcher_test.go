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

func TestDispatcherUnknownCommandUsesLegacyFallback(t *testing.T) {
	dispatcher, err := discordruntime.NewDispatcher(interactions.NewRouter(), nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("missing")
	interaction.Actor.AvatarURL = "https://cdn.test/avatar.png"
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); !errors.Is(err, interactions.ErrRouteNotFound) {
		t.Fatalf("dispatch fallback error: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Errors) != 0 {
		t.Fatalf("replies=%#v errors=%#v", responder.Replies, responder.Errors)
	}
	reply := responder.Replies[0]
	if len(reply.Embeds) != 1 || reply.Embeds[0].Title != "<a:error:980086028113182730> | 很抱歉，這個指令已不再支援或進行改名!" || reply.Embeds[0].Color != 0xA6FFA6 {
		t.Fatalf("fallback embed = %#v", reply.Embeds)
	}
	if reply.Embeds[0].Footer == nil || reply.Embeds[0].Footer.Text != "非常抱歉造成你的困擾，推薦使用/help進行查詢指令" || reply.Embeds[0].Footer.IconURL != interaction.Actor.AvatarURL {
		t.Fatalf("fallback footer = %#v", reply.Embeds[0].Footer)
	}
	if reply.AllowedMentions == nil {
		t.Fatal("fallback must suppress mentions")
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

func TestDispatcherShutdownHooksRunInReverseAndOnlyOnce(t *testing.T) {
	dispatcher, err := discordruntime.NewDispatcher(interactions.NewRouter(), nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	var calls []int
	dispatcher.RegisterShutdown(func(context.Context) error {
		calls = append(calls, 1)
		return errors.New("first")
	})
	dispatcher.RegisterShutdown(func(context.Context) error {
		calls = append(calls, 2)
		return errors.New("second")
	})

	shutdownErr := dispatcher.Shutdown(context.Background())
	if shutdownErr == nil || !strings.Contains(shutdownErr.Error(), "first") || !strings.Contains(shutdownErr.Error(), "second") {
		t.Fatalf("shutdown error = %v", shutdownErr)
	}
	if err := dispatcher.Shutdown(context.Background()); err == nil {
		t.Fatal("second shutdown should return the stored error")
	}
	if len(calls) != 2 || calls[0] != 2 || calls[1] != 1 {
		t.Fatalf("shutdown calls = %v", calls)
	}
}
