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

var _ ports.GachaPrizePoolRepository = (*GachaRepository)(nil)
var _ ports.GachaPrizeDeleteRepository = (*GachaRepository)(nil)
