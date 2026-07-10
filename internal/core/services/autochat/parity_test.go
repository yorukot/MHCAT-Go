package autochat

import (
	"crypto/sha256"
	"fmt"
	"math"
	"testing"
)

func TestLegacyAutoChatCorpusContract(t *testing.T) {
	const wantDigest = "9a04b0212a7b32f1cfbe1cb5579a0ac35ec3e1ec8f72843291ede61cc9ee667a"
	if got := fmt.Sprintf("%x", sha256.Sum256(legacyChatJSON)); got != wantDigest {
		t.Fatalf("legacy corpus sha256 = %s, want %s", got, wantDigest)
	}
	responses, err := decodeLegacyResponses(legacyChatJSON)
	if err != nil {
		t.Fatalf("decode legacy corpus: %v", err)
	}
	if len(responses) != 380 {
		t.Fatalf("legacy response count = %d, want 380", len(responses))
	}
	wantFirstKeys := []string{"0", "881", "v"}
	for index, want := range wantFirstKeys {
		if responses[index].key != want {
			t.Fatalf("legacy response key %d = %q, want %q", index, responses[index].key, want)
		}
	}
}

func TestLegacyAutoChatSayBranchContract(t *testing.T) {
	for _, test := range []struct {
		content string
		want    string
	}{
		{content: "說出說出", want: "說出甚麼?"},
		{content: "不是說出嗎", want: "不是嗎"},
		{content: "說出幹", want: "很抱歉，讀取到你說出了一些不好的字元，因此拒絕說出w\n字元:"},
		{content: "說出操", want: "操"},
		{content: "說出bitch", want: "bitch"},
		{content: "請說出我我", want: `"請你我"`},
	} {
		reply := (FallbackService{}).localReply(test.content)
		if reply.Content != test.want || reply.UseTypingDelay {
			t.Fatalf("localReply(%q) = %#v, want %q without delay", test.content, reply, test.want)
		}
	}
}

func TestLegacyAutoChatPriceContract(t *testing.T) {
	if got, want := LegacyAutoChatMessageCost("\x00ÿĀ🙂"), 8*LegacyAutoChatPricePerUnit; math.Abs(got-want) > 1e-15 {
		t.Fatalf("legacy message cost = %.12f, want %.12f", got, want)
	}
	for _, test := range []struct {
		value string
		want  bool
	}{
		{value: "0.00001", want: true},
		{value: "0"},
		{value: "-1"},
		{value: "NaN"},
		{value: "+Inf"},
		{value: "undefined"},
	} {
		if got := legacyPositiveAutoChatBalance(test.value); got != test.want {
			t.Fatalf("legacyPositiveAutoChatBalance(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}
