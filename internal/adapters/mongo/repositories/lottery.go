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

const LotteryCollectionName = "lotters"

type LotteryRepository struct {
	collection *drivermongo.Collection
}

func NewLotteryRepository(collection *drivermongo.Collection) (*LotteryRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo lottery collection is required")
	}
	return &LotteryRepository{collection: collection}, nil
}

func NewLotteryRepositoryFromDatabase(database *drivermongo.Database) (*LotteryRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewLotteryRepository(database.Collection(LotteryCollectionName))
}

func (r *LotteryRepository) GetLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error) {
	if err := ctx.Err(); err != nil {
		return domain.Lottery{}, err
	}
	if err := domain.ValidateLotteryKey(guildID, id); err != nil {
		return domain.Lottery{}, err
	}
	var document documents.LotteryDocument
	if err := r.collection.FindOne(ctx, lotteryKeyFilter(guildID, id)).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.Lottery{}, ports.ErrLotteryNotFound
		}
		return domain.Lottery{}, mhcatmongo.MapError(fmt.Errorf("get lottery: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *LotteryRepository) JoinLottery(ctx context.Context, request domain.LotteryJoinRequest) (domain.Lottery, error) {
	if err := ctx.Err(); err != nil {
		return domain.Lottery{}, err
	}
	request = request.Normalized()
	if err := request.Validate(); err != nil {
		return domain.Lottery{}, err
	}
	var document documents.LotteryDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		lotteryJoinFilter(request),
		lotteryJoinUpdate(request),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&document)
	if err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.Lottery{}, r.joinMissReason(ctx, request)
		}
		return domain.Lottery{}, mhcatmongo.MapError(fmt.Errorf("join lottery: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *LotteryRepository) EndLottery(ctx context.Context, guildID string, id string) (domain.Lottery, error) {
	if err := ctx.Err(); err != nil {
		return domain.Lottery{}, err
	}
	if err := domain.ValidateLotteryKey(guildID, id); err != nil {
		return domain.Lottery{}, err
	}
	var document documents.LotteryDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		lotteryKeyFilter(guildID, id),
		bson.D{{Key: "$set", Value: bson.D{{Key: "end", Value: true}}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&document)
	if err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.Lottery{}, ports.ErrLotteryNotFound
		}
		return domain.Lottery{}, mhcatmongo.MapError(fmt.Errorf("end lottery: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *LotteryRepository) joinMissReason(ctx context.Context, request domain.LotteryJoinRequest) error {
	lottery, err := r.GetLottery(ctx, request.GuildID, request.ID)
	if err != nil {
		return err
	}
	if lottery.HasParticipant(request.UserID) {
		if lottery.Ended {
			return ports.ErrLotteryEnded
		}
		return ports.ErrLotteryAlreadyJoined
	}
	if lottery.AtCapacity() {
		return ports.ErrLotteryFull
	}
	if lottery.Ended || lottery.EndsAtUnix <= 0 || lottery.EndsAtUnix < request.NowUnix {
		return ports.ErrLotteryEnded
	}
	return ports.ErrLotteryEnded
}

func lotteryKeyFilter(guildID string, id string) bson.D {
	return bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}, {Key: "id", Value: strings.TrimSpace(id)}}
}

func lotteryJoinFilter(request domain.LotteryJoinRequest) bson.D {
	maxParticipants := lotteryConvertLong("$maxNumber", int64(0))
	return append(lotteryKeyFilter(request.GuildID, request.ID),
		bson.E{Key: "end", Value: bson.D{{Key: "$nin", Value: bson.A{true, int64(1), "1", "t", "T", "TRUE", "true", "True"}}}},
		bson.E{Key: "member", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "id", Value: request.UserID}}}}}}},
		bson.E{Key: "$expr", Value: bson.D{{Key: "$and", Value: bson.A{
			bson.D{{Key: "$gte", Value: bson.A{lotteryConvertLong("$date", int64(-1)), request.NowUnix}}},
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{maxParticipants, int64(0)}}},
				bson.D{{Key: "$lt", Value: bson.A{lotteryArraySize("$member"), maxParticipants}}},
			}}},
		}}}},
	)
}

func lotteryJoinUpdate(request domain.LotteryJoinRequest) drivermongo.Pipeline {
	participant := bson.D{{Key: "id", Value: request.UserID}, {Key: "time", Value: request.JoinedAtMillis}}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: bson.D{{Key: "member", Value: bson.D{{Key: "$concatArrays", Value: bson.A{
		lotteryArrayOrEmpty("$member"),
		bson.A{participant},
	}}}}}}}}
}

func lotteryConvertLong(input string, fallback int64) bson.D {
	return bson.D{{Key: "$convert", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "to", Value: "long"},
		{Key: "onError", Value: fallback},
		{Key: "onNull", Value: fallback},
	}}}
}

func lotteryArrayOrEmpty(input string) bson.D {
	return bson.D{{Key: "$cond", Value: bson.A{
		bson.D{{Key: "$isArray", Value: input}},
		input,
		bson.A{},
	}}}
}

func lotteryArraySize(input string) bson.D {
	return bson.D{{Key: "$size", Value: lotteryArrayOrEmpty(input)}}
}

var _ ports.LotteryRepository = (*LotteryRepository)(nil)
