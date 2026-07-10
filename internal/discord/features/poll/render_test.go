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
