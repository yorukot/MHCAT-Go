package repositories

import (
	"context"
	"errors"
	"fmt"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const LoggingConfigCollectionName = "loggings"

type LoggingConfigRepository struct {
	collection *drivermongo.Collection
}

func NewLoggingConfigRepository(collection *drivermongo.Collection) (*LoggingConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo logging collection is required")
	}
	return &LoggingConfigRepository{collection: collection}, nil
}

func NewLoggingConfigRepositoryFromDatabase(database *drivermongo.Database) (*LoggingConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewLoggingConfigRepository(database.Collection(LoggingConfigCollectionName))
}

func (r *LoggingConfigRepository) SaveLoggingConfig(ctx context.Context, config domain.LoggingConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.LoggingConfigDocumentFromDomain(config)
	update, err := mhcatmongo.NewUpdate().
		Set("channel_id", document.ChannelID).
		Set("message_update", document.MessageUpdate).
		Set("message_delete", document.MessageDelete).
		Set("channel_update", document.ChannelUpdate).
		Set("member_voice_update", document.MemberVoiceUpdate).
		Build()
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save logging config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := mhcatmongo.NewUpdate().
		Set("channel_id", document.ChannelID).
		Set("message_update", document.MessageUpdate).
		Set("message_delete", document.MessageDelete).
		Set("channel_update", document.ChannelUpdate).
		Set("member_voice_update", document.MemberVoiceUpdate).
		SetOnInsert("guild", document.Guild).
		Build()
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		insertUpdate,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert logging config: %w", err))
	}
	return ctx.Err()
}

func (r *LoggingConfigRepository) GetLoggingConfig(ctx context.Context, guildID string) (domain.LoggingConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.LoggingConfig{}, err
	}
	var document documents.LoggingConfigReadDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.LoggingConfig{}, ports.ErrLoggingConfigMissing
		}
		return domain.LoggingConfig{}, mhcatmongo.MapError(fmt.Errorf("get logging config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

var _ ports.LoggingConfigRepository = (*LoggingConfigRepository)(nil)
var _ ports.LoggingConfigReader = (*LoggingConfigRepository)(nil)
