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
	filter := roleReactionConfigFilter(document.Guild, document.Message, document.React)
	update, err := roleReactionConfigUpdate(document, true)
	if err != nil {
		return err
	}
	_, err = r.reactions.UpdateMany(ctx, filter, update, options.UpdateMany().SetUpsert(true))
	if err != nil {
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
	result, err := r.reactions.DeleteMany(ctx, roleReactionConfigFilter(guildID, messageID, react))
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
	var document documents.RoleReactionReadDocument
	if err := r.reactions.FindOne(ctx, roleReactionConfigFilter(guildID, messageID, react)).Decode(&document); err != nil {
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
	models := make([]drivermongo.WriteModel, 0, len(configs))
	for _, config := range configs {
		config = config.Normalize()
		if err := config.Validate(); err != nil {
			return err
		}
		document := documents.RoleButtonDocumentFromDomain(config)
		filter := roleButtonConfigFilter(document.Guild, document.Number)
		update, err := roleButtonConfigUpdate(document, true)
		if err != nil {
			return err
		}
		models = append(models, drivermongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	if len(models) == 0 {
		return ctx.Err()
	}
	if _, err := r.buttons.BulkWrite(ctx, models); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save role button configs: %w", err))
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
	var document documents.RoleButtonReadDocument
	if err := r.buttons.FindOne(ctx, roleButtonConfigFilter(guildID, number)).Decode(&document); err != nil {
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

func roleReactionConfigFilter(guildID string, messageID string, react string) bson.D {
	return bson.D{{Key: "guild", Value: guildID}, {Key: "message", Value: messageID}, {Key: "react", Value: react}}
}

func roleReactionConfigUpdate(document documents.RoleReactionDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().Set("role", document.Role)
	if upsert {
		builder.SetOnInsert("guild", document.Guild).
			SetOnInsert("message", document.Message).
			SetOnInsert("react", document.React)
	}
	return builder.Build()
}

func roleButtonConfigFilter(guildID string, number string) bson.D {
	return bson.D{{Key: "guild", Value: guildID}, {Key: "number", Value: number}}
}

func roleButtonConfigUpdate(document documents.RoleButtonDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().Set("role", document.Role)
	if upsert {
		builder.SetOnInsert("guild", document.Guild).
			SetOnInsert("number", document.Number)
	}
	return builder.Build()
}

var _ ports.RoleSelectionRepository = (*RoleSelectionRepository)(nil)
