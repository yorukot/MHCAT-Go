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

func (r *TicketConfigRepository) CreateTicketConfig(ctx context.Context, config domain.TicketConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.ValidateForWrite(); err != nil {
		return err
	}
	document := documents.TicketConfigDocumentFromDomain(config)
	update, err := ticketConfigCreateUpdate(document)
	if err != nil {
		return err
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
			return ports.ErrTicketConfigExists
		}
		return mapped
	}
	if result.MatchedCount > 0 {
		return ports.ErrTicketConfigExists
	}
	return ctx.Err()
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

func ticketConfigCreateUpdate(document documents.TicketConfigDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		SetOnInsert("guild", document.Guild).
		SetOnInsert("ticket_channel", document.TicketChannel).
		SetOnInsert("admin_id", document.AdminID).
		SetOnInsert("everyone_id", document.EveryoneID).
		Build()
}
