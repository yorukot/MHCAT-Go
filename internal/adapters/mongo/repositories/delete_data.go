package repositories

import (
	"context"
	"errors"
	"fmt"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

const StatsConfigCollectionName = "numbers"

type DeleteDataRepository struct {
	collections map[domain.DeleteDataTarget]*drivermongo.Collection
}

func NewDeleteDataRepositoryFromDatabase(database *drivermongo.Database) (*DeleteDataRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	collections := map[domain.DeleteDataTarget]*drivermongo.Collection{}
	for _, target := range domain.LegacyDeleteDataTargets() {
		name, ok := DeleteDataCollectionName(target)
		if !ok {
			return nil, fmt.Errorf("missing delete data collection for %s", target)
		}
		collections[target] = database.Collection(name)
	}
	return NewDeleteDataRepository(collections)
}

func NewDeleteDataRepository(collections map[domain.DeleteDataTarget]*drivermongo.Collection) (*DeleteDataRepository, error) {
	if len(collections) == 0 {
		return nil, errors.New("mongo delete-data collections are required")
	}
	for _, target := range domain.LegacyDeleteDataTargets() {
		if collections[target] == nil {
			return nil, fmt.Errorf("mongo delete-data collection is required for %s", target)
		}
	}
	return &DeleteDataRepository{collections: collections}, nil
}

func DeleteDataCollectionName(target domain.DeleteDataTarget) (string, bool) {
	switch target {
	case domain.DeleteDataTargetJoinMessage:
		return JoinMessageCollectionName, true
	case domain.DeleteDataTargetLeaveMessage:
		return LeaveMessageCollectionName, true
	case domain.DeleteDataTargetLogging:
		return LoggingConfigCollectionName, true
	case domain.DeleteDataTargetStats:
		return StatsConfigCollectionName, true
	case domain.DeleteDataTargetAutoChat:
		return AutoChatConfigCollectionName, true
	case domain.DeleteDataTargetVerification:
		return VerificationCollectionName, true
	case domain.DeleteDataTargetTextXP:
		return TextXPChannelCollectionName, true
	case domain.DeleteDataTargetVoiceXP:
		return VoiceXPChannelCollectionName, true
	case domain.DeleteDataTargetTicket:
		return TicketConfigCollectionName, true
	default:
		return "", false
	}
}

func (r *DeleteDataRepository) DeleteGuildConfig(ctx context.Context, request domain.DeleteDataRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	request = request.Normalize()
	if err := request.Validate(); err != nil {
		return err
	}
	collection := r.collections[request.Target]
	if collection == nil {
		return domain.ErrInvalidDeleteDataRequest
	}
	result, err := collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: request.GuildID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete legacy config data: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrDeleteDataTargetMissing
	}
	return ctx.Err()
}

var _ ports.DeleteDataRepository = (*DeleteDataRepository)(nil)
