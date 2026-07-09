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
	return document.ToDomain().Normalize(), ctx.Err()
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
	return document.ToDomain().Normalize(), ctx.Err()
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
