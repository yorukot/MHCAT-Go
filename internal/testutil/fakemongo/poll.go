package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type PollRepository struct {
	mu    sync.Mutex
	polls map[string]domain.Poll
	Err   error
}

func NewPollRepository() *PollRepository {
	return &PollRepository{polls: map[string]domain.Poll{}}
}

func (r *PollRepository) CreatePoll(ctx context.Context, create domain.PollCreate) (domain.Poll, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Poll{}, err
	}
	if err := create.Validate(); err != nil {
		return domain.Poll{}, err
	}
	poll := domain.NewPoll(create)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.polls[pollKey(create.GuildID, create.MessageID)] = clonePoll(poll)
	return poll, nil
}

func (r *PollRepository) GetPoll(ctx context.Context, guildID string, messageID string) (domain.Poll, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Poll{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	poll, ok := r.polls[pollKey(guildID, messageID)]
	if !ok {
		return domain.Poll{}, ports.ErrPollNotFound
	}
	return clonePoll(poll), nil
}

func (r *PollRepository) Vote(ctx context.Context, guildID string, messageID string, userID string, choice string, voteTime string) (domain.PollVoteChange, error) {
	if err := r.ready(ctx); err != nil {
		return domain.PollVoteChange{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pollKey(guildID, messageID)
	poll, ok := r.polls[key]
	if !ok {
		return domain.PollVoteChange{}, ports.ErrPollNotFound
	}
	if poll.Ended {
		return domain.PollVoteChange{}, ports.ErrPollEnded
	}
	if !pollHasChoice(poll, choice) {
		return domain.PollVoteChange{}, ports.ErrPollChoiceNotFound
	}
	for index, vote := range poll.Votes {
		if vote.UserID == userID && vote.Choice == choice {
			if !poll.CanChangeChoice {
				return domain.PollVoteChange{}, ports.ErrPollChangeNotAllowed
			}
			poll.Votes = append(poll.Votes[:index], poll.Votes[index+1:]...)
			r.polls[key] = clonePoll(poll)
			return domain.PollVoteChange{Removed: true, Poll: clonePoll(poll)}, nil
		}
	}
	if len(poll.UserChoices(userID)) >= poll.MaxChoices {
		return domain.PollVoteChange{}, ports.ErrPollChoiceLimit
	}
	poll.Votes = append(poll.Votes, domain.PollVote{UserID: userID, Choice: choice, Time: voteTime})
	r.polls[key] = clonePoll(poll)
	return domain.PollVoteChange{Added: true, Poll: clonePoll(poll)}, nil
}

func (r *PollRepository) TogglePoll(ctx context.Context, guildID string, messageID string, toggle domain.PollToggle) (domain.Poll, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Poll{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pollKey(guildID, messageID)
	poll, ok := r.polls[key]
	if !ok {
		return domain.Poll{}, ports.ErrPollNotFound
	}
	switch toggle {
	case domain.PollTogglePublicResult:
		poll.CanSeeResult = !poll.CanSeeResult
	case domain.PollToggleChangeChoice:
		poll.CanChangeChoice = !poll.CanChangeChoice
	case domain.PollToggleAnonymous:
		if poll.Anonymous {
			return domain.Poll{}, ports.ErrPollAnonymousLocked
		}
		poll.Anonymous = true
	case domain.PollToggleEnd:
		poll.Ended = !poll.Ended
	default:
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	r.polls[key] = clonePoll(poll)
	return clonePoll(poll), nil
}

func (r *PollRepository) SetMaxChoices(ctx context.Context, guildID string, messageID string, maxChoices int) (domain.Poll, error) {
	if err := r.ready(ctx); err != nil {
		return domain.Poll{}, err
	}
	if maxChoices < 1 {
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pollKey(guildID, messageID)
	poll, ok := r.polls[key]
	if !ok {
		return domain.Poll{}, ports.ErrPollNotFound
	}
	poll.MaxChoices = maxChoices
	r.polls[key] = clonePoll(poll)
	return clonePoll(poll), nil
}

func (r *PollRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func pollKey(guildID string, messageID string) string {
	return guildID + ":" + messageID
}

func pollHasChoice(poll domain.Poll, choice string) bool {
	for _, pollChoice := range poll.Choices {
		if pollChoice == choice {
			return true
		}
	}
	return false
}

func clonePoll(poll domain.Poll) domain.Poll {
	poll.Choices = append([]string(nil), poll.Choices...)
	poll.Votes = append([]domain.PollVote(nil), poll.Votes...)
	return poll
}
