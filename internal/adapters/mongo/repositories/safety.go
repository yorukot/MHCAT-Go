package repositories

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const GoodWebConfigCollectionName = "good_webs"
const NotAGoodWebCollectionName = "not_a_good_webs"

type AntiScamConfigRepository struct {
	collection *drivermongo.Collection
}

func NewAntiScamConfigRepository(collection *drivermongo.Collection) (*AntiScamConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo anti-scam config collection is required")
	}
	return &AntiScamConfigRepository{collection: collection}, nil
}

func NewAntiScamConfigRepositoryFromDatabase(database *drivermongo.Database) (*AntiScamConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAntiScamConfigRepository(database.Collection(GoodWebConfigCollectionName))
}

func (r *AntiScamConfigRepository) FindAntiScamConfig(ctx context.Context, guildID string) (domain.AntiScamConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AntiScamConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.AntiScamConfig{}, domain.ErrInvalidAntiScamConfig
	}
	var document documents.GoodWebConfigDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.AntiScamConfig{}, ports.ErrAntiScamConfigMissing
		}
		return domain.AntiScamConfig{}, mhcatmongo.MapError(fmt.Errorf("find anti-scam config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *AntiScamConfigRepository) SaveAntiScamConfig(ctx context.Context, config domain.AntiScamConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.GoodWebConfigDocumentFromDomain(config)
	update, err := antiScamConfigUpdate(document, false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save anti-scam config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := antiScamConfigUpdate(document, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert anti-scam config: %w", err))
	}
	return ctx.Err()
}

func antiScamConfigUpdate(document documents.GoodWebConfigDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().Set("open", document.Open)
	if upsert {
		builder.SetOnInsert("guild", document.Guild)
	}
	return builder.Build()
}

var _ ports.AntiScamConfigRepository = (*AntiScamConfigRepository)(nil)

type ScamURLCatalogRepository struct {
	collection *drivermongo.Collection
}

func NewScamURLCatalogRepository(collection *drivermongo.Collection) (*ScamURLCatalogRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo scam URL catalog collection is required")
	}
	return &ScamURLCatalogRepository{collection: collection}, nil
}

func NewScamURLCatalogRepositoryFromDatabase(database *drivermongo.Database) (*ScamURLCatalogRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewScamURLCatalogRepository(database.Collection(NotAGoodWebCollectionName))
}

func (r *ScamURLCatalogRepository) ContainsScamURL(ctx context.Context, rawURL string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false, nil
	}
	filter := scamURLContainsFilter(rawURL)
	var document documents.ScamURLDocument
	if err := r.collection.FindOne(ctx, filter).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return false, ctx.Err()
		}
		return false, mhcatmongo.MapError(fmt.Errorf("find scam URL report: %w", err))
	}
	return strings.Contains(document.Web, rawURL), ctx.Err()
}

func scamURLContainsFilter(rawURL string) bson.D {
	return bson.D{{Key: "web", Value: bson.D{{Key: "$regex", Value: regexp.QuoteMeta(strings.TrimSpace(rawURL))}}}}
}

var _ ports.ScamURLCatalog = (*ScamURLCatalogRepository)(nil)
