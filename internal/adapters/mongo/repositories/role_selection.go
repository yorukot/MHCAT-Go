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
)

const (
	RoleReactionCollectionName = "message_reactions"
	RoleButtonCollectionName   = "btns"
)

type RoleSelectionRepository struct {
	reactions *drivermongo.Collection
	buttons   *drivermongo.Collection
}

func NewRoleSelectionRepository(reactions *drivermongo.Collection, buttons *drivermongo.Collection) (*RoleSelectionRepository, error) {
	if reactions == nil {
		return nil, errors.New("mongo role reaction collection is required")
	}
	if buttons == nil {
		return nil, errors.New("mongo role button collection is required")
	}
	return &RoleSelectionRepository{reactions: reactions, buttons: buttons}, nil
}

func NewRoleSelectionRepositoryFromDatabase(database *drivermongo.Database) (*RoleSelectionRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewRoleSelectionRepository(database.Collection(RoleReactionCollectionName), database.Collection(RoleButtonCollectionName))
}

func (r *RoleSelectionRepository) SaveRoleReactionConfig(ctx context.Context, config domain.RoleReactionConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.RoleReactionDocumentFromDomain(config)
	filter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "message", Value: document.Message}, {Key: "react", Value: document.React}}
	if _, err := r.reactions.DeleteMany(ctx, filter); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("replace role reaction config: %w", err))
	}
	if _, err := r.reactions.InsertOne(ctx, bson.D{
		{Key: "guild", Value: document.Guild},
		{Key: "message", Value: document.Message},
		{Key: "react", Value: document.React},
		{Key: "role", Value: document.Role},
	}); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save role reaction config: %w", err))
	}
	return ctx.Err()
}

func (r *RoleSelectionRepository) DeleteRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	messageID = strings.TrimSpace(messageID)
	react = strings.TrimSpace(react)
	if guildID == "" || messageID == "" || react == "" {
		return domain.ErrInvalidRoleSelectionConfig
	}
	result, err := r.reactions.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "message", Value: messageID}, {Key: "react", Value: react}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete role reaction config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrRoleReactionConfigMissing
	}
	return ctx.Err()
}

func (r *RoleSelectionRepository) GetRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) (domain.RoleReactionConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	messageID = strings.TrimSpace(messageID)
	react = strings.TrimSpace(react)
	if guildID == "" || messageID == "" || react == "" {
		return domain.RoleReactionConfig{}, domain.ErrInvalidRoleSelectionConfig
	}
	var document documents.RoleReactionDocument
	if err := r.reactions.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "message", Value: messageID}, {Key: "react", Value: react}}).Decode(&document); err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.RoleReactionConfig{}, ports.ErrRoleReactionConfigMissing
		}
		return domain.RoleReactionConfig{}, mhcatmongo.MapError(fmt.Errorf("get role reaction config: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	if err := config.Validate(); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	return config, ctx.Err()
}

func (r *RoleSelectionRepository) SaveRoleButtonConfigs(ctx context.Context, configs ...domain.RoleButtonConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, config := range configs {
		config = config.Normalize()
		if err := config.Validate(); err != nil {
			return err
		}
		document := documents.RoleButtonDocumentFromDomain(config)
		filter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "number", Value: document.Number}}
		if _, err := r.buttons.DeleteMany(ctx, filter); err != nil {
			return mhcatmongo.MapError(fmt.Errorf("replace role button config: %w", err))
		}
		if _, err := r.buttons.InsertOne(ctx, bson.D{
			{Key: "guild", Value: document.Guild},
			{Key: "number", Value: document.Number},
			{Key: "role", Value: document.Role},
		}); err != nil {
			return mhcatmongo.MapError(fmt.Errorf("save role button config: %w", err))
		}
	}
	return ctx.Err()
}

func (r *RoleSelectionRepository) GetRoleButtonConfig(ctx context.Context, guildID string, number string) (domain.RoleButtonConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.RoleButtonConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	number = strings.TrimSpace(number)
	if guildID == "" || number == "" {
		return domain.RoleButtonConfig{}, domain.ErrInvalidRoleSelectionConfig
	}
	var document documents.RoleButtonDocument
	if err := r.buttons.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "number", Value: number}}).Decode(&document); err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.RoleButtonConfig{}, ports.ErrRoleButtonConfigMissing
		}
		return domain.RoleButtonConfig{}, mhcatmongo.MapError(fmt.Errorf("get role button config: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	if err := config.Validate(); err != nil {
		return domain.RoleButtonConfig{}, err
	}
	return config, ctx.Err()
}

var _ ports.RoleSelectionRepository = (*RoleSelectionRepository)(nil)
