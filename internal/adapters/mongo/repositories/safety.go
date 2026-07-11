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
	var document documents.GoodWebConfigReadDocument
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
	config.GuildID = strings.TrimSpace(config.GuildID)
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.GoodWebConfigDocumentFromDomain(config)
	update, err := antiScamConfigUpdate(document)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateMany(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		update,
		options.UpdateMany().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save anti-scam config: %w", err))
	}
	return ctx.Err()
}

func antiScamConfigUpdate(document documents.GoodWebConfigDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("open", document.Open).
		SetOnInsert("guild", document.Guild).
		Build()
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
	if rawURL == "" {
		return false, nil
	}
	filter := scamURLContainsFilter(rawURL)
	var document documents.ScamURLReadDocument
	if err := r.collection.FindOne(ctx, filter).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return false, ctx.Err()
		}
		return false, mhcatmongo.MapError(fmt.Errorf("find scam URL report: %w", err))
	}
	web, ok := document.ToWeb()
	return ok && strings.Contains(web, rawURL), ctx.Err()
}

func (r *ScamURLCatalogRepository) FindScamURLInContent(ctx context.Context, content string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	if strings.TrimSpace(content) == "" {
		return "", false, nil
	}
	webs, err := r.ListScamURLs(ctx)
	if err != nil {
		return "", false, err
	}
	for _, web := range webs {
		if web != "" && strings.Contains(content, web) {
			return web, true, ctx.Err()
		}
	}
	return "", false, ctx.Err()
}

func (r *ScamURLCatalogRepository) ListScamURLs(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.D{}, options.Find().SetProjection(bson.D{{Key: "web", Value: 1}}))
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list scam URLs: %w", err))
	}
	defer cursor.Close(ctx)
	webs := make([]string, 0)
	for cursor.Next(ctx) {
		var document documents.ScamURLReadDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode scam URL: %w", err))
		}
		web, ok := document.ToWeb()
		if ok && web != "" {
			webs = append(webs, web)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate scam URLs: %w", err))
	}
	return webs, ctx.Err()
}

func scamURLContainsFilter(rawURL string) bson.D {
	return bson.D{{Key: "web", Value: bson.D{{Key: "$regex", Value: regexp.QuoteMeta(rawURL)}}}}
}

var _ ports.ScamURLCatalog = (*ScamURLCatalogRepository)(nil)
var _ ports.ScamURLLister = (*ScamURLCatalogRepository)(nil)
