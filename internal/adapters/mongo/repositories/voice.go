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

const VoiceRoomConfigCollectionName = "voice_channels"

type VoiceRoomConfigRepository struct {
	collection *drivermongo.Collection
}

func NewVoiceRoomConfigRepository(collection *drivermongo.Collection) (*VoiceRoomConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo voice room config collection is required")
	}
	return &VoiceRoomConfigRepository{collection: collection}, nil
}

func NewVoiceRoomConfigRepositoryFromDatabase(database *drivermongo.Database) (*VoiceRoomConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewVoiceRoomConfigRepository(database.Collection(VoiceRoomConfigCollectionName))
}

func (r *VoiceRoomConfigRepository) SaveVoiceRoomConfig(ctx context.Context, config domain.VoiceRoomConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.VoiceRoomConfigDocumentFromDomain(config)
	filter := bson.D{
		{Key: "guild", Value: document.Guild},
		{Key: "ticket_channel", Value: document.TicketChannel},
	}
	update, err := voiceRoomConfigUpdate(document, false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save voice room config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := voiceRoomConfigUpdate(document, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, filter, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert voice room config: %w", err))
	}
	return ctx.Err()
}

func (r *VoiceRoomConfigRepository) DeleteVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	triggerChannelID = strings.TrimSpace(triggerChannelID)
	if guildID == "" || triggerChannelID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{
		{Key: "guild", Value: guildID},
		{Key: "ticket_channel", Value: triggerChannelID},
	})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete voice room config by trigger: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrVoiceRoomConfigMissing
	}
	return ctx.Err()
}

func (r *VoiceRoomConfigRepository) DeleteVoiceRoomConfigsByParent(ctx context.Context, guildID string, parentID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	parentID = strings.TrimSpace(parentID)
	if guildID == "" || parentID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{
		{Key: "guild", Value: guildID},
		{Key: "parent", Value: parentID},
	})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete voice room configs by parent: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrVoiceRoomConfigMissing
	}
	return ctx.Err()
}

func voiceRoomConfigUpdate(document documents.VoiceRoomConfigDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().
		Set("limit", document.Limit).
		Set("lock", document.Lock).
		Set("name", document.Name).
		Set("parent", nullableTrimmedString(document.Parent))
	if upsert {
		builder.SetOnInsert("guild", document.Guild).
			SetOnInsert("ticket_channel", document.TicketChannel)
	}
	return builder.Build()
}

var _ ports.VoiceRoomConfigRepository = (*VoiceRoomConfigRepository)(nil)
