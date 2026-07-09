package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrPollNotFound          = errors.New("poll not found")
	ErrPollEnded             = errors.New("poll ended")
	ErrPollChoiceNotFound    = errors.New("poll choice not found")
	ErrPollChoiceLimit       = errors.New("poll choice limit reached")
	ErrPollChangeNotAllowed  = errors.New("poll choice change is not allowed")
	ErrPollAnonymousLocked   = errors.New("anonymous poll cannot return to named")
	ErrPollOwnerOnly         = errors.New("poll owner only")
	ErrPollResultUnavailable = errors.New("poll result unavailable")
)

type PollRepository interface {
	CreatePoll(ctx context.Context, create domain.PollCreate) (domain.Poll, error)
	GetPoll(ctx context.Context, guildID string, messageID string) (domain.Poll, error)
	Vote(ctx context.Context, guildID string, messageID string, userID string, choice string, voteTime string) (domain.PollVoteChange, error)
	TogglePoll(ctx context.Context, guildID string, messageID string, toggle domain.PollToggle) (domain.Poll, error)
	SetMaxChoices(ctx context.Context, guildID string, messageID string, maxChoices int) (domain.Poll, error)
}

type DiscordGuildMemberCounter interface {
	CountNonBotMembers(ctx context.Context, guildID string) (int, error)
}

type DiscordGuildMemberReader interface {
	DiscordGuildMemberCounter
	MemberTag(ctx context.Context, guildID string, userID string) (string, bool, error)
	MemberTags(ctx context.Context, guildID string, userIDs []string) (map[string]string, error)
}
