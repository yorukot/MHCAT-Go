package repositories

import (
	"context"
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
	TextXPCollectionName  = "text_xps"
	VoiceXPCollectionName = "voice_xps"
)

type EconomyProfileRepository struct {
	coins       *drivermongo.Collection
	giftChanges *drivermongo.Collection
	workSets    *drivermongo.Collection
	workUsers   *drivermongo.Collection
	textXPs     *drivermongo.Collection
	voiceXPs    *drivermongo.Collection
}

func NewEconomyProfileRepository(coins *drivermongo.Collection, giftChanges *drivermongo.Collection, workSets *drivermongo.Collection, workUsers *drivermongo.Collection, textXPs *drivermongo.Collection, voiceXPs *drivermongo.Collection) (*EconomyProfileRepository, error) {
	if coins == nil {
		return nil, fmt.Errorf("coins collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	if workSets == nil {
		return nil, fmt.Errorf("work_sets collection is required")
	}
	if workUsers == nil {
		return nil, fmt.Errorf("work_users collection is required")
	}
	if textXPs == nil {
		return nil, fmt.Errorf("text_xps collection is required")
	}
	if voiceXPs == nil {
		return nil, fmt.Errorf("voice_xps collection is required")
	}
	return &EconomyProfileRepository{
		coins:       coins,
		giftChanges: giftChanges,
		workSets:    workSets,
		workUsers:   workUsers,
		textXPs:     textXPs,
		voiceXPs:    voiceXPs,
	}, nil
}

func NewEconomyProfileRepositoryFromDatabase(database *drivermongo.Database) (*EconomyProfileRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewEconomyProfileRepository(
		database.Collection(CoinCollectionName),
		database.Collection(GiftChangeCollectionName),
		database.Collection(WorkSetCollectionName),
		database.Collection(WorkUserCollectionName),
		database.Collection(TextXPCollectionName),
		database.Collection(VoiceXPCollectionName),
	)
}

func (r *EconomyProfileRepository) GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.CoinBalance{}, domain.ErrInvalidEconomyProfileQuery
	}
	var document documents.CoinDocument
	err := r.coins.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.CoinBalance{}, ports.ErrCoinBalanceNotFound
		}
		return domain.CoinBalance{}, mhcatmongo.MapError(fmt.Errorf("get profile coin balance: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyProfileRepository) GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomyProfileQuery
	}
	var document documents.GiftChangeDocument
	err := r.giftChanges.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.EconomyConfig{GuildID: guildID}, ports.ErrEconomyConfigMissing
		}
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("get profile economy config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyProfileRepository) ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidEconomyProfileQuery
	}
	cursor, err := r.coins.Find(ctx, bson.D{{Key: "guild", Value: guildID}}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list profile coin balances: %w", err))
	}
	defer cursor.Close(ctx)
	balances := []domain.CoinBalance{}
	for cursor.Next(ctx) {
		var document documents.CoinDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode profile coin balance: %w", err))
		}
		balances = append(balances, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate profile coin balances: %w", err))
	}
	return balances, ctx.Err()
}

func (r *EconomyProfileRepository) GetWorkConfig(ctx context.Context, guildID string) (domain.WorkConfig, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.WorkConfig{}, domain.ErrInvalidEconomyProfileQuery
	}
	var document documents.WorkConfigDocument
	err := r.workSets.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkConfig{}, ports.ErrWorkConfigMissing
		}
		return domain.WorkConfig{}, mhcatmongo.MapError(fmt.Errorf("get profile work config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyProfileRepository) GetWorkUser(ctx context.Context, guildID string, userID string) (domain.WorkUserState, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WorkUserState{}, domain.ErrInvalidEconomyProfileQuery
	}
	var document documents.WorkUserDocument
	err := r.workUsers.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "user", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkUserState{}, ports.ErrWorkUserMissing
		}
		return domain.WorkUserState{}, mhcatmongo.MapError(fmt.Errorf("get profile work user: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyProfileRepository) GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return getXPProfile(ctx, r.textXPs, guildID, userID, ports.ErrTextXPProfileMissing, "text")
}

func (r *EconomyProfileRepository) ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return listXPProfiles(ctx, r.textXPs, guildID, "text")
}

func (r *EconomyProfileRepository) GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return getXPProfile(ctx, r.voiceXPs, guildID, userID, ports.ErrVoiceXPProfileMissing, "voice")
}

func (r *EconomyProfileRepository) ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return listXPProfiles(ctx, r.voiceXPs, guildID, "voice")
}

func getXPProfile(ctx context.Context, collection *drivermongo.Collection, guildID string, userID string, missing error, label string) (domain.XPProfile, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.XPProfile{}, domain.ErrInvalidEconomyProfileQuery
	}
	var document documents.XPProfileDocument
	err := collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.XPProfile{}, missing
		}
		return domain.XPProfile{}, mhcatmongo.MapError(fmt.Errorf("get profile %s xp: %w", label, err))
	}
	return document.ToDomain(), ctx.Err()
}

func listXPProfiles(ctx context.Context, collection *drivermongo.Collection, guildID string, label string) ([]domain.XPProfile, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidEconomyProfileQuery
	}
	cursor, err := collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list profile %s xp: %w", label, err))
	}
	defer cursor.Close(ctx)
	profiles := []domain.XPProfile{}
	for cursor.Next(ctx) {
		var document documents.XPProfileDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode profile %s xp: %w", label, err))
		}
		profiles = append(profiles, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate profile %s xp: %w", label, err))
	}
	return profiles, ctx.Err()
}

var _ ports.EconomyProfileRepository = (*EconomyProfileRepository)(nil)
