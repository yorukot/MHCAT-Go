package repositories

import (
	"context"
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

const LegacyIdleWorkState = "待業中"

type WorkPayoutRepository struct {
	coins       *drivermongo.Collection
	giftChanges *drivermongo.Collection
	workUsers   *drivermongo.Collection
}

type workUserPayoutDocument struct {
	ID      any           `bson:"_id"`
	Guild   string        `bson:"guild"`
	User    string        `bson:"user"`
	State   string        `bson:"state"`
	EndTime bson.RawValue `bson:"end_time"`
	GetCoin bson.RawValue `bson:"get_coin"`
}

func NewWorkPayoutRepository(coins *drivermongo.Collection, giftChanges *drivermongo.Collection, workUsers *drivermongo.Collection) (*WorkPayoutRepository, error) {
	if coins == nil {
		return nil, fmt.Errorf("coins collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	if workUsers == nil {
		return nil, fmt.Errorf("work_users collection is required")
	}
	return &WorkPayoutRepository{coins: coins, giftChanges: giftChanges, workUsers: workUsers}, nil
}

func NewWorkPayoutRepositoryFromDatabase(database *drivermongo.Database) (*WorkPayoutRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewWorkPayoutRepository(
		database.Collection(CoinCollectionName),
		database.Collection(GiftChangeCollectionName),
		database.Collection(WorkUserCollectionName),
	)
}

func (r *WorkPayoutRepository) PreviewWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkPayoutResult{}, err
	}
	if nowUnix <= 0 {
		return domain.WorkPayoutResult{}, domain.ErrInvalidWorkPayout
	}
	eligible, err := r.workUsers.CountDocuments(ctx, workPayoutEligibleFilter(nowUnix))
	if err != nil {
		return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("preview work payout: %w", err))
	}
	return domain.WorkPayoutResult{EligibleJobs: eligible}, ctx.Err()
}

func (r *WorkPayoutRepository) RunWorkPayout(ctx context.Context, nowUnix int64) (domain.WorkPayoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkPayoutResult{}, err
	}
	if nowUnix <= 0 {
		return domain.WorkPayoutResult{}, domain.ErrInvalidWorkPayout
	}
	cursor, err := r.workUsers.Find(ctx, workPayoutEligibleFilter(nowUnix))
	if err != nil {
		return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("list due work payouts: %w", err))
	}
	defer cursor.Close(ctx)
	var result domain.WorkPayoutResult
	for cursor.Next(ctx) {
		result.EligibleJobs++
		var document workUserPayoutDocument
		if err := cursor.Decode(&document); err != nil {
			return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("decode due work payout: %w", err))
		}
		if !validWorkPayoutDocument(document) {
			result.SkippedInvalidJobs++
			continue
		}
		today, err := r.workPayoutTodayValue(ctx, document.Guild, nowUnix)
		if err != nil {
			return domain.WorkPayoutResult{}, err
		}
		coinResult, err := r.coins.UpdateOne(
			ctx,
			bson.D{{Key: "guild", Value: document.Guild}, {Key: "member", Value: document.User}},
			bson.D{
				{Key: "$inc", Value: bson.D{{Key: "coin", Value: rawInt64(document.GetCoin)}}},
				{Key: "$setOnInsert", Value: bson.D{{Key: "guild", Value: document.Guild}, {Key: "member", Value: document.User}, {Key: "today", Value: today}}},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("apply work payout coin for guild %s user %s: %w", document.Guild, document.User, err))
		}
		stateResult, err := r.workUsers.UpdateOne(ctx, workPayoutStateResetFilter(document, nowUnix), bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: LegacyIdleWorkState}}}})
		if err != nil {
			return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("reset work state for guild %s user %s: %w", document.Guild, document.User, err))
		}
		result.CoinMatched += coinResult.MatchedCount
		result.CoinModified += coinResult.ModifiedCount
		result.CoinUpserted += coinResult.UpsertedCount
		result.StateMatched += stateResult.MatchedCount
		result.StateModified += stateResult.ModifiedCount
		if stateResult.MatchedCount == 0 {
			return result, fmt.Errorf("%w: reset matched no document for guild %s user %s", domain.ErrWorkPayoutStateConflict, document.Guild, document.User)
		}
		result.ProcessedJobs++
	}
	if err := cursor.Err(); err != nil {
		return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("iterate work payouts: %w", err))
	}
	return result, ctx.Err()
}

func (r *WorkPayoutRepository) workPayoutTodayValue(ctx context.Context, guildID string, nowUnix int64) (int64, error) {
	var document documents.GiftChangeDocument
	err := r.giftChanges.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return workPayoutTodayFromConfig(false, 0, nowUnix), nil
		}
		return 0, mhcatmongo.MapError(fmt.Errorf("get work payout economy config for guild %s: %w", guildID, err))
	}
	config := document.ToDomain()
	return workPayoutTodayFromConfig(true, config.ResetMarker, nowUnix), ctx.Err()
}

func workPayoutEligibleFilter(nowUnix int64) bson.D {
	return bson.D{
		{Key: "state", Value: bson.D{{Key: "$ne", Value: LegacyIdleWorkState}}},
		{Key: "end_time", Value: bson.D{{Key: "$lt", Value: nowUnix}}},
	}
}

func workPayoutStateResetFilter(document workUserPayoutDocument, nowUnix int64) bson.D {
	if document.ID != nil {
		return bson.D{
			{Key: "_id", Value: document.ID},
			{Key: "state", Value: bson.D{{Key: "$ne", Value: LegacyIdleWorkState}}},
			{Key: "end_time", Value: bson.D{{Key: "$lt", Value: nowUnix}}},
		}
	}
	return bson.D{
		{Key: "guild", Value: document.Guild},
		{Key: "user", Value: document.User},
		{Key: "state", Value: bson.D{{Key: "$ne", Value: LegacyIdleWorkState}}},
		{Key: "end_time", Value: bson.D{{Key: "$lt", Value: nowUnix}}},
	}
}

func workPayoutTodayFromConfig(configFound bool, resetMarker int64, nowUnix int64) int64 {
	if !configFound || resetMarker == 0 {
		return 1
	}
	return nowUnix
}

func validWorkPayoutDocument(document workUserPayoutDocument) bool {
	return strings.TrimSpace(document.Guild) != "" &&
		strings.TrimSpace(document.User) != "" &&
		strings.TrimSpace(document.State) != LegacyIdleWorkState &&
		rawInt64(document.EndTime) > 0
}

var _ ports.WorkPayoutRepository = (*WorkPayoutRepository)(nil)
