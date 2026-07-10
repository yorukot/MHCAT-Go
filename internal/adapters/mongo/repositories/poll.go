package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const PollCollectionName = "polls"

type PollRepository struct {
	collection *drivermongo.Collection
}

func NewPollRepository(collection *drivermongo.Collection) (*PollRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo poll collection is required")
	}
	return &PollRepository{collection: collection}, nil
}

func NewPollRepositoryFromDatabase(database *drivermongo.Database) (*PollRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewPollRepository(database.Collection(PollCollectionName))
}

func (r *PollRepository) CreatePoll(ctx context.Context, create domain.PollCreate) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	if err := create.Validate(); err != nil {
		return domain.Poll{}, err
	}
	poll := domain.NewPoll(create)
	document := documents.PollDocumentFromDomain(poll)
	_, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("create poll: %w", err))
	}
	return poll, ctx.Err()
}

func (r *PollRepository) GetPoll(ctx context.Context, guildID string, messageID string) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	var document documents.PollDocument
	if err := r.collection.FindOne(ctx, pollKeyFilter(guildID, messageID)).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.Poll{}, ports.ErrPollNotFound
		}
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("get poll: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *PollRepository) Vote(ctx context.Context, guildID string, messageID string, userID string, choice string, voteTime string) (domain.PollVoteChange, error) {
	if err := ctx.Err(); err != nil {
		return domain.PollVoteChange{}, err
	}
	guildID = strings.TrimSpace(guildID)
	messageID = strings.TrimSpace(messageID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || messageID == "" || userID == "" || choice == "" {
		return domain.PollVoteChange{}, domain.ErrInvalidPoll
	}

	removeFilter := pollRemoveVoteFilter(guildID, messageID, userID, choice)
	removeResult, err := r.collection.UpdateOne(ctx, removeFilter, bson.D{{Key: "$pull", Value: bson.D{{Key: "join_member", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}}})
	if err != nil {
		return domain.PollVoteChange{}, mhcatmongo.MapError(fmt.Errorf("remove poll vote: %w", err))
	}
	if removeResult.ModifiedCount > 0 {
		poll, err := r.GetPoll(ctx, guildID, messageID)
		return domain.PollVoteChange{Removed: true, Poll: poll}, err
	}

	addFilter := pollAddVoteFilter(guildID, messageID, userID, choice)
	addResult, err := r.collection.UpdateOne(ctx, addFilter, bson.D{{Key: "$push", Value: bson.D{{Key: "join_member", Value: bson.D{
		{Key: "id", Value: userID},
		{Key: "choise", Value: choice},
		{Key: "time", Value: voteTime},
	}}}}})
	if err != nil {
		return domain.PollVoteChange{}, mhcatmongo.MapError(fmt.Errorf("add poll vote: %w", err))
	}
	if addResult.ModifiedCount > 0 {
		poll, err := r.GetPoll(ctx, guildID, messageID)
		return domain.PollVoteChange{Added: true, Poll: poll}, err
	}

	return domain.PollVoteChange{}, r.voteMissReason(ctx, guildID, messageID, userID, choice)
}

func pollRemoveVoteFilter(guildID string, messageID string, userID string, choice string) bson.D {
	return append(pollActiveFilter(guildID, messageID),
		bson.E{Key: "can_change_choose", Value: true},
		bson.E{Key: "join_member", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}},
	)
}

func pollAddVoteFilter(guildID string, messageID string, userID string, choice string) bson.D {
	return append(pollActiveFilter(guildID, messageID),
		bson.E{Key: "choose_data", Value: choice},
		bson.E{Key: "join_member", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}}}},
		bson.E{Key: "$expr", Value: bson.D{{Key: "$lt", Value: bson.A{
			bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$join_member"},
				{Key: "as", Value: "vote"},
				{Key: "cond", Value: bson.D{{Key: "$eq", Value: bson.A{"$$vote.id", userID}}}},
			}}}}},
			"$many_choose",
		}}}},
	)
}

func (r *PollRepository) TogglePoll(ctx context.Context, guildID string, messageID string, toggle domain.PollToggle) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	poll, err := r.GetPoll(ctx, guildID, messageID)
	if err != nil {
		return domain.Poll{}, err
	}
	update := bson.D{}
	switch toggle {
	case domain.PollTogglePublicResult:
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "can_see_result", Value: !poll.CanSeeResult}}}}
	case domain.PollToggleChangeChoice:
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "can_change_choose", Value: !poll.CanChangeChoice}}}}
	case domain.PollToggleAnonymous:
		if poll.Anonymous {
			return domain.Poll{}, ports.ErrPollAnonymousLocked
		}
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "anonymous", Value: true}}}}
	case domain.PollToggleEnd:
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "end", Value: !poll.Ended}}}}
	default:
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	result, err := r.collection.UpdateOne(ctx, pollKeyFilter(guildID, messageID), update)
	if err != nil {
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("toggle poll: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.Poll{}, ports.ErrPollNotFound
	}
	return r.GetPoll(ctx, guildID, messageID)
}

func (r *PollRepository) SetMaxChoices(ctx context.Context, guildID string, messageID string, maxChoices int) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	if maxChoices < 1 {
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	result, err := r.collection.UpdateOne(ctx, pollKeyFilter(guildID, messageID), bson.D{{Key: "$set", Value: bson.D{{Key: "many_choose", Value: maxChoices}}}})
	if err != nil {
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("set poll max choices: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.Poll{}, ports.ErrPollNotFound
	}
	return r.GetPoll(ctx, guildID, messageID)
}

func (r *PollRepository) voteMissReason(ctx context.Context, guildID string, messageID string, userID string, choice string) error {
	poll, err := r.GetPoll(ctx, guildID, messageID)
	if err != nil {
		return err
	}
	if poll.Ended {
		return ports.ErrPollEnded
	}
	foundChoice := false
	for _, pollChoice := range poll.Choices {
		if pollChoice == choice {
			foundChoice = true
			break
		}
	}
	if !foundChoice {
		return ports.ErrPollChoiceNotFound
	}
	userChoices := poll.UserChoices(userID)
	for _, existing := range userChoices {
		if existing == choice && !poll.CanChangeChoice {
			return ports.ErrPollChangeNotAllowed
		}
	}
	if len(userChoices) >= poll.MaxChoices {
		return ports.ErrPollChoiceLimit
	}
	return ports.ErrPollChoiceNotFound
}

func pollKeyFilter(guildID string, messageID string) bson.D {
	return bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}, {Key: "messageid", Value: strings.TrimSpace(messageID)}}
}

func pollActiveFilter(guildID string, messageID string) bson.D {
	return append(pollKeyFilter(guildID, messageID), bson.E{Key: "end", Value: bson.D{{Key: "$ne", Value: true}}})
}

func (r *PollRepository) FindDuplicateKeys(ctx context.Context) ([]string, error) {
	pipeline := drivermongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "guild", Value: "$guild"}, {Key: "messageid", Value: "$messageid"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "count", Value: bson.D{{Key: "$gt", Value: 1}}}}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("audit poll duplicate keys: %w", err))
	}
	defer cursor.Close(ctx)
	var duplicates []string
	for cursor.Next(ctx) {
		var row struct {
			ID struct {
				Guild     string `bson:"guild"`
				MessageID string `bson:"messageid"`
			} `bson:"_id"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode poll duplicate key: %w", err))
		}
		duplicates = append(duplicates, row.ID.Guild+":"+row.ID.MessageID)
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate poll duplicate keys: %w", err))
	}
	return duplicates, ctx.Err()
}
