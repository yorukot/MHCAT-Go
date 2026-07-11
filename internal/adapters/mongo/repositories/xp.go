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

type XPAdminRepository struct {
	textProfiles  *drivermongo.Collection
	voiceProfiles *drivermongo.Collection
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

func NewXPAdminRepository(textProfiles *drivermongo.Collection, voiceProfiles *drivermongo.Collection) (*XPAdminRepository, error) {
	if textProfiles == nil {
		return nil, errors.New("mongo text xp profile collection is required")
	}
	if voiceProfiles == nil {
		return nil, errors.New("mongo voice xp profile collection is required")
	}
	return &XPAdminRepository{textProfiles: textProfiles, voiceProfiles: voiceProfiles}, nil
}

func NewXPAdminRepositoryFromDatabase(database *drivermongo.Database) (*XPAdminRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewXPAdminRepository(database.Collection(TextXPCollectionName), database.Collection(VoiceXPCollectionName))
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

func (r *TextXPConfigRepository) GetTextXPConfig(ctx context.Context, guildID string) (domain.TextXPConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.TextXPConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.TextXPConfig{}, domain.ErrInvalidTextXPConfig
	}
	var document documents.TextXPChannelDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.TextXPConfig{}, ports.ErrTextXPConfigMissing
	}
	if err != nil {
		return domain.TextXPConfig{}, mhcatmongo.MapError(fmt.Errorf("get text xp config: %w", err))
	}
	config := document.ToDomain()
	if strings.TrimSpace(config.GuildID) == "" {
		config.GuildID = guildID
	}
	if err := config.Validate(); err != nil {
		return domain.TextXPConfig{}, err
	}
	return config, ctx.Err()
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

func (r *VoiceXPConfigRepository) GetVoiceXPConfig(ctx context.Context, guildID string) (domain.VoiceXPConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.VoiceXPConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.VoiceXPConfig{}, domain.ErrInvalidVoiceXPConfig
	}
	var document documents.VoiceXPChannelDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.VoiceXPConfig{}, ports.ErrVoiceXPConfigMissing
	}
	if err != nil {
		return domain.VoiceXPConfig{}, mhcatmongo.MapError(fmt.Errorf("get voice xp config: %w", err))
	}
	config := document.ToDomain()
	if strings.TrimSpace(config.GuildID) == "" {
		config.GuildID = guildID
	}
	if err := config.Validate(); err != nil {
		return domain.VoiceXPConfig{}, err
	}
	return config, ctx.Err()
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

func (r *XPAdminRepository) GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return getAdminXPProfile(ctx, r.textProfiles, guildID, userID, ports.ErrTextXPProfileMissing, "text")
}

func (r *XPAdminRepository) SaveTextXPProfile(ctx context.Context, profile domain.XPProfile) error {
	return saveAdminXPProfile(ctx, r.textProfiles, profile, "text", false)
}

func (r *XPAdminRepository) AccrueTextXP(ctx context.Context, guildID string, userID string, gained int64) (domain.XPProfile, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.XPProfile{}, false, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.XPProfile{}, false, domain.ErrInvalidXPAdjustment
	}
	if gained < 0 {
		gained = 0
	}
	var previous documents.XPProfileDocument
	err := r.textProfiles.FindOneAndUpdate(
		ctx,
		xpProfileFilter(guildID, userID),
		textXPAccrualPipeline(guildID, userID, gained),
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.Before),
	).Decode(&previous)
	profile := domain.XPProfile{GuildID: guildID, UserID: userID}
	if err == nil {
		profile = previous.ToDomain().Normalize()
	} else if !errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.XPProfile{}, false, mhcatmongo.MapError(fmt.Errorf("accrue text xp: %w", err))
	}
	profile, leveled := domain.ApplyTextXPMessage(profile, gained)
	return profile, leveled, ctx.Err()
}

func (r *XPAdminRepository) GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return getAdminXPProfile(ctx, r.voiceProfiles, guildID, userID, ports.ErrVoiceXPProfileMissing, "voice")
}

func (r *XPAdminRepository) SaveVoiceXPProfile(ctx context.Context, profile domain.XPProfile) error {
	return saveAdminXPProfile(ctx, r.voiceProfiles, profile, "voice", true)
}

func (r *XPAdminRepository) AccrueVoiceXP(ctx context.Context, guildID string, userID string, gained int64) (domain.XPProfile, bool, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.XPProfile{}, false, false, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.XPProfile{}, false, false, domain.ErrInvalidXPAdjustment
	}
	if gained < 0 {
		gained = 0
	}
	filter := append(xpProfileFilter(guildID, userID), bson.E{Key: "leavejoin", Value: domain.VoiceXPSessionJoined})
	var previous documents.XPProfileDocument
	err := r.voiceProfiles.FindOneAndUpdate(
		ctx,
		filter,
		voiceXPAccrualPipeline(guildID, userID, gained),
		options.FindOneAndUpdate().SetReturnDocument(options.Before),
	).Decode(&previous)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		profile, getErr := r.GetVoiceXPProfile(ctx, guildID, userID)
		if getErr != nil {
			return domain.XPProfile{}, false, false, getErr
		}
		return profile, false, false, ctx.Err()
	}
	if err != nil {
		return domain.XPProfile{}, false, false, mhcatmongo.MapError(fmt.Errorf("accrue voice xp: %w", err))
	}
	profile := previous.ToDomain().Normalize()
	profile, leveled := domain.ApplyVoiceXPTick(profile, gained)
	return profile, true, leveled, ctx.Err()
}

func (r *XPAdminRepository) MarkVoiceXPJoined(ctx context.Context, guildID string, userID string) error {
	return markVoiceXPSession(ctx, r.voiceProfiles, guildID, userID, domain.VoiceXPSessionJoined)
}

func (r *XPAdminRepository) MarkVoiceXPLeft(ctx context.Context, guildID string, userID string) error {
	return markVoiceXPSession(ctx, r.voiceProfiles, guildID, userID, domain.VoiceXPSessionLeft)
}

func (r *XPAdminRepository) ListJoinedVoiceXPSessions(ctx context.Context) ([]domain.XPProfile, error) {
	return listJoinedVoiceXPSessions(ctx, r.voiceProfiles)
}

func (r *XPAdminRepository) ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return listAdminXPProfiles(ctx, r.textProfiles, guildID, "text")
}

func (r *XPAdminRepository) ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return listAdminXPProfiles(ctx, r.voiceProfiles, guildID, "voice")
}

func (r *XPAdminRepository) DeleteTextXPProfile(ctx context.Context, guildID string, userID string) error {
	return deleteAdminXPProfile(ctx, r.textProfiles, guildID, userID, ports.ErrTextXPProfileMissing, "text")
}

func (r *XPAdminRepository) DeleteVoiceXPProfile(ctx context.Context, guildID string, userID string) error {
	return deleteAdminXPProfile(ctx, r.voiceProfiles, guildID, userID, ports.ErrVoiceXPProfileMissing, "voice")
}

func (r *XPAdminRepository) DeleteTextXPGuild(ctx context.Context, guildID string) error {
	return deleteAdminXPGuild(ctx, r.textProfiles, guildID, ports.ErrTextXPProfileMissing, "text")
}

func (r *XPAdminRepository) DeleteVoiceXPGuild(ctx context.Context, guildID string) error {
	return deleteAdminXPGuild(ctx, r.voiceProfiles, guildID, ports.ErrVoiceXPProfileMissing, "voice")
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
		Set("color", nullableRawString(color)).
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

func getAdminXPProfile(ctx context.Context, collection *drivermongo.Collection, guildID string, userID string, missing error, label string) (domain.XPProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.XPProfile{}, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.XPProfile{}, domain.ErrInvalidXPAdjustment
	}
	var document documents.XPProfileDocument
	err := collection.FindOne(ctx, xpProfileFilter(guildID, userID)).Decode(&document)
	if err != nil {
		if errors.Is(err, drivermongo.ErrNoDocuments) {
			return domain.XPProfile{}, missing
		}
		return domain.XPProfile{}, mhcatmongo.MapError(fmt.Errorf("get admin %s xp profile: %w", label, err))
	}
	return document.ToDomain().Normalize(), ctx.Err()
}

func listAdminXPProfiles(ctx context.Context, collection *drivermongo.Collection, guildID string, label string) ([]domain.XPProfile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidXPRankQuery
	}
	cursor, err := collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list admin %s xp profiles: %w", label, err))
	}
	defer cursor.Close(ctx)
	profiles := []domain.XPProfile{}
	for cursor.Next(ctx) {
		var document documents.XPProfileDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode admin %s xp profile: %w", label, err))
		}
		profiles = append(profiles, document.ToDomain().Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate admin %s xp profiles: %w", label, err))
	}
	return profiles, ctx.Err()
}

func saveAdminXPProfile(ctx context.Context, collection *drivermongo.Collection, profile domain.XPProfile, label string, voice bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	profile = profile.Normalize()
	if err := profile.Validate(); err != nil {
		return err
	}
	_, err := collection.UpdateOne(ctx, xpProfileFilter(profile.GuildID, profile.UserID), xpProfileUpdate(profile, voice), options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save admin %s xp profile: %w", label, err))
	}
	return ctx.Err()
}

func markVoiceXPSession(ctx context.Context, collection *drivermongo.Collection, guildID string, userID string, state string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	_, err := collection.UpdateOne(ctx, xpProfileFilter(guildID, userID), voiceXPSessionUpdate(guildID, userID, state), options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("mark voice xp session: %w", err))
	}
	return ctx.Err()
}

func listJoinedVoiceXPSessions(ctx context.Context, collection *drivermongo.Collection) ([]domain.XPProfile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cursor, err := collection.Find(ctx, voiceXPJoinedSessionFilter())
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list joined voice xp sessions: %w", err))
	}
	defer cursor.Close(ctx)
	profiles := []domain.XPProfile{}
	for cursor.Next(ctx) {
		var document documents.XPProfileDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode joined voice xp session: %w", err))
		}
		profile := document.ToDomain().Normalize()
		if profile.GuildID == "" || profile.UserID == "" {
			continue
		}
		profiles = append(profiles, profile)
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate joined voice xp sessions: %w", err))
	}
	return profiles, ctx.Err()
}

func deleteAdminXPProfile(ctx context.Context, collection *drivermongo.Collection, guildID string, userID string, missing error, label string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	result, err := collection.DeleteMany(ctx, xpProfileFilter(guildID, userID))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete admin %s xp profile: %w", label, err))
	}
	if result.DeletedCount == 0 {
		return missing
	}
	return ctx.Err()
}

func deleteAdminXPGuild(ctx context.Context, collection *drivermongo.Collection, guildID string, missing error, label string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	result, err := collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete admin %s xp guild: %w", label, err))
	}
	if result.DeletedCount == 0 {
		return missing
	}
	return ctx.Err()
}

func xpProfileFilter(guildID string, userID string) bson.D {
	return bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}, {Key: "member", Value: strings.TrimSpace(userID)}}
}

func voiceXPJoinedSessionFilter() bson.D {
	return bson.D{{Key: "leavejoin", Value: domain.VoiceXPSessionJoined}}
}

func xpProfileUpdate(profile domain.XPProfile, voice bool) bson.D {
	profile = profile.Normalize()
	setOnInsert := bson.D{
		{Key: "guild", Value: profile.GuildID},
		{Key: "member", Value: profile.UserID},
	}
	if voice {
		setOnInsert = append(setOnInsert, bson.E{Key: "leavejoin", Value: "leave"})
	}
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "xp", Value: fmt.Sprintf("%d", profile.XP)},
			{Key: "leavel", Value: fmt.Sprintf("%d", profile.Level)},
		}},
		{Key: "$setOnInsert", Value: setOnInsert},
	}
}

func textXPAccrualPipeline(guildID string, userID string, gained int64) drivermongo.Pipeline {
	level := legacyXPInt64Expression("$leavel")
	xp := legacyXPInt64Expression("$xp")
	total := bson.D{{Key: "$add", Value: bson.A{xp, gained}}}
	levelDouble := bson.D{{Key: "$toDouble", Value: level}}
	requiredFloat := bson.D{{Key: "$add", Value: bson.A{
		bson.D{{Key: "$multiply", Value: bson.A{
			levelDouble,
			bson.D{{Key: "$divide", Value: bson.A{levelDouble, 3}}},
			100,
		}}},
		100,
	}}}
	required := bson.D{{Key: "$toLong", Value: requiredFloat}}
	leveled := bson.D{{Key: "$gt", Value: bson.A{total, required}}}
	nextXP := bson.D{{Key: "$cond", Value: bson.A{leveled, int64(0), total}}}
	nextLevel := bson.D{{Key: "$cond", Value: bson.A{
		leveled,
		bson.D{{Key: "$add", Value: bson.A{level, int64(1)}}},
		level,
	}}}
	set := bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "xp", Value: bson.D{{Key: "$toString", Value: nextXP}}},
		{Key: "leavel", Value: bson.D{{Key: "$toString", Value: nextLevel}}},
	}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: set}}}
}

func voiceXPAccrualPipeline(guildID string, userID string, gained int64) drivermongo.Pipeline {
	level := legacyXPInt64Expression("$leavel")
	xp := legacyXPInt64Expression("$xp")
	total := bson.D{{Key: "$add", Value: bson.A{xp, gained}}}
	levelDouble := bson.D{{Key: "$toDouble", Value: level}}
	requiredFloat := bson.D{{Key: "$add", Value: bson.A{
		bson.D{{Key: "$multiply", Value: bson.A{
			levelDouble,
			levelDouble,
			0.5,
			100,
		}}},
		100,
	}}}
	required := bson.D{{Key: "$toLong", Value: requiredFloat}}
	leveled := bson.D{{Key: "$gt", Value: bson.A{total, required}}}
	nextXP := bson.D{{Key: "$cond", Value: bson.A{leveled, gained, total}}}
	nextLevel := bson.D{{Key: "$cond", Value: bson.A{
		leveled,
		bson.D{{Key: "$add", Value: bson.A{level, int64(1)}}},
		level,
	}}}
	set := bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "xp", Value: bson.D{{Key: "$toString", Value: nextXP}}},
		{Key: "leavel", Value: bson.D{{Key: "$toString", Value: nextLevel}}},
	}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: set}}}
}

func legacyXPInt64Expression(field string) bson.D {
	fieldType := bson.D{{Key: "$type", Value: field}}
	convert := func(input any) bson.D {
		return bson.D{{Key: "$convert", Value: bson.D{
			{Key: "input", Value: input},
			{Key: "to", Value: "long"},
			{Key: "onError", Value: int64(0)},
			{Key: "onNull", Value: int64(0)},
		}}}
	}
	return bson.D{{Key: "$switch", Value: bson.D{
		{Key: "branches", Value: bson.A{
			bson.D{
				{Key: "case", Value: bson.D{{Key: "$eq", Value: bson.A{fieldType, "string"}}}},
				{Key: "then", Value: convert(bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: field}}}})},
			},
			bson.D{
				{Key: "case", Value: bson.D{{Key: "$in", Value: bson.A{fieldType, bson.A{"int", "long", "double", "decimal"}}}}},
				{Key: "then", Value: convert(field)},
			},
		}},
		{Key: "default", Value: int64(0)},
	}}}
}

func voiceXPSessionUpdate(guildID string, userID string, state string) bson.D {
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "leavejoin", Value: state},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "guild", Value: strings.TrimSpace(guildID)},
			{Key: "member", Value: strings.TrimSpace(userID)},
			{Key: "xp", Value: "0"},
			{Key: "leavel", Value: "0"},
		}},
	}
}

var _ ports.TextXPConfigRepository = (*TextXPConfigRepository)(nil)
var _ ports.VoiceXPConfigRepository = (*VoiceXPConfigRepository)(nil)
var _ ports.VoiceXPConfigReader = (*VoiceXPConfigRepository)(nil)
var _ ports.TextXPRewardRoleRepository = (*TextXPRewardRoleRepository)(nil)
var _ ports.VoiceXPRewardRoleRepository = (*VoiceXPRewardRoleRepository)(nil)
var _ ports.XPAdminRepository = (*XPAdminRepository)(nil)
var _ ports.TextXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.AtomicTextXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.VoiceXPSessionRepository = (*XPAdminRepository)(nil)
var _ ports.VoiceXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.AtomicVoiceXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.XPResetRepository = (*XPAdminRepository)(nil)
var _ ports.XPRankRepository = (*XPAdminRepository)(nil)
