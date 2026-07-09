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
	var document documents.TicketConfigDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.TicketConfig{}, ports.ErrTicketConfigNotFound
		}
		return domain.TicketConfig{}, mhcatmongo.MapError(fmt.Errorf("get ticket config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *TicketConfigRepository) SaveTicketConfig(ctx context.Context, config domain.TicketConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.ValidateForWrite(); err != nil {
		return err
	}
	document := documents.TicketConfigDocumentFromDomain(config)
	update, err := mhcatmongo.NewUpdate().
		Set("ticket_channel", document.TicketChannel).
		Set("admin_id", document.AdminID).
		Set("everyone_id", document.EveryoneID).
		SetOnInsert("guild", document.Guild).
		Build()
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save ticket config: %w", err))
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
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete ticket config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrTicketConfigNotFound
	}
	return ctx.Err()
}
