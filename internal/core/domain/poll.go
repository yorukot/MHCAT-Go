package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

var ErrInvalidPoll = errors.New("invalid poll")

type Poll struct {
	GuildID         string
	MessageID       string
	Question        string
	CreatorID       string
	MaxChoices      int
	CanChangeChoice bool
	CanSeeResult    bool
	Ended           bool
	Anonymous       bool
	Choices         []string
	Votes           []PollVote
}

type PollVote struct {
	UserID string
	Choice string
	Time   string
}

type PollCreate struct {
	GuildID   string
	MessageID string
	Question  string
	CreatorID string
	Choices   []string
}

type PollVoteChange struct {
	Added   bool
	Removed bool
	Poll    Poll
}

type PollToggle string

const (
	PollTogglePublicResult PollToggle = "poll_public_result"
	PollToggleChangeChoice PollToggle = "poll_can_change_choose"
	PollToggleAnonymous    PollToggle = "poll_anonymous"
	PollToggleEnd          PollToggle = "poll_end_poll"
)

func (p PollCreate) Validate() error {
	if strings.TrimSpace(p.GuildID) == "" ||
		strings.TrimSpace(p.MessageID) == "" ||
		strings.TrimSpace(p.Question) == "" ||
		strings.TrimSpace(p.CreatorID) == "" {
		return ErrInvalidPoll
	}
	if len(p.Choices) < 2 || len(p.Choices) > 19 {
		return ErrInvalidPoll
	}
	seen := map[string]struct{}{}
	for _, choice := range p.Choices {
		if strings.TrimSpace(choice) == "" || len(utf16.Encode([]rune(choice))) > 80 {
			return ErrInvalidPoll
		}
		if _, ok := seen[choice]; ok {
			return ErrInvalidPoll
		}
		seen[choice] = struct{}{}
	}
	return nil
}

func (p Poll) UniqueVoterCount() int {
	seen := map[string]struct{}{}
	for _, vote := range p.Votes {
		if vote.UserID != "" {
			seen[vote.UserID] = struct{}{}
		}
	}
	return len(seen)
}

func (p Poll) CountChoice(choice string) int {
	count := 0
	for _, vote := range p.Votes {
		if vote.Choice == choice {
			count++
		}
	}
	return count
}

func (p Poll) UserChoices(userID string) []string {
	var choices []string
	for _, vote := range p.Votes {
		if vote.UserID == userID {
			choices = append(choices, vote.Choice)
		}
	}
	return choices
}

func NewPoll(create PollCreate) Poll {
	return Poll{
		GuildID:         create.GuildID,
		MessageID:       create.MessageID,
		Question:        create.Question,
		CreatorID:       create.CreatorID,
		MaxChoices:      1,
		CanChangeChoice: false,
		CanSeeResult:    false,
		Ended:           false,
		Anonymous:       false,
		Choices:         append([]string(nil), create.Choices...),
		Votes:           []PollVote{},
	}
}

func LegacyVoteTime(now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	return strconv.FormatInt(now.UnixMilli(), 10)
}
