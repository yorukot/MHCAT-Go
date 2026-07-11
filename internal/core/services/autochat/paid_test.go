package autochat

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestNewPaidHandoffServiceRequiresRepositories(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	balances := fakemongo.NewBalanceRepository()
	handoff := &fakemongo.AutoChatPaidRepository{}
	for _, test := range []struct {
		configs  ports.AutoChatConfigReader
		balances ports.BalanceRepository
		handoff  ports.AutoChatPaidRepository
	}{
		{balances: balances, handoff: handoff},
		{configs: configs, handoff: handoff},
		{configs: configs, balances: balances},
	} {
		if _, err := NewPaidHandoffService(test.configs, test.balances, test.handoff); err == nil {
			t.Fatal("expected missing repository error")
		}
	}
}

func TestPaidHandoffSubmitQueuesLegacyPricedRequest(t *testing.T) {
	service, handoff := paidHandoffFixture(t, "10")
	now := time.UnixMilli(1_700_000_000_123)
	handoff.QueueResult = domain.AutoChatPaidDispatch{RequestTimeMilli: now.UnixMilli(), Cost: 0.00015}

	result, err := service.Submit(context.Background(), " guild-1 ", " channel-1 ", "A你🙂", now)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if result.State != domain.AutoChatPaidQueued || result.Dispatch.RequestTimeMilli != now.UnixMilli() {
		t.Fatalf("result = %#v", result)
	}
	if len(handoff.Requests) != 1 {
		t.Fatalf("requests = %#v", handoff.Requests)
	}
	request := handoff.Requests[0]
	if request.GuildID != "guild-1" || request.Content != "A你🙂" || request.RequestedAtMilli != now.UnixMilli() {
		t.Fatalf("request = %#v", request)
	}
	if math.Abs(request.Cost-0.00021) > 1e-12 {
		t.Fatalf("cost = %.12f", request.Cost)
	}
}

func TestPaidHandoffSubmitPreservesEligibilityAndSafetyBranches(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_123)
	for _, test := range []struct {
		name       string
		amount     string
		channelID  string
		content    string
		configure  bool
		wantState  domain.AutoChatPaidSubmissionState
		wantQueued int
	}{
		{name: "missing config", amount: "10", channelID: "channel-1", content: "hello", wantState: domain.AutoChatPaidIgnored},
		{name: "wrong channel", amount: "10", channelID: "channel-2", content: "hello", configure: true, wantState: domain.AutoChatPaidIgnored},
		{name: "zero", amount: "0", channelID: "channel-1", content: "hello", configure: true, wantState: domain.AutoChatPaidIgnored},
		{name: "negative", amount: "-1", channelID: "channel-1", content: "hello", configure: true, wantState: domain.AutoChatPaidIgnored},
		{name: "malformed", amount: "nope", channelID: "channel-1", content: "hello", configure: true, wantState: domain.AutoChatPaidIgnored},
		{name: "infinite", amount: "+Inf", channelID: "channel-1", content: "hello", configure: true, wantState: domain.AutoChatPaidIgnored},
		{name: "mention", amount: "10", channelID: "channel-1", content: "hello @everyone", configure: true, wantState: domain.AutoChatPaidUnsafeInput},
	} {
		t.Run(test.name, func(t *testing.T) {
			configs := fakemongo.NewAutoChatConfigRepository()
			if test.configure {
				configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
			}
			balances := fakemongo.NewBalanceRepository()
			balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: test.amount}
			handoff := &fakemongo.AutoChatPaidRepository{}
			service, err := NewPaidHandoffService(configs, balances, handoff)
			if err != nil {
				t.Fatalf("new service: %v", err)
			}
			result, err := service.Submit(context.Background(), "guild-1", test.channelID, test.content, now)
			if err != nil {
				t.Fatalf("submit: %v", err)
			}
			if result.State != test.wantState || len(handoff.Requests) != test.wantQueued {
				t.Fatalf("result=%#v requests=%#v", result, handoff.Requests)
			}
		})
	}
}

func TestPaidHandoffSubmitPreservesStoredChannelWhitespace(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: " channel-1 "}
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "5"}
	service, err := NewPaidHandoffService(configs, balances, &fakemongo.AutoChatPaidRepository{})
	if err != nil {
		t.Fatalf("new paid service: %v", err)
	}
	submission, err := service.Submit(context.Background(), "guild-1", "channel-1", "hello", time.UnixMilli(1_700_000_000_000))
	if err != nil || submission.State != domain.AutoChatPaidIgnored {
		t.Fatalf("submission=%#v err=%v", submission, err)
	}
}

func TestPaidHandoffSubmitChecksConfigBeforeBalance(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Err = errors.New("config unavailable")
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "0"}
	service, err := NewPaidHandoffService(configs, balances, &fakemongo.AutoChatPaidRepository{})
	if err != nil {
		t.Fatalf("new paid service: %v", err)
	}
	if _, err := service.Submit(context.Background(), "guild-1", "channel-1", "hello", time.UnixMilli(1_700_000_000_000)); err == nil {
		t.Fatal("config lookup must fail before an irrelevant balance lookup")
	}

	configs.Err = nil
	submission, err := service.Submit(context.Background(), "guild-1", "channel-1", "hello", time.UnixMilli(1_700_000_000_000))
	if err != nil || submission.State != domain.AutoChatPaidIgnored {
		t.Fatalf("missing config submission=%#v err=%v", submission, err)
	}
}

func TestPaidHandoffSubmitMapsQueueOutcomes(t *testing.T) {
	for _, test := range []struct {
		name      string
		queueErr  error
		wantState domain.AutoChatPaidSubmissionState
		wantErr   error
	}{
		{name: "balance raced", queueErr: ports.ErrAutoChatPaidBalanceUnavailable, wantState: domain.AutoChatPaidIgnored},
		{name: "busy", queueErr: ports.ErrAutoChatPaidBusy, wantState: domain.AutoChatPaidBusy},
		{name: "failure", queueErr: ports.ErrAutoChatPaidStateConflict, wantErr: ports.ErrAutoChatPaidStateConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, handoff := paidHandoffFixture(t, "10")
			handoff.QueueErr = test.queueErr
			result, err := service.Submit(context.Background(), "guild-1", "channel-1", "hello", time.UnixMilli(1_700_000_000_123))
			if !errors.Is(err, test.wantErr) || result.State != test.wantState {
				t.Fatalf("result=%#v err=%v", result, err)
			}
		})
	}
}

func TestPaidHandoffResponseDelegatesRequestIdentity(t *testing.T) {
	service, handoff := paidHandoffFixture(t, "10")
	handoff.Response = domain.AutoChatPaidResponse{GuildID: "guild-1", Content: "answer", RequestTimeMilli: 123, Reply: true}
	response, err := service.Response(context.Background(), " guild-1 ", 123)
	if err != nil {
		t.Fatalf("response: %v", err)
	}
	if response.Content != "answer" || handoff.ResponseGuild != "guild-1" || handoff.ResponseTime != 123 {
		t.Fatalf("response=%#v guild=%q time=%d", response, handoff.ResponseGuild, handoff.ResponseTime)
	}
}

func paidHandoffFixture(t *testing.T, amount string) (PaidHandoffService, *fakemongo.AutoChatPaidRepository) {
	t.Helper()
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: amount}
	handoff := &fakemongo.AutoChatPaidRepository{}
	service, err := NewPaidHandoffService(configs, balances, handoff)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service, handoff
}
