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

const BalanceCollectionName = "chatgpt_gets"

type BalanceRepository struct {
	collection *drivermongo.Collection
}

func NewBalanceRepository(collection *drivermongo.Collection) (*BalanceRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo balance collection is required")
	}
	return &BalanceRepository{collection: collection}, nil
}

func NewBalanceRepositoryFromDatabase(database *drivermongo.Database) (*BalanceRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewBalanceRepository(database.Collection(BalanceCollectionName))
}

func (r *BalanceRepository) GetBalance(ctx context.Context, guildID string) (domain.Balance, error) {
	if err := ctx.Err(); err != nil {
		return domain.Balance{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.Balance{}, domain.ErrInvalidBalanceQuery
	}
	var document documents.BalanceDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.Balance{}, ports.ErrBalanceMissing
		}
		return domain.Balance{}, mhcatmongo.MapError(fmt.Errorf("get balance: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

var _ ports.BalanceRepository = (*BalanceRepository)(nil)
