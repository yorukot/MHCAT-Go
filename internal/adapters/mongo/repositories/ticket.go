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

const TicketConfigCollectionName = "tickets"

type TicketConfigRepository struct {
	collection *drivermongo.Collection
}

func NewTicketConfigRepository(collection *drivermongo.Collection) (*TicketConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo ticket collection is required")
	}
	return &TicketConfigRepository{collection: collection}, nil
}

func NewTicketConfigRepositoryFromDatabase(database *drivermongo.Database) (*TicketConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewTicketConfigRepository(database.Collection(TicketConfigCollectionName))
}

func (r *TicketConfigRepository) GetTicketConfig(ctx context.Context, guildID string) (domain.TicketConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.TicketConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.TicketConfig{}, domain.ErrInvalidTicketConfig
	}
	var document documents.TicketConfigReadDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.TicketConfig{}, ports.ErrTicketConfigNotFound
		}
		return domain.TicketConfig{}, mhcatmongo.MapError(fmt.Errorf("get ticket config: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	return config, ctx.Err()
}

func (r *TicketConfigRepository) CreateTicketConfig(ctx context.Context, config domain.TicketConfig) (ports.TicketConfigCreation, error) {
	if err := ctx.Err(); err != nil {
		return ports.TicketConfigCreation{}, err
	}
	if err := config.ValidateForWrite(); err != nil {
		return ports.TicketConfigCreation{}, err
	}
	document := documents.TicketConfigDocumentFromDomain(config)
	creationID := bson.NewObjectID()
	update, err := ticketConfigCreateUpdate(document, creationID)
	if err != nil {
		return ports.TicketConfigCreation{}, err
	}
	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		mapped := mhcatmongo.MapError(fmt.Errorf("create ticket config: %w", err))
		if mhcatmongo.ErrorIs(mapped, mhcatmongo.ErrorKindConflict) {
			return ports.TicketConfigCreation{}, ports.ErrTicketConfigExists
		}
		return ports.TicketConfigCreation{}, mapped
	}
	if result.MatchedCount > 0 {
		return ports.TicketConfigCreation{}, ports.ErrTicketConfigExists
	}
	return ports.TicketConfigCreation{GuildID: document.Guild, ID: creationID.Hex()}, nil
}

func (r *TicketConfigRepository) RollbackTicketConfigCreation(ctx context.Context, creation ports.TicketConfigCreation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	filter, err := ticketConfigCreationFilter(creation)
	if err != nil {
		return err
	}
	if _, err := r.collection.DeleteOne(ctx, filter); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("rollback ticket config creation: %w", err))
	}
	return nil
}

func (r *TicketConfigRepository) DeleteTicketConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidTicketConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete ticket config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrTicketConfigNotFound
	}
	return ctx.Err()
}

func ticketConfigCreateUpdate(document documents.TicketConfigDocument, creationID bson.ObjectID) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		SetOnInsert("_id", creationID).
		SetOnInsert("guild", document.Guild).
		SetOnInsert("ticket_channel", document.TicketChannel).
		SetOnInsert("admin_id", document.AdminID).
		SetOnInsert("everyone_id", document.EveryoneID).
		Build()
}

func ticketConfigCreationFilter(creation ports.TicketConfigCreation) (bson.D, error) {
	guildID := strings.TrimSpace(creation.GuildID)
	creationID, err := bson.ObjectIDFromHex(strings.TrimSpace(creation.ID))
	if guildID == "" || err != nil {
		return nil, domain.ErrInvalidTicketConfig
	}
	return bson.D{
		{Key: "_id", Value: creationID},
		{Key: "guild", Value: guildID},
	}, nil
}
