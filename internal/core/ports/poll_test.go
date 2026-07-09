package ports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestPollRepositoryContractWithFake(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	var port ports.PollRepository = repo
	poll, err := port.CreatePoll(context.Background(), domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "問題",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B", "C"},
	})
	if err != nil {
		t.Fatalf("create poll: %v", err)
	}
	if poll.MaxChoices != 1 || len(poll.Choices) != 3 {
		t.Fatalf("poll = %#v", poll)
	}

	change, err := port.Vote(context.Background(), "guild-1", "message-1", "user-1", "A", "1")
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if !change.Added || len(change.Poll.Votes) != 1 {
		t.Fatalf("change = %#v", change)
	}

	if _, err := port.Vote(context.Background(), "guild-1", "message-1", "user-1", "B", "2"); !errors.Is(err, ports.ErrPollChoiceLimit) {
		t.Fatalf("expected choice limit, got %v", err)
	}

	updated, err := port.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollToggleChangeChoice)
	if err != nil {
		t.Fatalf("toggle change: %v", err)
	}
	if !updated.CanChangeChoice {
		t.Fatalf("updated = %#v", updated)
	}
	change, err = port.Vote(context.Background(), "guild-1", "message-1", "user-1", "A", "3")
	if err != nil {
		t.Fatalf("remove vote: %v", err)
	}
	if !change.Removed || len(change.Poll.Votes) != 0 {
		t.Fatalf("remove change = %#v", change)
	}
}

func TestPollRepositoryContractContextCancellation(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.GetPoll(ctx, "guild-1", "message-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestPollRepositoryContractAnonymousOneWay(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	_, err := repo.CreatePoll(context.Background(), domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "問題",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B"},
	})
	if err != nil {
		t.Fatalf("create poll: %v", err)
	}
	if _, err := repo.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollToggleAnonymous); err != nil {
		t.Fatalf("first anonymous toggle: %v", err)
	}
	if _, err := repo.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollToggleAnonymous); !errors.Is(err, ports.ErrPollAnonymousLocked) {
		t.Fatalf("expected anonymous locked, got %v", err)
	}
}
