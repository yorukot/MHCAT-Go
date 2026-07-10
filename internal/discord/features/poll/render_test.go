package poll

import (
	"fmt"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestPollChoiceVotersUsesLegacyGlobalSuppression(t *testing.T) {
	poll := domain.NewPoll(domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "question",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B"},
	})
	poll.Votes = []domain.PollVote{{UserID: "user-1", Choice: "A"}}

	poll.Anonymous = true
	if got := pollChoiceVoters(poll, "B"); got != "該投票為匿名，無法查看誰有進行投票" {
		t.Fatalf("anonymous empty choice = %q", got)
	}

	poll.Anonymous = false
	poll.Votes = make([]domain.PollVote, 51)
	for index := range poll.Votes {
		poll.Votes[index] = domain.PollVote{UserID: fmt.Sprintf("user-%d", index), Choice: "A"}
	}
	if got := pollChoiceVoters(poll, "B"); got != "由於人數過多，無法顯示所有人" {
		t.Fatalf("large empty choice = %q", got)
	}

	poll.Votes = poll.Votes[:1]
	if got := pollChoiceVoters(poll, "B"); got != "<a:Discord_AnimatedNo:1015989839809757295> | 還沒有人投給這個選項" {
		t.Fatalf("ordinary empty choice = %q", got)
	}
}

func TestLegacyPollPercentageMatchesJavaScriptToFixed(t *testing.T) {
	tests := []struct {
		name        string
		numerator   int
		denominator int
		want        string
	}{
		{name: "exact halfway rounds up", numerator: 1, denominator: 32, want: "3.13"},
		{name: "binary value below halfway stays down", numerator: 23, denominator: 160, want: "14.37"},
		{name: "repeating fraction", numerator: 1, denominator: 3, want: "33.33"},
		{name: "zero denominator", numerator: 0, denominator: 0, want: "NaN"},
		{name: "positive zero denominator", numerator: 1, denominator: 0, want: "Infinity"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := legacyPollPercentage(tc.numerator, tc.denominator); got != tc.want {
				t.Fatalf("percentage = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPollResultEmbedUsesLegacyPercentageRounding(t *testing.T) {
	poll := domain.NewPoll(domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "question",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B"},
	})
	poll.Votes = append(poll.Votes, domain.PollVote{UserID: "user-a", Choice: "A"})
	for index := 1; index < 32; index++ {
		poll.Votes = append(poll.Votes, domain.PollVote{UserID: fmt.Sprintf("user-%d", index), Choice: "B"})
	}

	message := pollResultEmbedMessage(poll, 0)
	if got := message.Embeds[0].Fields[0].Name; got != "A(共1人 `3.13`%)" {
		t.Fatalf("field name = %q", got)
	}
}
