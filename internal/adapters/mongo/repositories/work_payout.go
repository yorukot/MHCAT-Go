package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
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

const WorkPayoutMarkerField = "mhcat_work_payouts"

const workPayoutMarkerVersion = "v1"

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

type workPayoutIdentity struct {
	MarkerKey string
	Token     string
	EndTime   float64
	Reward    float64
}

type workPayoutCoinTarget struct {
	ID     any
	Upsert bool
}

type workPayoutCoinIDDocument struct {
	ID any `bson:"_id"`
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
		if !validWorkPayoutDocument(document, nowUnix) {
			result.SkippedInvalidJobs++
			continue
		}
		identity, err := newWorkPayoutIdentity(document)
		if err != nil {
			return domain.WorkPayoutResult{}, fmt.Errorf("build work payout identity for guild %s user %s: %w", document.Guild, document.User, err)
		}
		coinTarget, err := r.workPayoutCoinTarget(ctx, document.Guild, document.User)
		if err != nil {
			return result, err
		}
		today, err := r.workPayoutTodayValue(ctx, document.Guild, nowUnix)
		if err != nil {
			return domain.WorkPayoutResult{}, err
		}
		coinResult, err := r.coins.UpdateOne(
			ctx,
			workPayoutCoinFilter(coinTarget.ID, document.Guild, document.User, identity),
			workPayoutCoinPipeline(document.Guild, document.User, today, identity),
			options.UpdateOne().SetUpsert(coinTarget.Upsert),
		)
		if err != nil {
			return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("apply work payout coin for guild %s user %s: %w", document.Guild, document.User, err))
		}
		if coinResult.MatchedCount == 0 && coinResult.UpsertedCount == 0 {
			return result, fmt.Errorf("%w: payout marker rejected guild %s user %s work_user %v", domain.ErrWorkPayoutMarkerConflict, document.Guild, document.User, document.ID)
		}
		stateResult, err := r.workUsers.UpdateOne(ctx, workPayoutStateResetFilter(document, nowUnix), bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: LegacyIdleWorkState}}}})
		if err != nil {
			return domain.WorkPayoutResult{}, mhcatmongo.MapError(fmt.Errorf("reset work state for guild %s user %s: %w", document.Guild, document.User, err))
		}
		result.CoinMatched += coinResult.MatchedCount
		result.CoinModified += coinResult.ModifiedCount
		result.CoinUpserted += coinResult.UpsertedCount
		if coinResult.MatchedCount == 1 && coinResult.ModifiedCount == 0 {
			result.IdempotentReplays++
		}
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

func (r *WorkPayoutRepository) workPayoutCoinTarget(ctx context.Context, guildID string, userID string) (workPayoutCoinTarget, error) {
	var document workPayoutCoinIDDocument
	err := r.coins.FindOne(
		ctx,
		bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}},
		options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}}),
	).Decode(&document)
	if err == nil {
		return workPayoutCoinTarget{ID: document.ID}, nil
	}
	if err != drivermongo.ErrNoDocuments {
		return workPayoutCoinTarget{}, mhcatmongo.MapError(fmt.Errorf("resolve work payout coin for guild %s user %s: %w", guildID, userID, err))
	}
	id, err := newWorkPayoutCoinID(guildID, userID)
	if err != nil {
		return workPayoutCoinTarget{}, fmt.Errorf("build work payout coin id for guild %s user %s: %w", guildID, userID, err)
	}
	return workPayoutCoinTarget{ID: id, Upsert: true}, nil
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
	return bson.D{
		{Key: "_id", Value: document.ID},
		{Key: "guild", Value: document.Guild},
		{Key: "user", Value: document.User},
		{Key: "state", Value: document.State},
		{Key: "end_time", Value: bson.D{
			{Key: "$eq", Value: document.EndTime},
			{Key: "$lt", Value: nowUnix},
		}},
		{Key: "get_coin", Value: document.GetCoin},
	}
}

func workPayoutCoinFilter(coinID any, guildID string, userID string, identity workPayoutIdentity) bson.D {
	markerPath := WorkPayoutMarkerField + "." + identity.MarkerKey
	markerGuard := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: markerPath, Value: bson.D{{Key: "$exists", Value: false}}}},
		bson.D{{Key: markerPath + ".token", Value: identity.Token}},
		bson.D{{Key: markerPath + ".end_time", Value: bson.D{{Key: "$lt", Value: identity.EndTime}}}},
	}}}
	return bson.D{
		{Key: "_id", Value: coinID},
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "$and", Value: bson.A{markerGuard}},
	}
}

func workPayoutCoinPipeline(guildID string, userID string, today int64, identity workPayoutIdentity) drivermongo.Pipeline {
	markerPath := WorkPayoutMarkerField + "." + identity.MarkerKey
	markerTokenPath := "$" + markerPath + ".token"
	sameToken := bson.D{{Key: "$eq", Value: bson.A{markerTokenPath, identity.Token}}}
	coinBase := bson.D{{Key: "$convert", Value: bson.D{
		{Key: "input", Value: "$coin"},
		{Key: "to", Value: "double"},
		{Key: "onError", Value: math.NaN()},
		{Key: "onNull", Value: float64(0)},
	}}}
	todayMissing := bson.D{{Key: "$eq", Value: bson.A{bson.D{{Key: "$type", Value: "$today"}}, "missing"}}}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "coin", Value: bson.D{{Key: "$cond", Value: bson.A{
			sameToken,
			"$coin",
			bson.D{{Key: "$add", Value: bson.A{coinBase, identity.Reward}}},
		}}}},
		{Key: "today", Value: bson.D{{Key: "$cond", Value: bson.A{todayMissing, today, "$today"}}}},
		{Key: markerPath, Value: bson.D{
			{Key: "token", Value: identity.Token},
			{Key: "end_time", Value: identity.EndTime},
		}},
	}}}}
}

func newWorkPayoutIdentity(document workUserPayoutDocument) (workPayoutIdentity, error) {
	markerDigest, err := workPayoutDigest("mhcat-work-payout-marker-v1", document.ID)
	if err != nil {
		return workPayoutIdentity{}, err
	}
	endTime := workPayoutNumber(document.EndTime)
	reward := workPayoutNumber(document.GetCoin)
	tokenDigest, err := workPayoutDigest(
		"mhcat-work-payout-token-v1",
		document.ID,
		document.Guild,
		document.User,
		document.State,
		document.EndTime,
		document.GetCoin,
	)
	if err != nil {
		return workPayoutIdentity{}, err
	}
	return workPayoutIdentity{
		MarkerKey: workPayoutMarkerVersion + "_" + hex.EncodeToString(markerDigest),
		Token:     workPayoutMarkerVersion + "_" + hex.EncodeToString(tokenDigest),
		EndTime:   endTime,
		Reward:    reward,
	}, nil
}

func newWorkPayoutCoinID(guildID string, userID string) (bson.ObjectID, error) {
	digest, err := workPayoutDigest("mhcat-work-payout-coin-id-v1", guildID, userID)
	if err != nil {
		return bson.NilObjectID, err
	}
	var id bson.ObjectID
	copy(id[:], digest[:len(id)])
	return id, nil
}

func workPayoutDigest(namespace string, values ...any) ([]byte, error) {
	digest := sha256.New()
	digest.Write([]byte(namespace))
	for _, value := range values {
		valueType, encoded, err := bson.MarshalValue(value)
		if err != nil {
			return nil, fmt.Errorf("marshal digest value: %w", err)
		}
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(encoded)))
		digest.Write([]byte{byte(valueType)})
		digest.Write(length[:])
		digest.Write(encoded)
	}
	return digest.Sum(nil), nil
}

func workPayoutTodayFromConfig(configFound bool, resetMarker int64, nowUnix int64) int64 {
	if !configFound || resetMarker == 0 {
		return 1
	}
	return nowUnix
}

func validWorkPayoutDocument(document workUserPayoutDocument, nowUnix int64) bool {
	return document.ID != nil &&
		strings.TrimSpace(document.Guild) != "" &&
		strings.TrimSpace(document.User) != "" &&
		strings.TrimSpace(document.State) != LegacyIdleWorkState &&
		workPayoutNumber(document.EndTime) < float64(nowUnix)
}

func workPayoutNumber(value bson.RawValue) float64 {
	switch value.Type {
	case bson.TypeDouble:
		return value.Double()
	case bson.TypeInt32:
		return float64(value.Int32())
	case bson.TypeInt64:
		return float64(value.Int64())
	case bson.TypeString:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value.StringValue()), 64)
		if err != nil {
			return math.NaN()
		}
		return parsed
	case bson.TypeNull:
		return 0
	case bson.TypeBoolean:
		if value.Boolean() {
			return 1
		}
		return 0
	default:
		return math.NaN()
	}
}

var _ ports.WorkPayoutRepository = (*WorkPayoutRepository)(nil)
