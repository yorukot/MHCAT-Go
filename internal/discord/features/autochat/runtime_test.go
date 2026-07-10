package autochat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAutoChatFallbackMessageCreateRepliesWithTypingDelay(t *testing.T) {
	module, sideEffects := autoChatRuntimeFixture(t, "-1")
	var waited time.Duration
	module.randomInt = func(maximum int) int { return 1250 }
	module.wait = func(_ context.Context, delay time.Duration) error {
		waited = delay
		return nil
	}

	err := module.MessageCreateHandler()(context.Background(), autoChatMessageEvent("你好"))
	if err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.TypingChannels) != 1 || sideEffects.TypingChannels[0] != "channel-1" {
		t.Fatalf("typing = %#v", sideEffects.TypingChannels)
	}
	if waited != 2250*time.Millisecond {
		t.Fatalf("waited = %s", waited)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	message := sideEffects.Sent[0].Message
	if message.Content != "你好，有甚麼我能幫忙的嗎?" || message.ReplyToMessageID != "message-1" {
		t.Fatalf("message = %#v", message)
	}
	if message.AllowedMentions.ParseUsers || message.AllowedMentions.ParseRoles || message.AllowedMentions.ParseEveryone || message.AllowedMentions.RepliedUser {
		t.Fatalf("allowed mentions = %#v", message.AllowedMentions)
	}
}

func TestAutoChatFallbackSpecialReplyIsImmediate(t *testing.T) {
	module, sideEffects := autoChatRuntimeFixture(t, "not-a-number")
	module.wait = func(context.Context, time.Duration) error {
		t.Fatal("special reply should not wait")
		return nil
	}

	if err := module.MessageCreateHandler()(context.Background(), autoChatMessageEvent("說出我是誰")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.TypingChannels) != 0 {
		t.Fatalf("typing = %#v", sideEffects.TypingChannels)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != `"你是誰"` {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestAutoChatFallbackMessageCreateIgnoresIneligibleEvents(t *testing.T) {
	module, sideEffects := autoChatRuntimeFixture(t, "-1")
	for _, event := range []events.Event{
		{Type: events.TypeMessageDelete, GuildID: "guild-1", ChannelID: "channel-1", MessageID: "message-1"},
		{Type: events.TypeMessageCreate, GuildID: "guild-1", ChannelID: "channel-1", MessageID: "message-1", IsBot: true},
		{Type: events.TypeMessageCreate, GuildID: "guild-1", ChannelID: "channel-1", MessageID: "message-1", Member: &events.Member{IsBot: true}},
		{Type: events.TypeMessageCreate, ChannelID: "channel-1", MessageID: "message-1"},
		{Type: events.TypeMessageCreate, GuildID: "guild-1", ChannelID: "channel-2", MessageID: "message-1"},
	} {
		if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
			t.Fatalf("event %#v: %v", event, err)
		}
	}
	if len(sideEffects.Sent) != 0 || len(sideEffects.TypingChannels) != 0 {
		t.Fatalf("side effects sent=%#v typing=%#v", sideEffects.Sent, sideEffects.TypingChannels)
	}
}

func TestAutoChatFallbackMessageCreatePropagatesSendAndWaitErrors(t *testing.T) {
	module, sideEffects := autoChatRuntimeFixture(t, "-1")
	waitErr := errors.New("wait canceled")
	module.wait = func(context.Context, time.Duration) error { return waitErr }
	if err := module.MessageCreateHandler()(context.Background(), autoChatMessageEvent("你好")); !errors.Is(err, waitErr) {
		t.Fatalf("wait error = %v", err)
	}

	module.wait = func(context.Context, time.Duration) error { return nil }
	sideEffects.Err = errors.New("discord unavailable")
	if err := module.MessageCreateHandler()(context.Background(), autoChatMessageEvent("你好")); !errors.Is(err, sideEffects.Err) {
		t.Fatalf("send error = %v", err)
	}
}

func TestAutoChatFallbackRegistersMessageCreateRoute(t *testing.T) {
	module, _ := autoChatRuntimeFixture(t, "-1")
	dispatcher := events.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("message-create route was not registered")
	}

	empty := events.NewDispatcher(nil)
	RuntimeModule{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("empty runtime module registered a route")
	}
}

func autoChatRuntimeFixture(t *testing.T, amount string) (RuntimeModule, *fakediscord.SideEffects) {
	t.Helper()
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: amount}
	sideEffects := fakediscord.NewSideEffects()
	module, err := NewRuntimeModule(configs, balances, sideEffects, sideEffects)
	if err != nil {
		t.Fatalf("new runtime module: %v", err)
	}
	return module, sideEffects
}

func autoChatMessageEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		MessageID: "message-1",
		UserID:    "user-1",
		Content:   content,
	}
}
