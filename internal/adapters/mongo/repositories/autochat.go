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

const AutoChatConfigCollectionName = "chats"

type AutoChatConfigRepository struct {
	collection *drivermongo.Collection
}

func NewAutoChatConfigRepository(collection *drivermongo.Collection) (*AutoChatConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo autochat collection is required")
	}
	return &AutoChatConfigRepository{collection: collection}, nil
}

func NewAutoChatConfigRepositoryFromDatabase(database *drivermongo.Database) (*AutoChatConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAutoChatConfigRepository(database.Collection(AutoChatConfigCollectionName))
}

func (r *AutoChatConfigRepository) SaveAutoChatConfig(ctx context.Context, config domain.AutoChatConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.AutoChatConfigDocumentFromDomain(config)
	update, err := mhcatAutoChatConfigUpdate(document, false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save autochat config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := mhcatAutoChatConfigUpdate(document, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert autochat config: %w", err))
	}
	return ctx.Err()
}

func (r *AutoChatConfigRepository) DeleteAutoChatConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidAutoChatConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete autochat config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrAutoChatConfigMissing
	}
	return ctx.Err()
}

func mhcatAutoChatConfigUpdate(document documents.AutoChatConfigDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().Set("channel", document.Channel)
	if upsert {
		builder.SetOnInsert("guild", document.Guild)
	}
	return builder.Build()
}

var _ ports.AutoChatConfigRepository = (*AutoChatConfigRepository)(nil)
