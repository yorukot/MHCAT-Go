package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPollRepositoryConstructorsRejectNilDependencies(t *testing.T) {
	if _, err := NewPollRepository(nil); err == nil {
		t.Fatal("expected nil poll collection error")
	}
	if _, err := NewPollRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestPollMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewPollRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new poll repository: %v", err)
	}
	ctx := context.Background()
	created, err := repository.CreatePoll(ctx, domain.PollCreate{
		GuildID: " guild-1 ", MessageID: " message-1 ", Question: "question", CreatorID: " owner-1 ", Choices: []string{"A", "B", "C"},
	})
	if err != nil {
		t.Fatalf("create poll: %v", err)
	}
	if created.GuildID != "guild-1" || created.MessageID != "message-1" || created.CreatorID != "owner-1" || created.MaxChoices != 1 {
		t.Fatalf("created poll = %#v", created)
	}
	loaded, err := repository.GetPoll(ctx, " guild-1 ", " message-1 ")
	if err != nil || loaded.GuildID != "guild-1" || len(loaded.Choices) != 3 {
		t.Fatalf("loaded poll = %#v err=%v", loaded, err)
	}

	change, err := repository.Vote(ctx, "guild-1", "message-1", " user-1 ", "A", "100")
	if err != nil || !change.Added || change.Removed || len(change.Poll.Votes) != 1 {
		t.Fatalf("added vote = %#v err=%v", change, err)
	}
	if _, err := repository.Vote(ctx, "guild-1", "message-1", "user-1", "A", "101"); !errors.Is(err, ports.ErrPollChangeNotAllowed) {
		t.Fatalf("unchangeable vote error = %v", err)
	}
	if _, err := repository.Vote(ctx, "guild-1", "message-1", "user-1", "B", "102"); !errors.Is(err, ports.ErrPollChoiceLimit) {
		t.Fatalf("choice limit error = %v", err)
	}
	if _, err := repository.Vote(ctx, "guild-1", "message-1", "user-2", "missing", "103"); !errors.Is(err, ports.ErrPollChoiceNotFound) {
		t.Fatalf("missing choice error = %v", err)
	}

	loaded, err = repository.SetMaxChoices(ctx, "guild-1", "message-1", 2)
	if err != nil || loaded.MaxChoices != 2 {
		t.Fatalf("poll max choices = %#v err=%v", loaded, err)
	}
	change, err = repository.Vote(ctx, "guild-1", "message-1", "user-1", "B", "104")
	if err != nil || !change.Added || len(change.Poll.Votes) != 2 {
		t.Fatalf("second choice vote = %#v err=%v", change, err)
	}
	loaded, err = repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollToggleChangeChoice)
	if err != nil || !loaded.CanChangeChoice {
		t.Fatalf("change-choice toggle = %#v err=%v", loaded, err)
	}
	change, err = repository.Vote(ctx, "guild-1", "message-1", "user-1", "A", "105")
	if err != nil || !change.Removed || len(change.Poll.Votes) != 1 {
		t.Fatalf("removed vote = %#v err=%v", change, err)
	}
	loaded, err = repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollTogglePublicResult)
	if err != nil || !loaded.CanSeeResult {
		t.Fatalf("public-result toggle = %#v err=%v", loaded, err)
	}
	loaded, err = repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollToggleAnonymous)
	if err != nil || !loaded.Anonymous {
		t.Fatalf("anonymous toggle = %#v err=%v", loaded, err)
	}
	if _, err := repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollToggleAnonymous); !errors.Is(err, ports.ErrPollAnonymousLocked) {
		t.Fatalf("second anonymous toggle error = %v", err)
	}
	loaded, err = repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollToggleEnd)
	if err != nil || !loaded.Ended {
		t.Fatalf("end toggle = %#v err=%v", loaded, err)
	}
	if _, err := repository.Vote(ctx, "guild-1", "message-1", "user-2", "A", "106"); !errors.Is(err, ports.ErrPollEnded) {
		t.Fatalf("ended poll vote error = %v", err)
	}

	if _, err := repository.TogglePoll(ctx, "guild-1", "message-1", domain.PollToggle("invalid")); !errors.Is(err, domain.ErrInvalidPoll) {
		t.Fatalf("invalid toggle error = %v", err)
	}
	if _, err := repository.SetMaxChoices(ctx, "guild-1", "message-1", 0); !errors.Is(err, domain.ErrInvalidPoll) {
		t.Fatalf("invalid max choices error = %v", err)
	}
	if _, err := repository.GetPoll(ctx, "guild-1", "missing"); !errors.Is(err, ports.ErrPollNotFound) {
		t.Fatalf("missing poll error = %v", err)
	}
	if _, err := repository.SetMaxChoices(ctx, "guild-1", "missing", 2); !errors.Is(err, ports.ErrPollNotFound) {
		t.Fatalf("missing poll max choices error = %v", err)
	}
}

func TestPollMongoIntegrationHandlesMissingAndMalformedVoteArrays(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewPollRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new poll repository: %v", err)
	}
	ctx := context.Background()
	base := bson.D{
		{Key: "question", Value: "question"}, {Key: "create_member_id", Value: "owner"},
		{Key: "many_choose", Value: 1}, {Key: "can_change_choose", Value: false},
		{Key: "can_see_result", Value: false}, {Key: "end", Value: false},
		{Key: "anonymous", Value: false}, {Key: "choose_data", Value: bson.A{"A", "B"}},
	}
	missingArray := append(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "messageid", Value: "missing-array"}}, base...)
	if _, err := database.Collection(PollCollectionName).InsertOne(ctx, missingArray); err != nil {
		t.Fatalf("seed poll without vote array: %v", err)
	}
	change, err := repository.Vote(ctx, "guild-1", "missing-array", "user-1", "A", "100")
	if err != nil || !change.Added || len(change.Poll.Votes) != 1 {
		t.Fatalf("vote on missing array poll = %#v err=%v", change, err)
	}

	malformedArray := append(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "messageid", Value: "malformed-array"}}, base...)
	malformedArray = append(malformedArray, bson.E{Key: "join_member", Value: "broken"})
	if _, err := database.Collection(PollCollectionName).InsertOne(ctx, malformedArray); err != nil {
		t.Fatalf("seed malformed vote array: %v", err)
	}
	if _, err := repository.Vote(ctx, "guild-1", "malformed-array", "user-1", "A", "100"); !errors.Is(err, ports.ErrPollChoiceNotFound) {
		t.Fatalf("malformed vote array error = %v", err)
	}
}

func TestPollMongoIntegrationConcurrentVoteIsUnique(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewPollRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new poll repository: %v", err)
	}
	ctx := context.Background()
	if _, err := repository.CreatePoll(ctx, domain.PollCreate{
		GuildID: "guild-1", MessageID: "message-1", Question: "question", CreatorID: "owner", Choices: []string{"A", "B"},
	}); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	const voters = 10
	errorsByVote := make([]error, voters)
	var wait sync.WaitGroup
	wait.Add(voters)
	for index := range errorsByVote {
		go func() {
			defer wait.Done()
			_, errorsByVote[index] = repository.Vote(ctx, "guild-1", "message-1", "user-1", "A", "100")
		}()
	}
	wait.Wait()
	succeeded := 0
	rejected := 0
	for _, err := range errorsByVote {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ports.ErrPollChangeNotAllowed):
			rejected++
		default:
			t.Fatalf("concurrent vote error = %v", err)
		}
	}
	if succeeded != 1 || rejected != voters-1 {
		t.Fatalf("concurrent votes succeeded=%d rejected=%d errors=%v", succeeded, rejected, errorsByVote)
	}
	loaded, err := repository.GetPoll(ctx, "guild-1", "message-1")
	if err != nil || len(loaded.Votes) != 1 {
		t.Fatalf("poll after concurrent vote = %#v err=%v", loaded, err)
	}
}

func TestPollMongoIntegrationFindsDuplicateKeys(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewPollRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new poll repository: %v", err)
	}
	ctx := context.Background()
	for range 2 {
		if _, err := database.Collection(PollCollectionName).InsertOne(ctx, bson.D{
			{Key: "guild", Value: "guild-1"}, {Key: "messageid", Value: "message-1"},
		}); err != nil {
			t.Fatalf("seed duplicate poll: %v", err)
		}
	}
	if _, err := database.Collection(PollCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-2"}, {Key: "messageid", Value: "message-2"},
	}); err != nil {
		t.Fatalf("seed unique poll: %v", err)
	}
	duplicates, err := repository.FindDuplicateKeys(ctx)
	if err != nil || len(duplicates) != 1 || duplicates[0] != "guild-1:message-1" {
		t.Fatalf("poll duplicate keys = %#v err=%v", duplicates, err)
	}
}
