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
	create.GuildID = strings.TrimSpace(create.GuildID)
	create.MessageID = strings.TrimSpace(create.MessageID)
	create.CreatorID = strings.TrimSpace(create.CreatorID)
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
	var document documents.PollReadDocument
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
	var document documents.PollReadDocument
	after := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.collection.FindOneAndUpdate(
		ctx,
		removeFilter,
		bson.D{{Key: "$pull", Value: bson.D{{Key: "join_member", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}}},
		after,
	).Decode(&document)
	if err == nil {
		return domain.PollVoteChange{Removed: true, Poll: document.ToDomain()}, ctx.Err()
	}
	if !errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.PollVoteChange{}, mhcatmongo.MapError(fmt.Errorf("remove poll vote: %w", err))
	}

	addFilter := pollAddVoteFilter(guildID, messageID, userID, choice)
	err = r.collection.FindOneAndUpdate(
		ctx,
		addFilter,
		bson.D{{Key: "$push", Value: bson.D{{Key: "join_member", Value: bson.D{
			{Key: "id", Value: userID},
			{Key: "choise", Value: choice},
			{Key: "time", Value: voteTime},
		}}}}},
		after,
	).Decode(&document)
	if err == nil {
		return domain.PollVoteChange{Added: true, Poll: document.ToDomain()}, ctx.Err()
	}
	if !errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.PollVoteChange{}, mhcatmongo.MapError(fmt.Errorf("add poll vote: %w", err))
	}

	return domain.PollVoteChange{}, r.voteMissReason(ctx, guildID, messageID, userID, choice)
}

func pollRemoveVoteFilter(guildID string, messageID string, userID string, choice string) bson.D {
	return append(pollActiveFilter(guildID, messageID),
		bson.E{Key: "can_change_choose", Value: bson.D{{Key: "$in", Value: pollMongooseTrueValues()}}},
		bson.E{Key: "join_member", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}},
	)
}

func pollAddVoteFilter(guildID string, messageID string, userID string, choice string) bson.D {
	return append(pollActiveFilter(guildID, messageID),
		bson.E{Key: "choose_data", Value: choice},
		bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "join_member", Value: bson.D{{Key: "$type", Value: "array"}}}},
			bson.D{{Key: "join_member", Value: bson.D{{Key: "$exists", Value: false}}}},
		}},
		bson.E{Key: "join_member", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "id", Value: userID}, {Key: "choise", Value: choice}}}}}}},
		bson.E{Key: "$expr", Value: bson.D{{Key: "$lt", Value: bson.A{
			bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$isArray", Value: "$join_member"}},
					"$join_member",
					bson.A{},
				}}}},
				{Key: "as", Value: "vote"},
				{Key: "cond", Value: bson.D{{Key: "$eq", Value: bson.A{"$$vote.id", userID}}}},
			}}}}},
			pollMaxChoicesExpression(),
		}}}},
	)
}

func (r *PollRepository) TogglePoll(ctx context.Context, guildID string, messageID string, toggle domain.PollToggle) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	field := ""
	oneWay := false
	switch toggle {
	case domain.PollTogglePublicResult:
		field = "can_see_result"
	case domain.PollToggleChangeChoice:
		field = "can_change_choose"
	case domain.PollToggleAnonymous:
		field = "anonymous"
		oneWay = true
	case domain.PollToggleEnd:
		field = "end"
	default:
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	filter := pollKeyFilter(guildID, messageID)
	if oneWay {
		filter = append(filter, bson.E{Key: field, Value: bson.D{{Key: "$nin", Value: pollMongooseTrueValues()}}})
	}
	var document documents.PollReadDocument
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.collection.FindOneAndUpdate(ctx, filter, pollTogglePipeline(field, oneWay), opts).Decode(&document)
	if err == nil {
		return document.ToDomain(), ctx.Err()
	}
	mapped := mhcatmongo.MapError(err)
	if !mhcatmongo.ErrorIs(mapped, mhcatmongo.ErrorKindNotFound) {
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("toggle poll: %w", err))
	}
	if oneWay {
		poll, getErr := r.GetPoll(ctx, guildID, messageID)
		if getErr != nil {
			return domain.Poll{}, getErr
		}
		if poll.Anonymous {
			return domain.Poll{}, ports.ErrPollAnonymousLocked
		}
	}
	return domain.Poll{}, ports.ErrPollNotFound
}

func pollTogglePipeline(field string, oneWay bool) drivermongo.Pipeline {
	value := any(true)
	if !oneWay {
		value = bson.D{{Key: "$not", Value: bson.A{
			bson.D{{Key: "$in", Value: bson.A{"$" + field, pollMongooseTrueValues()}}},
		}}}
	}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: bson.D{{Key: field, Value: value}}}}}
}

func (r *PollRepository) SetMaxChoices(ctx context.Context, guildID string, messageID string, maxChoices int) (domain.Poll, error) {
	if err := ctx.Err(); err != nil {
		return domain.Poll{}, err
	}
	if maxChoices < 1 {
		return domain.Poll{}, domain.ErrInvalidPoll
	}
	var document documents.PollReadDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		pollKeyFilter(guildID, messageID),
		bson.D{{Key: "$set", Value: bson.D{{Key: "many_choose", Value: maxChoices}}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&document)
	if err == nil {
		return document.ToDomain(), ctx.Err()
	}
	if !errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.Poll{}, mhcatmongo.MapError(fmt.Errorf("set poll max choices: %w", err))
	}
	return domain.Poll{}, ports.ErrPollNotFound
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
	return append(pollKeyFilter(guildID, messageID), bson.E{Key: "end", Value: bson.D{{Key: "$nin", Value: pollMongooseTrueValues()}}})
}

func pollMongooseTrueValues() bson.A {
	return bson.A{true, "true", 1, "1", "yes"}
}

func pollMaxChoicesExpression() bson.D {
	converted := bson.D{{Key: "$convert", Value: bson.D{
		{Key: "input", Value: "$many_choose"},
		{Key: "to", Value: "double"},
		{Key: "onError", Value: 1},
		{Key: "onNull", Value: 1},
	}}}
	return bson.D{{Key: "$let", Value: bson.D{
		{Key: "vars", Value: bson.D{{Key: "maxChoices", Value: converted}}},
		{Key: "in", Value: bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{"$$maxChoices", 0}}},
			"$$maxChoices",
			1,
		}}}},
	}}}
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
