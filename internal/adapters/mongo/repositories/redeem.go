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

func (r *RedeemRepository) ConsumeRedeemCode(ctx context.Context, command domain.RedeemCommand, price float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	if err := command.Validate(); err != nil {
		return err
	}
	if price < 0 {
		return domain.ErrInvalidRedeemCode
	}
	result, err := r.codes.DeleteOne(ctx, bson.D{{Key: "code", Value: command.Code}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("consume redeem code: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrRedeemCodeNotFound
	}
	if err := r.creditBalances(ctx, command.GuildID, price); err != nil {
		return err
	}
	return ctx.Err()
}

func (r *RedeemRepository) creditBalances(ctx context.Context, guildID string, price float64) error {
	cursor, err := r.balances.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("list redeem balances: %w", err))
	}
	defer cursor.Close(ctx)
	type balanceRow struct {
		ID    bson.ObjectID `bson:"_id"`
		Price bson.RawValue `bson:"price"`
	}
	rows := []balanceRow{}
	for cursor.Next(ctx) {
		var row balanceRow
		if err := cursor.Decode(&row); err != nil {
			return mhcatmongo.MapError(fmt.Errorf("decode redeem balance: %w", err))
		}
		rows = append(rows, row)
	}
	if err := cursor.Err(); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("iterate redeem balances: %w", err))
	}
	if len(rows) == 0 {
		_, err := r.balances.UpdateOne(ctx,
			bson.D{{Key: "guild", Value: guildID}},
			bson.D{
				{Key: "$setOnInsert", Value: bson.D{{Key: "guild", Value: guildID}}},
				{Key: "$set", Value: bson.D{{Key: "price", Value: price}}},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return mhcatmongo.MapError(fmt.Errorf("upsert redeem balance: %w", err))
		}
		return ctx.Err()
	}
	for _, row := range rows {
		next := documents.LegacyBalancePriceFloat(row.Price) + price
		if _, err := r.balances.UpdateOne(ctx, bson.D{{Key: "_id", Value: row.ID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "price", Value: next}}}}); err != nil {
			return mhcatmongo.MapError(fmt.Errorf("update redeem balance: %w", err))
		}
	}
	return ctx.Err()
}

var _ ports.RedeemRepository = (*RedeemRepository)(nil)
