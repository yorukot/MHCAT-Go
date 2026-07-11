package repositories

import (
	"context"
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
)

const RedeemCodeCollectionName = "codes"

type RedeemRepository struct {
	codes    *drivermongo.Collection
	balances *drivermongo.Collection
}

func NewRedeemRepository(codes *drivermongo.Collection, balances *drivermongo.Collection) (*RedeemRepository, error) {
	if codes == nil {
		return nil, errors.New("mongo redeem code collection is required")
	}
	if balances == nil {
		return nil, errors.New("mongo redeem balance collection is required")
	}
	return &RedeemRepository{codes: codes, balances: balances}, nil
}

func NewRedeemRepositoryFromDatabase(database *drivermongo.Database) (*RedeemRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewRedeemRepository(database.Collection(RedeemCodeCollectionName), database.Collection(BalanceCollectionName))
}

func (r *RedeemRepository) GetRedeemCode(ctx context.Context, code string) (domain.RedeemCode, error) {
	if err := ctx.Err(); err != nil {
		return domain.RedeemCode{}, err
	}
	if code == "" {
		return domain.RedeemCode{}, domain.ErrInvalidRedeemCode
	}
	var document documents.RedeemCodeDocument
	if err := r.codes.FindOne(ctx, bson.D{{Key: "code", Value: code}}).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.RedeemCode{}, ports.ErrRedeemCodeNotFound
		}
		return domain.RedeemCode{}, mhcatmongo.MapError(fmt.Errorf("get redeem code: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *RedeemRepository) ConsumeRedeemCode(ctx context.Context, command domain.RedeemCommand, code domain.RedeemCode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	if err := command.Validate(); err != nil {
		return err
	}
	if code.Identity == nil || code.Code != command.Code || math.IsNaN(code.Price) {
		return domain.ErrInvalidRedeemCode
	}
	result, err := r.codes.DeleteOne(ctx, bson.D{{Key: "_id", Value: code.Identity}, {Key: "code", Value: command.Code}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("consume redeem code: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrRedeemCodeNotFound
	}
	if err := r.creditBalance(ctx, command.GuildID, code.Price); err != nil {
		return err
	}
	return ctx.Err()
}

func (r *RedeemRepository) creditBalance(ctx context.Context, guildID string, price float64) error {
	type balanceRow struct {
		ID    any           `bson:"_id"`
		Price bson.RawValue `bson:"price"`
	}
	var row balanceRow
	err := r.balances.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&row)
	next := price
	if err != nil && err != drivermongo.ErrNoDocuments {
		return mhcatmongo.MapError(fmt.Errorf("get redeem balance: %w", err))
	}
	if err == nil {
		current := documents.LegacyBalancePriceFloat(row.Price)
		if math.IsNaN(current) {
			return domain.ErrInvalidRedeemCode
		}
		next += current
		result, err := r.balances.DeleteOne(ctx, bson.D{{Key: "_id", Value: row.ID}, {Key: "guild", Value: guildID}})
		if err != nil {
			return mhcatmongo.MapError(fmt.Errorf("delete redeem balance: %w", err))
		}
		if result.DeletedCount == 0 {
			return mhcatmongo.MapError(errors.New("redeem balance changed before delete"))
		}
	}
	if _, err := r.balances.InsertOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "price", Value: next}}); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("insert redeem balance: %w", err))
	}
	return ctx.Err()
}

var _ ports.RedeemRepository = (*RedeemRepository)(nil)
