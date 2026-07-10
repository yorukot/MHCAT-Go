package repositories

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const AutoChatPaidCollectionName = "chatgpts"

const (
	legacyAutoChatBusyMillis  int64 = 10_000
	legacyAutoChatResetMillis int64 = 40_000
)

type AutoChatPaidRepository struct {
	balances     *drivermongo.Collection
	handoffs     *drivermongo.Collection
	transactions ports.TransactionRunner
}

type autoChatPaidBalanceDocument struct {
	ID    any           `bson:"_id"`
	Price bson.RawValue `bson:"price"`
}

type autoChatPaidHandoffTarget struct {
	ID                any
	Exists            bool
	ConversationReset bool
}

func NewAutoChatPaidRepository(balances *drivermongo.Collection, handoffs *drivermongo.Collection, transactions ports.TransactionRunner) (*AutoChatPaidRepository, error) {
	if balances == nil {
		return nil, errors.New("chatgpt_gets collection is required")
	}
	if handoffs == nil {
		return nil, errors.New("chatgpts collection is required")
	}
	if transactions == nil {
		return nil, errors.New("paid autochat transaction runner is required")
	}
	return &AutoChatPaidRepository{balances: balances, handoffs: handoffs, transactions: transactions}, nil
}

func NewAutoChatPaidRepositoryFromDatabase(database *drivermongo.Database, transactions ports.TransactionRunner) (*AutoChatPaidRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAutoChatPaidRepository(database.Collection(BalanceCollectionName), database.Collection(AutoChatPaidCollectionName), transactions)
}

func (r *AutoChatPaidRepository) QueuePaidAutoChat(ctx context.Context, request domain.AutoChatPaidRequest) (domain.AutoChatPaidDispatch, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidDispatch{}, err
	}
	if err := request.Validate(); err != nil {
		return domain.AutoChatPaidDispatch{}, err
	}
	request.GuildID = strings.TrimSpace(request.GuildID)
	dispatch := domain.AutoChatPaidDispatch{
		RequestTimeMilli: request.RequestedAtMilli,
		Cost:             request.Cost,
	}
	err := r.transactions.WithinTransaction(ctx, func(txCtx context.Context) error {
		balance, amount, err := r.resolvePaidBalance(txCtx, request.GuildID)
		if err != nil {
			return err
		}
		target, err := r.resolvePaidHandoff(txCtx, request.GuildID, request.RequestedAtMilli)
		if err != nil {
			return err
		}
		balanceResult, err := r.balances.UpdateOne(
			txCtx,
			bson.D{{Key: "_id", Value: balance.ID}, {Key: "guild", Value: request.GuildID}, {Key: "price", Value: amount}},
			bson.D{{Key: "$set", Value: bson.D{{Key: "price", Value: amount - request.Cost}}}},
		)
		if err != nil {
			return mhcatmongo.MapError(fmt.Errorf("debit paid autochat balance for guild %s: %w", request.GuildID, err))
		}
		if balanceResult.MatchedCount != 1 {
			return fmt.Errorf("%w: balance changed for guild %s", ports.ErrAutoChatPaidStateConflict, request.GuildID)
		}
		if err := r.writePaidHandoff(txCtx, target, request); err != nil {
			return err
		}
		dispatch.ConversationReset = target.ConversationReset
		return txCtx.Err()
	})
	if err != nil {
		return domain.AutoChatPaidDispatch{}, err
	}
	return dispatch, ctx.Err()
}

func (r *AutoChatPaidRepository) GetPaidAutoChatResponse(ctx context.Context, guildID string, requestTimeMilli int64) (domain.AutoChatPaidResponse, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidResponse{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || requestTimeMilli <= 0 {
		return domain.AutoChatPaidResponse{}, domain.ErrInvalidAutoChatPaidRequest
	}
	cursor, err := r.handoffs.Find(
		ctx,
		bson.D{{Key: "guild", Value: guildID}},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(2),
	)
	if err != nil {
		return domain.AutoChatPaidResponse{}, mhcatmongo.MapError(fmt.Errorf("find paid autochat response for guild %s: %w", guildID, err))
	}
	defer cursor.Close(ctx)
	var rows []documents.AutoChatPaidDocument
	if err := cursor.All(ctx, &rows); err != nil {
		return domain.AutoChatPaidResponse{}, mhcatmongo.MapError(fmt.Errorf("decode paid autochat response for guild %s: %w", guildID, err))
	}
	if len(rows) == 0 {
		return domain.AutoChatPaidResponse{}, ports.ErrAutoChatPaidResponseMissing
	}
	if len(rows) > 1 {
		return domain.AutoChatPaidResponse{}, fmt.Errorf("%w: multiple chatgpts rows for guild %s", ports.ErrAutoChatPaidStateConflict, guildID)
	}
	response, ok := rows[0].ToResponse()
	if !ok || response.RequestTimeMilli != requestTimeMilli {
		return domain.AutoChatPaidResponse{}, ports.ErrAutoChatPaidResponseMissing
	}
	return response, ctx.Err()
}

func (r *AutoChatPaidRepository) resolvePaidBalance(ctx context.Context, guildID string) (autoChatPaidBalanceDocument, float64, error) {
	cursor, err := r.balances.Find(
		ctx,
		bson.D{{Key: "guild", Value: guildID}},
		options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}, {Key: "price", Value: 1}}).SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(2),
	)
	if err != nil {
		return autoChatPaidBalanceDocument{}, 0, mhcatmongo.MapError(fmt.Errorf("resolve paid autochat balance for guild %s: %w", guildID, err))
	}
	defer cursor.Close(ctx)
	var rows []autoChatPaidBalanceDocument
	if err := cursor.All(ctx, &rows); err != nil {
		return autoChatPaidBalanceDocument{}, 0, mhcatmongo.MapError(fmt.Errorf("decode paid autochat balance for guild %s: %w", guildID, err))
	}
	if len(rows) == 0 {
		return autoChatPaidBalanceDocument{}, 0, ports.ErrAutoChatPaidBalanceUnavailable
	}
	if len(rows) > 1 {
		return autoChatPaidBalanceDocument{}, 0, fmt.Errorf("%w: multiple chatgpt_gets rows for guild %s", ports.ErrAutoChatPaidStateConflict, guildID)
	}
	amount, ok := autoChatPaidNumeric(rows[0].Price)
	if !ok || amount <= 0 {
		return autoChatPaidBalanceDocument{}, 0, ports.ErrAutoChatPaidBalanceUnavailable
	}
	return rows[0], amount, ctx.Err()
}

func (r *AutoChatPaidRepository) resolvePaidHandoff(ctx context.Context, guildID string, requestTimeMilli int64) (autoChatPaidHandoffTarget, error) {
	cursor, err := r.handoffs.Find(
		ctx,
		bson.D{{Key: "guild", Value: guildID}},
		options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}, {Key: "time", Value: 1}}).SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(2),
	)
	if err != nil {
		return autoChatPaidHandoffTarget{}, mhcatmongo.MapError(fmt.Errorf("resolve paid autochat handoff for guild %s: %w", guildID, err))
	}
	defer cursor.Close(ctx)
	var rows []documents.AutoChatPaidDocument
	if err := cursor.All(ctx, &rows); err != nil {
		return autoChatPaidHandoffTarget{}, mhcatmongo.MapError(fmt.Errorf("decode paid autochat handoff for guild %s: %w", guildID, err))
	}
	if len(rows) > 1 {
		return autoChatPaidHandoffTarget{}, fmt.Errorf("%w: multiple chatgpts rows for guild %s", ports.ErrAutoChatPaidStateConflict, guildID)
	}
	if len(rows) == 0 {
		return autoChatPaidHandoffTarget{ID: newAutoChatPaidID(guildID)}, nil
	}
	target := autoChatPaidHandoffTarget{ID: rows[0].ID, Exists: true}
	busy, reset := legacyAutoChatTiming(requestTimeMilli, rows[0].Time)
	if busy {
		return autoChatPaidHandoffTarget{}, ports.ErrAutoChatPaidBusy
	}
	target.ConversationReset = reset
	return target, nil
}

func legacyAutoChatTiming(requestTimeMilli int64, previous bson.RawValue) (busy bool, reset bool) {
	previousTime, ok := documents.LegacyMongooseNumber(previous)
	if !ok {
		return false, false
	}
	age := float64(requestTimeMilli) - previousTime
	return age < float64(legacyAutoChatBusyMillis), age > float64(legacyAutoChatResetMillis)
}

func (r *AutoChatPaidRepository) writePaidHandoff(ctx context.Context, target autoChatPaidHandoffTarget, request domain.AutoChatPaidRequest) error {
	set := bson.D{
		{Key: "guild", Value: request.GuildID},
		{Key: "reply", Value: false},
		{Key: "message", Value: request.Content},
		{Key: "time", Value: request.RequestedAtMilli},
	}
	if !target.Exists || target.ConversationReset {
		set = append(set, bson.E{Key: "resid_c", Value: nil}, bson.E{Key: "resid_p", Value: nil})
	}
	result, err := r.handoffs.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: target.ID}, {Key: "guild", Value: request.GuildID}},
		bson.D{{Key: "$set", Value: set}},
		options.UpdateOne().SetUpsert(!target.Exists),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("write paid autochat handoff for guild %s: %w", request.GuildID, err))
	}
	if (target.Exists && result.MatchedCount != 1) || (!target.Exists && result.MatchedCount == 0 && result.UpsertedCount == 0) {
		return fmt.Errorf("%w: handoff changed for guild %s", ports.ErrAutoChatPaidStateConflict, request.GuildID)
	}
	return ctx.Err()
}

func autoChatPaidNumeric(value bson.RawValue) (float64, bool) {
	var parsed float64
	switch value.Type {
	case bson.TypeDouble:
		parsed = value.Double()
	case bson.TypeInt32:
		parsed = float64(value.Int32())
	case bson.TypeInt64:
		parsed = float64(value.Int64())
	default:
		return 0, false
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	return parsed, true
}

func newAutoChatPaidID(guildID string) bson.ObjectID {
	digest := sha256.Sum256([]byte("mhcat-paid-autochat-id-v1\x00" + strings.TrimSpace(guildID)))
	var id bson.ObjectID
	copy(id[:], digest[:len(id)])
	return id
}

var _ ports.AutoChatPaidRepository = (*AutoChatPaidRepository)(nil)
