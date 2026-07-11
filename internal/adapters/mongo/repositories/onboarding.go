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
	JoinRoleCollectionName     = "join_roles"
	JoinMessageCollectionName  = "join_messages"
	LeaveMessageCollectionName = "leave_messages"
	VerificationCollectionName = "verifications"
	AccountAgeCollectionName   = "create_hours"
)

type JoinRoleConfigRepository struct {
	collection *drivermongo.Collection
}

type JoinMessageConfigRepository struct {
	collection *drivermongo.Collection
}

type LeaveMessageConfigRepository struct {
	collection *drivermongo.Collection
}

type VerificationConfigRepository struct {
	collection *drivermongo.Collection
}

type AccountAgeConfigRepository struct {
	collection *drivermongo.Collection
}

func NewJoinRoleConfigRepository(collection *drivermongo.Collection) (*JoinRoleConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo join role collection is required")
	}
	return &JoinRoleConfigRepository{collection: collection}, nil
}

func NewJoinRoleConfigRepositoryFromDatabase(database *drivermongo.Database) (*JoinRoleConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewJoinRoleConfigRepository(database.Collection(JoinRoleCollectionName))
}

func NewJoinMessageConfigRepository(collection *drivermongo.Collection) (*JoinMessageConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo join message collection is required")
	}
	return &JoinMessageConfigRepository{collection: collection}, nil
}

func NewJoinMessageConfigRepositoryFromDatabase(database *drivermongo.Database) (*JoinMessageConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewJoinMessageConfigRepository(database.Collection(JoinMessageCollectionName))
}

func NewLeaveMessageConfigRepository(collection *drivermongo.Collection) (*LeaveMessageConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo leave message collection is required")
	}
	return &LeaveMessageConfigRepository{collection: collection}, nil
}

func NewLeaveMessageConfigRepositoryFromDatabase(database *drivermongo.Database) (*LeaveMessageConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewLeaveMessageConfigRepository(database.Collection(LeaveMessageCollectionName))
}

func NewVerificationConfigRepository(collection *drivermongo.Collection) (*VerificationConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo verification collection is required")
	}
	return &VerificationConfigRepository{collection: collection}, nil
}

func NewVerificationConfigRepositoryFromDatabase(database *drivermongo.Database) (*VerificationConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewVerificationConfigRepository(database.Collection(VerificationCollectionName))
}

func NewAccountAgeConfigRepository(collection *drivermongo.Collection) (*AccountAgeConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo account age collection is required")
	}
	return &AccountAgeConfigRepository{collection: collection}, nil
}

func NewAccountAgeConfigRepositoryFromDatabase(database *drivermongo.Database) (*AccountAgeConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAccountAgeConfigRepository(database.Collection(AccountAgeCollectionName))
}

func (r *JoinRoleConfigRepository) CreateJoinRoleConfig(ctx context.Context, config domain.JoinRoleConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.JoinRoleDocumentFromDomain(config)
	filter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "role", Value: document.Role}}
	update, err := joinRoleInsertUpdate(document)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("create join role config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ports.ErrJoinRoleConfigExists
	}
	return ctx.Err()
}

func (r *JoinRoleConfigRepository) ListJoinRoleConfigs(ctx context.Context, guildID string) ([]domain.JoinRoleConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidJoinRoleConfig
	}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list join role configs: %w", err))
	}
	defer cursor.Close(ctx)
	var configs []domain.JoinRoleConfig
	for cursor.Next(ctx) {
		var document documents.JoinRoleDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode join role config: %w", err))
		}
		config := document.ToDomain()
		if config.GuildID == "" {
			config.GuildID = guildID
		}
		if err := config.Validate(); err != nil {
			continue
		}
		configs = append(configs, config)
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate join role configs: %w", err))
	}
	for left, right := 0, len(configs)-1; left < right; left, right = left+1, right-1 {
		configs[left], configs[right] = configs[right], configs[left]
	}
	return configs, ctx.Err()
}

func (r *JoinRoleConfigRepository) DeleteJoinRoleConfig(ctx context.Context, guildID string, roleID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	if guildID == "" || roleID == "" {
		return domain.ErrInvalidJoinRoleConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "role", Value: roleID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete join role config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrJoinRoleConfigMissing
	}
	return ctx.Err()
}

func joinRoleInsertUpdate(document documents.JoinRoleDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		SetOnInsert("guild", document.Guild).
		SetOnInsert("role", document.Role).
		SetOnInsert("give_to_who", document.GiveToWho).
		Build()
}

func (r *JoinMessageConfigRepository) GetJoinMessageConfig(ctx context.Context, guildID string) (domain.JoinMessageConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.JoinMessageConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.JoinMessageConfig{}, domain.ErrInvalidJoinMessageConfig
	}
	var document documents.JoinMessageDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.JoinMessageConfig{}, ports.ErrJoinMessageConfigMissing
		}
		return domain.JoinMessageConfig{}, mhcatmongo.MapError(fmt.Errorf("get join message config: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	return config, ctx.Err()
}

func (r *LeaveMessageConfigRepository) PrepareLeaveMessageConfig(ctx context.Context, guildID string, channelID string) (domain.LeaveMessageConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	config := domain.LeaveMessageConfig{GuildID: strings.TrimSpace(guildID), ChannelID: strings.TrimSpace(channelID)}
	if err := config.ValidateChannel(); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	var document documents.LeaveMessageDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: config.GuildID}}).Decode(&document)
	if err != nil && !errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.LeaveMessageConfig{}, mhcatmongo.MapError(fmt.Errorf("load leave message config: %w", err))
	}
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		update, buildErr := leaveMessagePrepareInsertUpdate(config)
		if buildErr != nil {
			return domain.LeaveMessageConfig{}, buildErr
		}
		_, upsertErr := r.collection.UpdateOne(
			ctx,
			bson.D{{Key: "guild", Value: config.GuildID}},
			update,
			options.UpdateOne().SetUpsert(true),
		)
		if upsertErr != nil {
			return domain.LeaveMessageConfig{}, mhcatmongo.MapError(fmt.Errorf("prepare leave message config: %w", upsertErr))
		}
		return config, ctx.Err()
	}
	update, buildErr := mhcatmongo.NewUpdate().Set("channel", config.ChannelID).Build()
	if buildErr != nil {
		return domain.LeaveMessageConfig{}, buildErr
	}
	if _, updateErr := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: config.GuildID}}, update); updateErr != nil {
		return domain.LeaveMessageConfig{}, mhcatmongo.MapError(fmt.Errorf("update leave message channel: %w", updateErr))
	}
	loaded := document.ToDomain()
	loaded.ChannelID = config.ChannelID
	return loaded, ctx.Err()
}

func (r *LeaveMessageConfigRepository) GetLeaveMessageConfig(ctx context.Context, guildID string) (domain.LeaveMessageConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.LeaveMessageConfig{}, domain.ErrInvalidLeaveMessageConfig
	}
	var document documents.LeaveMessageDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.LeaveMessageConfig{}, ports.ErrLeaveMessageConfigMissing
	}
	if err != nil {
		return domain.LeaveMessageConfig{}, mhcatmongo.MapError(fmt.Errorf("load leave message config: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	return config, ctx.Err()
}

func (r *LeaveMessageConfigRepository) SaveLeaveMessageContent(ctx context.Context, config domain.LeaveMessageConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	if err := config.ValidateContent(); err != nil {
		return err
	}
	document := documents.LeaveMessageDocumentFromDomain(config)
	update, err := leaveMessageContentUpdate(document)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save leave message content: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrLeaveMessageConfigMissing
	}
	return ctx.Err()
}

func leaveMessagePrepareInsertUpdate(config domain.LeaveMessageConfig) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("channel", config.ChannelID).
		SetOnInsert("guild", config.GuildID).
		SetOnInsert("message_content", nil).
		SetOnInsert("title", nil).
		SetOnInsert("color", nil).
		Build()
}

func leaveMessageContentUpdate(document documents.LeaveMessageDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("message_content", stringValueOrNil(document.MessageContent)).
		Set("title", stringValueOrNil(document.Title)).
		Set("color", stringValueOrNil(document.Color)).
		Build()
}

func stringValueOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func (r *VerificationConfigRepository) SaveVerificationConfig(ctx context.Context, config domain.VerificationConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.collection == nil {
		return mhcatmongo.MapError(errors.New("mongo verification repository is not configured"))
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.RoleID = strings.TrimSpace(config.RoleID)
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.VerificationDocumentFromDomain(config)
	update, err := verificationConfigUpdate(document)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save verification config: %w", err))
	}
	if result.MatchedCount == 0 {
		insert, buildErr := verificationConfigInsertUpdate(document)
		if buildErr != nil {
			return buildErr
		}
		if _, upsertErr := r.collection.UpdateOne(
			ctx,
			bson.D{{Key: "guild", Value: document.Guild}},
			insert,
			options.UpdateOne().SetUpsert(true),
		); upsertErr != nil {
			return mhcatmongo.MapError(fmt.Errorf("insert verification config: %w", upsertErr))
		}
	}
	return ctx.Err()
}

func (r *VerificationConfigRepository) GetVerificationConfig(ctx context.Context, guildID string) (domain.VerificationConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.VerificationConfig{}, err
	}
	if r == nil || r.collection == nil {
		return domain.VerificationConfig{}, mhcatmongo.MapError(errors.New("mongo verification repository is not configured"))
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.VerificationConfig{}, domain.ErrInvalidVerificationConfig
	}
	var document documents.VerificationDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.VerificationConfig{}, ports.ErrVerificationConfigMissing
		}
		return domain.VerificationConfig{}, mhcatmongo.MapError(fmt.Errorf("get verification config: %w", err))
	}
	config := document.ToDomain()
	if err := config.Validate(); err != nil {
		return domain.VerificationConfig{}, err
	}
	return config, nil
}

func verificationConfigUpdate(document documents.VerificationDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("role", document.Role).
		Set("name", stringValueOrNil(document.Name)).
		Build()
}

func verificationConfigInsertUpdate(document documents.VerificationDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		SetOnInsert("guild", document.Guild).
		SetOnInsert("role", document.Role).
		SetOnInsert("name", stringValueOrNil(document.Name)).
		Build()
}

func (r *AccountAgeConfigRepository) SaveAccountAgeRequirement(ctx context.Context, guildID string, requiredSeconds int64) (domain.AccountAgeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	config := domain.AccountAgeConfig{GuildID: guildID, RequiredSeconds: float64(requiredSeconds)}
	if err := config.Validate(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	var existing documents.AccountAgeReadDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&existing); err != nil {
		if !errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.AccountAgeConfig{}, mhcatmongo.MapError(fmt.Errorf("load existing account age config before save: %w", err))
		}
	} else {
		config.ChannelID = existing.ChannelID()
	}
	document := documents.AccountAgeDocumentFromDomain(config)
	update, err := mhcatmongo.NewUpdate().
		Set("guild", document.Guild).
		Set("hours", document.Hours).
		Set("channel", stringValueOrNil(document.Channel)).
		Build()
	if err != nil {
		return domain.AccountAgeConfig{}, err
	}
	if _, err := r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return domain.AccountAgeConfig{}, mhcatmongo.MapError(fmt.Errorf("save account age config: %w", err))
	}
	return config, ctx.Err()
}

func (r *AccountAgeConfigRepository) SetAccountAgeLogChannel(ctx context.Context, guildID string, channelID string) (domain.AccountAgeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	config, err := r.GetAccountAgeConfig(ctx, guildID)
	if err != nil {
		return domain.AccountAgeConfig{}, err
	}
	update, err := mhcatmongo.NewUpdate().Set("channel", channelID).Build()
	if err != nil {
		return domain.AccountAgeConfig{}, err
	}
	result, err := r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: guildID}}, update)
	if err != nil {
		return domain.AccountAgeConfig{}, mhcatmongo.MapError(fmt.Errorf("set account age log channel: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.AccountAgeConfig{}, ports.ErrAccountAgeConfigMissing
	}
	config.ChannelID = channelID
	return config, ctx.Err()
}

func (r *AccountAgeConfigRepository) DeleteAccountAgeConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidAccountAgeConfig
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete account age config: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrAccountAgeConfigMissing
	}
	return ctx.Err()
}

func (r *AccountAgeConfigRepository) DeleteAccountAgeLogChannel(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidAccountAgeConfig
	}
	update, err := mhcatmongo.NewUpdate().Set("channel", nil).Build()
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: guildID}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete account age log channel: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrAccountAgeConfigMissing
	}
	return ctx.Err()
}

func (r *AccountAgeConfigRepository) GetAccountAgeConfig(ctx context.Context, guildID string) (domain.AccountAgeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	var document documents.AccountAgeReadDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.AccountAgeConfig{}, ports.ErrAccountAgeConfigMissing
		}
		return domain.AccountAgeConfig{}, mhcatmongo.MapError(fmt.Errorf("get account age config: %w", err))
	}
	config, err := document.ToDomain()
	if err != nil {
		return domain.AccountAgeConfig{}, err
	}
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	return config, ctx.Err()
}

var _ ports.JoinRoleConfigRepository = (*JoinRoleConfigRepository)(nil)
var _ ports.JoinMessageConfigReader = (*JoinMessageConfigRepository)(nil)
var _ ports.LeaveMessageConfigRepository = (*LeaveMessageConfigRepository)(nil)
var _ ports.VerificationConfigRepository = (*VerificationConfigRepository)(nil)
var _ ports.AccountAgeConfigRepository = (*AccountAgeConfigRepository)(nil)
