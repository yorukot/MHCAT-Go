package autochat

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestFallbackServiceUsesLegacyCorpusWithoutBalanceRow(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	service, err := NewFallbackService(configs, fakemongo.NewBalanceRepository())
	if err != nil {
		t.Fatalf("new fallback service: %v", err)
	}
	if len(service.responses) != 380 {
		t.Fatalf("legacy response count = %d", len(service.responses))
	}

	reply, err := service.Reply(context.Background(), "guild-1", "channel-1", "你好")
	if err != nil {
		t.Fatalf("reply: %v", err)
	}
	if reply.Content != "你好，有甚麼我能幫忙的嗎?" || !reply.UseTypingDelay {
		t.Fatalf("reply = %#v", reply)
	}
}

func TestFallbackServicePreservesLegacySayReplies(t *testing.T) {
	service := FallbackService{}
	for _, test := range []struct {
		content string
		want    string
	}{
		{content: "說出", want: "說出甚麼?"},
		{content: "說出hello", want: "hello"},
		{content: "說出幹", want: "很抱歉，讀取到你說出了一些不好的字元，因此拒絕說出w\n字元:"},
	} {
		reply := service.localReply(test.content)
		if reply.Content != test.want || reply.UseTypingDelay {
			t.Fatalf("local reply for %q = %#v", test.content, reply)
		}
	}
}

func TestFallbackServiceUsesLegacyUnknownReplyWhenNothingIsSimilar(t *testing.T) {
	reply := (FallbackService{}).localReply("anything")
	if reply.Content != "我看不懂你的意思，在講一次好不好w" || reply.UseTypingDelay {
		t.Fatalf("reply = %#v", reply)
	}
}

func TestFallbackServicePreservesBalanceAndChannelGates(t *testing.T) {
	for _, test := range []struct {
		name      string
		channelID string
		amount    string
	}{
		{name: "wrong channel", channelID: "channel-2", amount: "-1"},
		{name: "zero balance", channelID: "channel-1", amount: "0"},
		{name: "null balance", channelID: "channel-1", amount: "null"},
		{name: "positive infinity", channelID: "channel-1", amount: "Infinity"},
		{name: "positive balance", channelID: "channel-1", amount: "2.5"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configs := fakemongo.NewAutoChatConfigRepository()
			configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
			balances := fakemongo.NewBalanceRepository()
			balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: test.amount}
			service, err := NewFallbackService(configs, balances)
			if err != nil {
				t.Fatalf("new fallback service: %v", err)
			}
			reply, err := service.Reply(context.Background(), "guild-1", test.channelID, "你好")
			if err != nil {
				t.Fatalf("reply: %v", err)
			}
			if reply.Content != "" {
				t.Fatalf("reply = %#v", reply)
			}
		})
	}
}

func TestFallbackServicePreservesStoredChannelWhitespace(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: " channel-1 "}
	service, err := NewFallbackService(configs, fakemongo.NewBalanceRepository())
	if err != nil {
		t.Fatalf("new fallback service: %v", err)
	}
	reply, err := service.Reply(context.Background(), "guild-1", "channel-1", "你好")
	if err != nil || reply.Content != "" {
		t.Fatalf("reply=%#v err=%v", reply, err)
	}
}

func TestFallbackServiceUsesLocalRepliesForNegativeOrMalformedBalance(t *testing.T) {
	for _, amount := range []string{"-0.01", "not-a-number", "NaN", "undefined"} {
		t.Run(amount, func(t *testing.T) {
			configs := fakemongo.NewAutoChatConfigRepository()
			configs.Configs["guild-1"] = domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
			balances := fakemongo.NewBalanceRepository()
			balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: amount}
			service, err := NewFallbackService(configs, balances)
			if err != nil {
				t.Fatalf("new fallback service: %v", err)
			}
			reply, err := service.Reply(context.Background(), "guild-1", "channel-1", "說出我是誰")
			if err != nil {
				t.Fatalf("reply: %v", err)
			}
			if reply.Content != `"你是誰"` || reply.UseTypingDelay {
				t.Fatalf("reply = %#v", reply)
			}
		})
	}
}

func TestFallbackServiceIgnoresMissingConfigAndPropagatesRepositoryErrors(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	service, err := NewFallbackService(configs, fakemongo.NewBalanceRepository())
	if err != nil {
		t.Fatalf("new fallback service: %v", err)
	}
	reply, err := service.Reply(context.Background(), "guild-1", "channel-1", "你好")
	if err != nil || reply.Content != "" {
		t.Fatalf("missing config reply=%#v err=%v", reply, err)
	}

	configs.Err = errors.New("database unavailable")
	_, err = service.Reply(context.Background(), "guild-1", "channel-1", "你好")
	if err == nil || errors.Is(err, ports.ErrAutoChatConfigMissing) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestFallbackServiceChecksNonnegativeBalanceBeforeConfig(t *testing.T) {
	configs := fakemongo.NewAutoChatConfigRepository()
	configs.Err = errors.New("config unavailable")
	balances := fakemongo.NewBalanceRepository()
	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "1"}
	service, err := NewFallbackService(configs, balances)
	if err != nil {
		t.Fatalf("new fallback service: %v", err)
	}
	reply, err := service.Reply(context.Background(), "guild-1", "channel-1", "你好")
	if err != nil || reply.Content != "" {
		t.Fatalf("nonnegative balance reply=%#v err=%v", reply, err)
	}

	balances.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "-1"}
	if _, err := service.Reply(context.Background(), "guild-1", "channel-1", "你好"); err == nil {
		t.Fatal("negative balance should continue to the failing config lookup")
	}
}

func TestDecodeLegacyResponsesUsesJavaScriptObjectKeyOrder(t *testing.T) {
	responses, err := decodeLegacyResponses([]byte(`{"v":"version","20":"twenty","3":"three","01":"leading","text":"value"}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{"3", "20", "v", "01", "text"}
	if len(responses) != len(want) {
		t.Fatalf("responses = %#v", responses)
	}
	for i, key := range want {
		if responses[i].key != key {
			t.Fatalf("response[%d] key = %q, want %q", i, responses[i].key, key)
		}
	}
}

func TestLegacySimilarityUsesUTF16CodeUnits(t *testing.T) {
	if got := legacySimilarity("A😀", "a😀"); got != 1 {
		t.Fatalf("similarity = %v", got)
	}
	if got := legacySimilarity("😀", ""); got != 0 {
		t.Fatalf("empty similarity = %v", got)
	}
}

func TestLegacySimilarityUsesJavaScriptUnicodeLowercase(t *testing.T) {
	for _, test := range []struct {
		name  string
		left  string
		right string
		want  float64
	}{
		{name: "expanding lowercase", left: "İ", right: "i", want: 0},
		{name: "contextual final sigma", left: "ΟΣ", right: "οσ", want: 0.5},
		{name: "original utf16 denominator", left: "İ", right: "i̇", want: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := legacySimilarity(test.left, test.right); got != test.want {
				t.Fatalf("legacySimilarity(%q, %q) = %v, want %v", test.left, test.right, got, test.want)
			}
		})
	}
}
