package autochat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

type paidRuntimeClock struct{ now time.Time }

func (c paidRuntimeClock) Now() time.Time { return c.now }

type paidRuntimeFailingTyping struct{ err error }

func (t paidRuntimeFailingTyping) SendTyping(context.Context, string) error { return t.err }

func TestPaidAutoChatQueuesWaitsAndReplies(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
	now := time.UnixMilli(1_700_000_000_123)
	module.clock = paidRuntimeClock{now: now}
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: now.UnixMilli(), Cost: 0.00015}
	handoff.Response = domain.AutoChatPaidResponse{GuildID: "guild-1", Content: "worker answer", RequestTimeMilli: now.UnixMilli(), Reply: true}
	var waited time.Duration
	module.wait = func(_ context.Context, delay time.Duration) error {
		waited = delay
		return nil
	}

	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, events.ErrStopPropagation) {
		t.Fatalf("message create: %v", err)
	}
	if waited != LegacyAutoChatPaidResponseDelay || len(handoff.Requests) != 1 || handoff.ResponseTime != now.UnixMilli() {
		t.Fatalf("waited=%s requests=%#v response_time=%d", waited, handoff.Requests, handoff.ResponseTime)
	}
	if len(sideEffects.TypingChannels) != 1 || sideEffects.TypingChannels[0] != "channel-1" {
		t.Fatalf("typing = %#v", sideEffects.TypingChannels)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	message := sideEffects.Sent[0].Message
	if message.Content != "worker answer" || message.ReplyToMessageID != "message-1" {
		t.Fatalf("message = %#v", message)
	}
	assertPaidAutoChatMentionsSuppressed(t, message)
}

func TestPaidAutoChatAlwaysRepliesAndIgnoresTypingFailure(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
	now := module.clock.Now().UnixMilli()
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: now}
	handoff.Response = domain.AutoChatPaidResponse{
		GuildID:          "guild-1",
		Content:          "worker answer",
		RequestTimeMilli: now,
		Reply:            false,
	}
	module.typing = paidRuntimeFailingTyping{err: errors.New("typing unavailable")}
	module.wait = func(context.Context, time.Duration) error { return nil }

	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, events.ErrStopPropagation) {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	message := sideEffects.Sent[0].Message
	if message.Content != "worker answer" || message.ReplyToMessageID != "message-1" {
		t.Fatalf("message = %#v", message)
	}
	assertPaidAutoChatMentionsSuppressed(t, message)
}

func TestPaidAutoChatRejectsUnsafeInputAndDeletesTransientMessages(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
	var waited time.Duration
	module.wait = func(_ context.Context, delay time.Duration) error {
		waited = delay
		return nil
	}

	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello @everyone")); !errors.Is(err, events.ErrStopPropagation) {
		t.Fatalf("message create: %v", err)
	}
	if len(handoff.Requests) != 0 || waited != legacyAutoChatUnsafeInputDelay {
		t.Fatalf("requests=%#v waited=%s", handoff.Requests, waited)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != legacyAutoChatUnsafeInputMessage {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	if len(sideEffects.DeletedMessage) != 2 || sideEffects.DeletedMessage[0].MessageID != "message-1" || sideEffects.DeletedMessage[1] != sideEffects.Sent[0].Ref {
		t.Fatalf("deleted = %#v", sideEffects.DeletedMessage)
	}
	assertPaidAutoChatMentionsSuppressed(t, sideEffects.Sent[0].Message)
}

func TestPaidAutoChatBusyUsesLegacyTwoSecondWarning(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
	handoff.QueueErr = ports.ErrAutoChatPaidBusy
	var waited time.Duration
	module.wait = func(_ context.Context, delay time.Duration) error {
		waited = delay
		return nil
	}

	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, events.ErrStopPropagation) {
		t.Fatalf("message create: %v", err)
	}
	if waited != legacyAutoChatBusyWarningDelay || len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != legacyAutoChatBusyMessage {
		t.Fatalf("waited=%s sent=%#v", waited, sideEffects.Sent)
	}
	if len(sideEffects.DeletedMessage) != 2 {
		t.Fatalf("deleted = %#v", sideEffects.DeletedMessage)
	}
}

func TestPaidAutoChatBlocksWorkerMentions(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
	now := module.clock.Now().UnixMilli()
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: now}
	handoff.Response = domain.AutoChatPaidResponse{GuildID: "guild-1", Content: "hello @everyone", RequestTimeMilli: now, Reply: true}
	module.wait = func(context.Context, time.Duration) error { return nil }

	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, events.ErrStopPropagation) {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != legacyAutoChatUnsafeOutputMessage || sideEffects.Sent[0].Message.ReplyToMessageID != "message-1" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	assertPaidAutoChatMentionsSuppressed(t, sideEffects.Sent[0].Message)
}

func TestPaidAutoChatIgnoresIneligibleEventsAndRegistersRoute(t *testing.T) {
	module, handoff, sideEffects := paidRuntimeFixture(t)
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
	if len(handoff.Requests) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("requests=%#v sent=%#v", handoff.Requests, sideEffects.Sent)
	}
	dispatcher := events.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("expected paid autochat message route")
	}
}

func TestPaidAutoChatStopsLocalFallbackAfterBalanceDebit(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "0.00001"}
	handoff := &fakemongo.AutoChatPaidRepository{}
	sideEffects := fakediscord.NewSideEffects()
	clock := paidRuntimeClock{now: time.UnixMilli(1_700_000_000_123)}
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: clock.Now().UnixMilli()}
	handoff.Response = domain.AutoChatPaidResponse{GuildID: "guild-1", Content: "worker answer", RequestTimeMilli: clock.Now().UnixMilli(), Reply: true}
	handoff.QueueHook = func(domain.AutoChatPaidRequest) {
		balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "-1"}
	}
	paid, err := NewPaidRuntimeModule(configs, balances, handoff, sideEffects, sideEffects, clock)
	if err != nil {
		t.Fatalf("new paid runtime: %v", err)
	}
	paid.wait = func(context.Context, time.Duration) error { return nil }
	fallback, err := NewRuntimeModule(configs, balances, sideEffects, sideEffects)
	if err != nil {
		t.Fatalf("new fallback runtime: %v", err)
	}
	fallback.wait = func(context.Context, time.Duration) error { return nil }
	dispatcher := events.NewDispatcher(nil)
	paid.RegisterEventRoutes(dispatcher)
	fallback.RegisterEventRoutes(dispatcher)

	if err := dispatcher.Dispatch(context.Background(), paidAutoChatEvent("hello")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != "worker answer" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestPaidAutoChatPropagatesWaitAndResponseErrors(t *testing.T) {
	module, handoff, _ := paidRuntimeFixture(t)
	now := module.clock.Now().UnixMilli()
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: now}
	waitErr := errors.New("wait canceled")
	module.wait = func(context.Context, time.Duration) error { return waitErr }
	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, waitErr) {
		t.Fatalf("wait error = %v", err)
	}

	module.wait = func(context.Context, time.Duration) error { return nil }
	handoff.ResponseErr = ports.ErrAutoChatPaidResponseMissing
	if err := module.MessageCreateHandler()(context.Background(), paidAutoChatEvent("hello")); !errors.Is(err, ports.ErrAutoChatPaidResponseMissing) {
		t.Fatalf("response error = %v", err)
	}
}

func paidRuntimeFixture(t *testing.T) (PaidRuntimeModule, *fakemongo.AutoChatPaidRepository, *fakediscord.SideEffects) {
	t.Helper()
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "10"}
	handoff := &fakemongo.AutoChatPaidRepository{}
	sideEffects := fakediscord.NewSideEffects()
	clock := paidRuntimeClock{now: time.UnixMilli(1_700_000_000_123)}
	module, err := NewPaidRuntimeModule(configs, balances, handoff, sideEffects, sideEffects, clock)
	if err != nil {
		t.Fatalf("new paid runtime: %v", err)
	}
	return module, handoff, sideEffects
}

func paidAutoChatEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		MessageID: "message-1",
		UserID:    "user-1",
		Content:   content,
	}
}

func assertPaidAutoChatMentionsSuppressed(t *testing.T, message ports.OutboundMessage) {
	t.Helper()
	mentions := message.AllowedMentions
	if mentions.ParseUsers || mentions.ParseRoles || mentions.ParseEveryone || mentions.RepliedUser || len(mentions.UserIDs) != 0 || len(mentions.RoleIDs) != 0 {
		t.Fatalf("allowed mentions = %#v", mentions)
	}
}
