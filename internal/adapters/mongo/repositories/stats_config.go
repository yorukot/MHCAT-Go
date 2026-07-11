package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const StatsRoleConfigCollectionName = "role_numbers"

type StatsConfigRepository struct {
	numbers     *drivermongo.Collection
	roleNumbers *drivermongo.Collection
}

func NewStatsConfigRepository(numbers *drivermongo.Collection) (*StatsConfigRepository, error) {
	if numbers == nil {
		return nil, errors.New("numbers collection is required")
	}
	return &StatsConfigRepository{numbers: numbers}, nil
}

func NewStatsConfigRepositoryWithRoleNumbers(numbers *drivermongo.Collection, roleNumbers *drivermongo.Collection) (*StatsConfigRepository, error) {
	if numbers == nil {
		return nil, errors.New("numbers collection is required")
	}
	if roleNumbers == nil {
		return nil, errors.New("role_numbers collection is required")
	}
	return &StatsConfigRepository{numbers: numbers, roleNumbers: roleNumbers}, nil
}

func NewStatsConfigRepositoryFromDatabase(database *drivermongo.Database) (*StatsConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewStatsConfigRepositoryWithRoleNumbers(database.Collection(StatsConfigCollectionName), database.Collection(StatsRoleConfigCollectionName))
}

func (r *StatsConfigRepository) GetStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	var document documents.StatsConfigDocument
	err := r.numbers.FindOne(ctx, statsConfigFilter(guildID)).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.StatsConfig{}, ports.ErrStatsConfigMissing
		}
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("get stats config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *StatsConfigRepository) SaveStatsConfig(ctx context.Context, config domain.StatsConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config = config.Normalize()
	if config.GuildID == "" || config.ParentID == "" || config.MemberNumberID == "" || config.UserNumberID == "" || config.BotNumberID == "" {
		return domain.ErrInvalidStatsConfigRequest
	}
	update := statsConfigUpdate(config, true)
	_, err := r.numbers.UpdateOne(ctx, statsConfigFilter(config.GuildID), update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save stats config: %w", err))
	}
	return ctx.Err()
}

func (r *StatsConfigRepository) AddStatsConfigChannel(ctx context.Context, guildID string, option string, channelID string, currentValue int) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	idField, nameField, ok := statsOptionalFields(option)
	if guildID == "" || channelID == "" || !ok {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: idField, Value: channelID},
		{Key: nameField, Value: strconv.Itoa(currentValue)},
	}}}
	result, err := r.numbers.UpdateOne(ctx, statsConfigFilter(guildID), update)
	if err != nil {
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("add stats config channel: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	return r.GetStatsConfig(ctx, guildID)
}

func (r *StatsConfigRepository) DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	filter := statsConfigFilter(guildID)
	var document documents.StatsConfigDocument
	err := r.numbers.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.StatsConfig{}, ports.ErrStatsConfigMissing
		}
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("find stats config: %w", err))
	}
	result, err := r.numbers.DeleteMany(ctx, filter)
	if err != nil {
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("delete stats config: %w", err))
	}
	if result.DeletedCount == 0 {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	return document.ToDomain(), ctx.Err()
}

func (r *StatsConfigRepository) ListStatsConfigs(ctx context.Context) ([]domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cursor, err := r.numbers.Find(ctx, bson.D{})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list stats configs: %w", err))
	}
	defer cursor.Close(ctx)
	var configs []domain.StatsConfig
	for cursor.Next(ctx) {
		var document documents.StatsConfigDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode stats config: %w", err))
		}
		configs = append(configs, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate stats configs: %w", err))
	}
	return configs, ctx.Err()
}

func (r *StatsConfigRepository) UpdateStatsConfigCounters(ctx context.Context, guildID string, update domain.StatsConfigCounterUpdate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidStatsConfigRequest
	}
	if update.IsZero() {
		return ctx.Err()
	}
	set := bson.D{}
	appendCounter := func(key string, value *string) {
		if value != nil {
			set = append(set, bson.E{Key: key, Value: strings.TrimSpace(*value)})
		}
	}
	appendCounter("memberNumber_name", update.MemberNumberName)
	appendCounter("userNumber_name", update.UserNumberName)
	appendCounter("BotNumber_name", update.BotNumberName)
	appendCounter("channelnumber_name", update.ChannelNumberName)
	appendCounter("textnumber_name", update.TextNumberName)
	appendCounter("voicenumber_name", update.VoiceNumberName)
	result, err := r.numbers.UpdateOne(ctx, statsConfigFilter(guildID), bson.D{{Key: "$set", Value: set}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("update stats config counters: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrStatsConfigMissing
	}
	return ctx.Err()
}

func (r *StatsConfigRepository) SaveStatsRoleConfig(ctx context.Context, config domain.StatsRoleConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config = config.Normalize()
	if config.GuildID == "" || config.ChannelID == "" || config.ChannelName == "" || config.RoleID == "" {
		return domain.ErrInvalidStatsConfigRequest
	}
	if r.roleNumbers == nil {
		return errors.New("role_numbers collection is required")
	}
	filter := statsRoleConfigFilter(config.GuildID, config.RoleID)
	if _, err := r.roleNumbers.DeleteMany(ctx, filter); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("replace stats role config: %w", err))
	}
	if _, err := r.roleNumbers.InsertOne(ctx, documents.StatsRoleConfigDocumentFromDomain(config)); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save stats role config: %w", err))
	}
	return ctx.Err()
}

func (r *StatsConfigRepository) ListStatsRoleConfigs(ctx context.Context) ([]domain.StatsRoleConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.roleNumbers == nil {
		return nil, errors.New("role_numbers collection is required")
	}
	cursor, err := r.roleNumbers.Find(ctx, bson.D{})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list stats role configs: %w", err))
	}
	defer cursor.Close(ctx)
	var configs []domain.StatsRoleConfig
	for cursor.Next(ctx) {
		var document documents.StatsRoleConfigDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode stats role config: %w", err))
		}
		configs = append(configs, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate stats role configs: %w", err))
	}
	return configs, ctx.Err()
}

func (r *StatsConfigRepository) UpdateStatsRoleConfigCounter(ctx context.Context, guildID string, roleID string, currentValue string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	currentValue = strings.TrimSpace(currentValue)
	if guildID == "" || roleID == "" || currentValue == "" {
		return domain.ErrInvalidStatsConfigRequest
	}
	if r.roleNumbers == nil {
		return errors.New("role_numbers collection is required")
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "channel_name", Value: currentValue}}}}
	result, err := r.roleNumbers.UpdateOne(ctx, statsRoleConfigFilter(guildID, roleID), update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("update stats role config counter: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrStatsConfigMissing
	}
	return ctx.Err()
}

func statsConfigFilter(guildID string) bson.D {
	return bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}}
}

func statsRoleConfigFilter(guildID string, roleID string) bson.D {
	return bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}, {Key: "role", Value: strings.TrimSpace(roleID)}}
}

func statsConfigUpdate(config domain.StatsConfig, upsert bool) bson.D {
	config = config.Normalize()
	set := bson.D{
		{Key: "parent", Value: config.ParentID},
		{Key: "memberNumber", Value: nullableTrimmedString(config.MemberNumberID)},
		{Key: "memberNumber_name", Value: nullableTrimmedString(config.MemberNumberName)},
		{Key: "userNumber", Value: nullableTrimmedString(config.UserNumberID)},
		{Key: "userNumber_name", Value: nullableTrimmedString(config.UserNumberName)},
		{Key: "BotNumber", Value: nullableTrimmedString(config.BotNumberID)},
		{Key: "BotNumber_name", Value: nullableTrimmedString(config.BotNumberName)},
		{Key: "channelnumber", Value: nullableTrimmedString(config.ChannelNumberID)},
		{Key: "channelnumber_name", Value: nullableTrimmedString(config.ChannelNumberName)},
		{Key: "textnumber", Value: nullableTrimmedString(config.TextNumberID)},
		{Key: "textnumber_name", Value: nullableTrimmedString(config.TextNumberName)},
		{Key: "voicenumber", Value: nullableTrimmedString(config.VoiceNumberID)},
		{Key: "voicenumber_name", Value: nullableTrimmedString(config.VoiceNumberName)},
		{Key: "categoriesnumber", Value: nil},
		{Key: "categoriesnumber_name", Value: nil},
		{Key: "rolesnumber", Value: nil},
		{Key: "rolesnumber_name", Value: nil},
		{Key: "rolenumber", Value: nil},
		{Key: "rolenumber_name", Value: nil},
		{Key: "norolenumber", Value: nil},
		{Key: "norolenumber_name", Value: nil},
		{Key: "emojisnumber", Value: nil},
		{Key: "emojisnumber_name", Value: nil},
		{Key: "staticnumber", Value: nil},
		{Key: "staticnumber_name", Value: nil},
		{Key: "gifnumber", Value: nil},
		{Key: "gifnumber_name", Value: nil},
		{Key: "stickersnumber", Value: nil},
		{Key: "stickersnumber_name", Value: nil},
		{Key: "boostsnumber", Value: nil},
		{Key: "boostsnumber_name", Value: nil},
		{Key: "tier", Value: nil},
		{Key: "tier_name", Value: nil},
		{Key: "onlinenumber", Value: nil},
		{Key: "onlinenumber_name", Value: nil},
		{Key: "dndnumber", Value: nil},
		{Key: "dndnumber_name", Value: nil},
		{Key: "idlenumber", Value: nil},
		{Key: "idlenumber_name", Value: nil},
		{Key: "offlinenumber", Value: nil},
		{Key: "offlinenumber_name", Value: nil},
		{Key: "onlinebotnumber", Value: nil},
		{Key: "onlinebotnumber_name", Value: nil},
		{Key: "statusnumber", Value: nil},
		{Key: "statusnumber_name", Value: nil},
	}
	update := bson.D{{Key: "$set", Value: set}}
	if upsert {
		update = append(update, bson.E{Key: "$setOnInsert", Value: bson.D{{Key: "guild", Value: config.GuildID}}})
	}
	return update
}

func statsOptionalFields(option string) (string, string, bool) {
	switch strings.TrimSpace(option) {
	case domain.StatsOptionChannelCount:
		return "channelnumber", "channelnumber_name", true
	case domain.StatsOptionTextCount:
		return "textnumber", "textnumber_name", true
	case domain.StatsOptionVoiceCount:
		return "voicenumber", "voicenumber_name", true
	default:
		return "", "", false
	}
}

var _ ports.StatsConfigRepository = (*StatsConfigRepository)(nil)
var _ ports.StatsRoleConfigRepository = (*StatsConfigRepository)(nil)
var _ ports.StatsRenameRepository = (*StatsConfigRepository)(nil)
