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
)

const GiftCollectionName = "gifts"

type GachaRepository struct {
	gifts       *drivermongo.Collection
	giftChanges *drivermongo.Collection
}

func NewGachaRepository(gifts *drivermongo.Collection, giftChanges *drivermongo.Collection) (*GachaRepository, error) {
	if gifts == nil {
		return nil, fmt.Errorf("gifts collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	return &GachaRepository{gifts: gifts, giftChanges: giftChanges}, nil
}

func NewGachaRepositoryFromDatabase(database *drivermongo.Database) (*GachaRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewGachaRepository(database.Collection(GiftCollectionName), database.Collection(GiftChangeCollectionName))
}

func (r *GachaRepository) ListGachaPrizes(ctx context.Context, guildID string) ([]domain.GachaPrize, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidGachaQuery
	}
	cursor, err := r.gifts.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list gacha prizes: %w", err))
	}
	defer cursor.Close(ctx)
	var prizes []domain.GachaPrize
	for cursor.Next(ctx) {
		var document documents.GiftDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode gacha prize: %w", err))
		}
		prizes = append(prizes, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate gacha prizes: %w", err))
	}
	return prizes, ctx.Err()
}

func (r *GachaRepository) GetGachaConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.EconomyConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.EconomyConfig{}, domain.ErrInvalidGachaQuery
	}
	var document documents.GiftChangeDocument
	if err := r.giftChanges.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.EconomyConfig{GuildID: guildID}, ports.ErrEconomyConfigMissing
		}
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("get gacha config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *GachaRepository) DeleteGachaPrize(ctx context.Context, guildID string, prizeName string) (domain.GachaPrize, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaPrize{}, err
	}
	guildID = strings.TrimSpace(guildID)
	prizeName = strings.TrimSpace(prizeName)
	if guildID == "" || prizeName == "" {
		return domain.GachaPrize{}, domain.ErrInvalidGachaQuery
	}
	var document documents.GiftDocument
	err := r.gifts.FindOneAndDelete(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "gift_name", Value: prizeName}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.GachaPrize{}, ports.ErrGachaPrizeMissing
		}
		return domain.GachaPrize{}, mhcatmongo.MapError(fmt.Errorf("delete gacha prize: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *GachaRepository) CountGachaPrizes(ctx context.Context, guildID string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return 0, domain.ErrInvalidGachaQuery
	}
	count, err := r.gifts.CountDocuments(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return 0, mhcatmongo.MapError(fmt.Errorf("count gacha prizes: %w", err))
	}
	return count, ctx.Err()
}

func (r *GachaRepository) CreateGachaPrize(ctx context.Context, prize domain.GachaPrizeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	prize.GuildID = strings.TrimSpace(prize.GuildID)
	prize.Name = strings.TrimSpace(prize.Name)
	if prize.GuildID == "" || prize.Name == "" || prize.Count <= 0 {
		return domain.ErrInvalidGachaPrize
	}
	filter := bson.D{{Key: "guild", Value: prize.GuildID}, {Key: "gift_name", Value: prize.Name}}
	err := r.gifts.FindOne(ctx, filter).Err()
	if err == nil {
		return ports.ErrGachaPrizeExists
	}
	if err != drivermongo.ErrNoDocuments {
		return mhcatmongo.MapError(fmt.Errorf("find gacha prize before create: %w", err))
	}
	if _, err := r.gifts.InsertOne(ctx, documents.GiftWriteDocumentFromDomain(prize)); err != nil {
		mapped := mhcatmongo.MapError(fmt.Errorf("create gacha prize: %w", err))
		if mhcatmongo.ErrorIs(mapped, mhcatmongo.ErrorKindConflict) {
			return ports.ErrGachaPrizeExists
		}
		return mapped
	}
	return ctx.Err()
}

var _ ports.GachaPrizePoolRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeDeleteRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeCreateRepository = (*GachaRepository)(nil)
