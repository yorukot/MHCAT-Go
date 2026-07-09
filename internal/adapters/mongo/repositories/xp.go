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
	TextXPChannelCollectionName     = "text_xp_channels"
	VoiceXPChannelCollectionName    = "voice_xp_channels"
	TextXPRewardRoleCollectionName  = "chat_roles"
	VoiceXPRewardRoleCollectionName = "voice_roles"
)

type TextXPConfigRepository struct {
	collection *drivermongo.Collection
}

type VoiceXPConfigRepository struct {
	collection *drivermongo.Collection
}

type TextXPRewardRoleRepository struct {
	collection *drivermongo.Collection
}

type VoiceXPRewardRoleRepository struct {
	collection *drivermongo.Collection
}

func NewTextXPConfigRepository(collection *drivermongo.Collection) (*TextXPConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo text xp channel collection is required")
	}
	return &TextXPConfigRepository{collection: collection}, nil
}

func NewTextXPConfigRepositoryFromDatabase(database *drivermongo.Database) (*TextXPConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewTextXPConfigRepository(database.Collection(TextXPChannelCollectionName))
}

func NewVoiceXPConfigRepository(collection *drivermongo.Collection) (*VoiceXPConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo voice xp channel collection is required")
	}
	return &VoiceXPConfigRepository{collection: collection}, nil
}

func NewVoiceXPConfigRepositoryFromDatabase(database *drivermongo.Database) (*VoiceXPConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewVoiceXPConfigRepository(database.Collection(VoiceXPChannelCollectionName))
}

func NewTextXPRewardRoleRepository(collection *drivermongo.Collection) (*TextXPRewardRoleRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo text xp reward role collection is required")
	}
	return &TextXPRewardRoleRepository{collection: collection}, nil
}

func NewTextXPRewardRoleRepositoryFromDatabase(database *drivermongo.Database) (*TextXPRewardRoleRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewTextXPRewardRoleRepository(database.Collection(TextXPRewardRoleCollectionName))
}

func NewVoiceXPRewardRoleRepository(collection *drivermongo.Collection) (*VoiceXPRewardRoleRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo voice xp reward role collection is required")
	}
	return &VoiceXPRewardRoleRepository{collection: collection}, nil
}

func NewVoiceXPRewardRoleRepositoryFromDatabase(database *drivermongo.Database) (*VoiceXPRewardRoleRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewVoiceXPRewardRoleRepository(database.Collection(VoiceXPRewardRoleCollectionName))
}

func (r *TextXPConfigRepository) SaveTextXPConfig(ctx context.Context, config domain.TextXPConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.TextXPChannelDocumentFromDomain(config)
	update, err := xpChannelConfigUpdate(document.Channel, document.Color, document.Message, "", false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save text xp config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := xpChannelConfigUpdate(document.Channel, document.Color, document.Message, document.Guild, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert text xp config: %w", err))
	}
	return ctx.Err()
}

func (r *VoiceXPConfigRepository) SaveVoiceXPConfig(ctx context.Context, config domain.VoiceXPConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.VoiceXPChannelDocumentFromDomain(config)
	update, err := xpChannelConfigUpdate(document.Channel, document.Color, document.Message, "", false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save voice xp config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := xpChannelConfigUpdate(document.Channel, document.Color, document.Message, document.Guild, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert voice xp config: %w", err))
	}
	return ctx.Err()
}

func (r *TextXPConfigRepository) DeleteTextXPConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidTextXPConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete text xp config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrTextXPConfigMissing
	}
	return ctx.Err()
}

func (r *VoiceXPConfigRepository) DeleteVoiceXPConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidVoiceXPConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete voice xp config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrVoiceXPConfigMissing
	}
	return ctx.Err()
}

func nullableTrimmedString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableRawString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func xpChannelConfigUpdate(channel string, color string, message string, guild string, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().
		Set("channel", channel).
		Set("color", nullableTrimmedString(color)).
		Set("message", nullableRawString(message)).
		Unset("background")
	if upsert {
		builder.SetOnInsert("guild", guild)
	}
	return builder.Build()
}

func (r *TextXPRewardRoleRepository) ListTextXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	return listXPRewardRoles(ctx, r.collection, guildID)
}

func (r *TextXPRewardRoleRepository) SaveTextXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error {
	return saveXPRewardRole(ctx, r.collection, config)
}

func (r *TextXPRewardRoleRepository) DeleteTextXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error {
	return deleteXPRewardRole(ctx, r.collection, guildID, level, roleID, ports.ErrTextXPRewardRoleMissing)
}

func (r *VoiceXPRewardRoleRepository) ListVoiceXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	return listXPRewardRoles(ctx, r.collection, guildID)
}

func (r *VoiceXPRewardRoleRepository) SaveVoiceXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error {
	return saveXPRewardRole(ctx, r.collection, config)
}

func (r *VoiceXPRewardRoleRepository) DeleteVoiceXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error {
	return deleteXPRewardRole(ctx, r.collection, guildID, level, roleID, ports.ErrVoiceXPRewardRoleMissing)
}

func listXPRewardRoles(ctx context.Context, collection *drivermongo.Collection, guildID string) ([]domain.XPRewardRoleConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidXPRewardRoleConfig
	}
	cursor, err := collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list xp reward roles: %w", err))
	}
	defer cursor.Close(ctx)
	var configs []domain.XPRewardRoleConfig
	for cursor.Next(ctx) {
		var document documents.XPRewardRoleDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode xp reward role: %w", err))
		}
		configs = append(configs, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate xp reward roles: %w", err))
	}
	return configs, ctx.Err()
}

func saveXPRewardRole(ctx context.Context, collection *drivermongo.Collection, config domain.XPRewardRoleConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.XPRewardRoleDocumentFromDomain(config)
	filter := xpRewardRoleFilter(document.Guild, document.Leavel, document.Role)
	if _, err := collection.DeleteMany(ctx, filter); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("replace xp reward role: %w", err))
	}
	if _, err := collection.InsertOne(ctx, document); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("insert xp reward role: %w", err))
	}
	return ctx.Err()
}

func deleteXPRewardRole(ctx context.Context, collection *drivermongo.Collection, guildID string, level int64, roleID string, missing error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	if guildID == "" || roleID == "" {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	result, err := collection.DeleteMany(ctx, xpRewardRoleFilter(guildID, fmt.Sprintf("%d", level), roleID))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete xp reward role: %w", err))
	}
	if result.DeletedCount == 0 {
		return missing
	}
	return ctx.Err()
}

func xpRewardRoleFilter(guildID string, level string, roleID string) bson.D {
	return bson.D{
		{Key: "guild", Value: strings.TrimSpace(guildID)},
		{Key: "leavel", Value: strings.TrimSpace(level)},
		{Key: "role", Value: strings.TrimSpace(roleID)},
	}
}

var _ ports.TextXPConfigRepository = (*TextXPConfigRepository)(nil)
var _ ports.VoiceXPConfigRepository = (*VoiceXPConfigRepository)(nil)
var _ ports.TextXPRewardRoleRepository = (*TextXPRewardRoleRepository)(nil)
var _ ports.VoiceXPRewardRoleRepository = (*VoiceXPRewardRoleRepository)(nil)
