package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const SchedulerLocksCollectionName = "mhcat_scheduler_locks"

type SchedulerLeaseStore struct {
	collection *drivermongo.Collection
}

type schedulerLeaseDocument struct {
	ID        string    `bson:"_id"`
	LockName  string    `bson:"lock_name"`
	Owner     string    `bson:"owner"`
	Fence     int64     `bson:"fence"`
	ExpiresAt time.Time `bson:"expires_at"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func NewSchedulerLeaseStore(collection *drivermongo.Collection) (*SchedulerLeaseStore, error) {
	if collection == nil {
		return nil, errors.New("scheduler locks collection is required")
	}
	return &SchedulerLeaseStore{collection: collection}, nil
}

func NewSchedulerLeaseStoreFromDatabase(database *drivermongo.Database) (*SchedulerLeaseStore, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewSchedulerLeaseStore(database.Collection(SchedulerLocksCollectionName))
}

func (s *SchedulerLeaseStore) Inspect(ctx context.Context, name string, now time.Time) (domain.SchedulerLeaseStatus, error) {
	name = strings.TrimSpace(name)
	if name == "" || now.IsZero() {
		return domain.SchedulerLeaseStatus{}, domain.ErrInvalidSchedulerLease
	}
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLeaseStatus{}, err
	}
	var document schedulerLeaseDocument
	err := s.collection.FindOne(ctx, bson.D{{Key: "_id", Value: name}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.SchedulerLeaseStatus{Name: name, Held: false}, nil
		}
		return domain.SchedulerLeaseStatus{}, MapError(fmt.Errorf("inspect scheduler lease %s: %w", name, err))
	}
	return document.toStatus(now), ctx.Err()
}

func (s *SchedulerLeaseStore) TryAcquire(ctx context.Context, request domain.SchedulerLeaseRequest) (domain.SchedulerLease, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Owner = strings.TrimSpace(request.Owner)
	if err := request.Validate(); err != nil {
		return domain.SchedulerLease{}, err
	}
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	update := schedulerLeaseAcquireUpdate(request)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var document schedulerLeaseDocument
	err := s.collection.FindOneAndUpdate(ctx, schedulerLeaseAcquireFilter(request), update, opts).Decode(&document)
	if err == nil {
		return document.toDomain(true), ctx.Err()
	}
	if err != drivermongo.ErrNoDocuments {
		return domain.SchedulerLease{}, MapError(fmt.Errorf("acquire scheduler lease %s: %w", request.Name, err))
	}
	document = newSchedulerLeaseDocument(request)
	if _, err := s.collection.InsertOne(ctx, document); err != nil {
		if drivermongo.IsDuplicateKeyError(err) {
			return domain.SchedulerLease{Name: request.Name, Owner: request.Owner, Acquired: false}, nil
		}
		return domain.SchedulerLease{}, MapError(fmt.Errorf("insert scheduler lease %s: %w", request.Name, err))
	}
	return document.toDomain(true), ctx.Err()
}

func (s *SchedulerLeaseStore) Renew(ctx context.Context, lease domain.SchedulerLease, ttl time.Duration, now time.Time) (domain.SchedulerLease, error) {
	lease.Name = strings.TrimSpace(lease.Name)
	lease.Owner = strings.TrimSpace(lease.Owner)
	if err := lease.ValidateHeld(); err != nil {
		return domain.SchedulerLease{}, err
	}
	if ttl <= 0 || now.IsZero() {
		return domain.SchedulerLease{}, domain.ErrInvalidSchedulerLease
	}
	if err := ctx.Err(); err != nil {
		return domain.SchedulerLease{}, err
	}
	filter := schedulerLeaseHeldFilter(lease, now)
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "expires_at", Value: now.Add(ttl)}, {Key: "updated_at", Value: now}}}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var document schedulerLeaseDocument
	if err := s.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.SchedulerLease{}, domain.ErrSchedulerLeaseNotHeld
		}
		return domain.SchedulerLease{}, MapError(fmt.Errorf("renew scheduler lease %s: %w", lease.Name, err))
	}
	return document.toDomain(true), ctx.Err()
}

func (s *SchedulerLeaseStore) Release(ctx context.Context, lease domain.SchedulerLease) error {
	lease.Name = strings.TrimSpace(lease.Name)
	lease.Owner = strings.TrimSpace(lease.Owner)
	if err := lease.ValidateHeld(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	result, err := s.collection.UpdateOne(
		ctx,
		schedulerLeaseIdentityFilter(lease),
		bson.D{{Key: "$set", Value: bson.D{{Key: "owner", Value: ""}, {Key: "expires_at", Value: now}, {Key: "updated_at", Value: now}}}},
	)
	if err != nil {
		return MapError(fmt.Errorf("release scheduler lease %s: %w", lease.Name, err))
	}
	if result.MatchedCount == 0 {
		return domain.ErrSchedulerLeaseNotHeld
	}
	return ctx.Err()
}

func schedulerLeaseAcquireFilter(request domain.SchedulerLeaseRequest) bson.D {
	return bson.D{
		{Key: "_id", Value: request.Name},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "owner", Value: request.Owner}},
			bson.D{{Key: "expires_at", Value: bson.D{{Key: "$lte", Value: request.Now}}}},
		}},
	}
}

func schedulerLeaseAcquireUpdate(request domain.SchedulerLeaseRequest) bson.D {
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "lock_name", Value: request.Name},
			{Key: "owner", Value: request.Owner},
			{Key: "expires_at", Value: request.Now.Add(request.TTL)},
			{Key: "updated_at", Value: request.Now},
		}},
		{Key: "$inc", Value: bson.D{{Key: "fence", Value: int64(1)}}},
	}
}

func schedulerLeaseHeldFilter(lease domain.SchedulerLease, now time.Time) bson.D {
	filter := schedulerLeaseIdentityFilter(lease)
	filter = append(filter, bson.E{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}})
	return filter
}

func schedulerLeaseIdentityFilter(lease domain.SchedulerLease) bson.D {
	return bson.D{{Key: "_id", Value: lease.Name}, {Key: "owner", Value: lease.Owner}, {Key: "fence", Value: lease.Fence}}
}

func newSchedulerLeaseDocument(request domain.SchedulerLeaseRequest) schedulerLeaseDocument {
	return schedulerLeaseDocument{
		ID:        request.Name,
		LockName:  request.Name,
		Owner:     request.Owner,
		Fence:     1,
		ExpiresAt: request.Now.Add(request.TTL),
		CreatedAt: request.Now,
		UpdatedAt: request.Now,
	}
}

func (d schedulerLeaseDocument) toDomain(acquired bool) domain.SchedulerLease {
	name := strings.TrimSpace(d.LockName)
	if name == "" {
		name = d.ID
	}
	return domain.SchedulerLease{
		Name:      name,
		Owner:     d.Owner,
		Fence:     d.Fence,
		Acquired:  acquired,
		ExpiresAt: d.ExpiresAt,
	}
}

func (d schedulerLeaseDocument) toStatus(now time.Time) domain.SchedulerLeaseStatus {
	name := strings.TrimSpace(d.LockName)
	if name == "" {
		name = d.ID
	}
	return domain.SchedulerLeaseStatus{
		Name:      name,
		Owner:     d.Owner,
		Fence:     d.Fence,
		ExpiresAt: d.ExpiresAt,
		Held:      strings.TrimSpace(d.Owner) != "" && d.ExpiresAt.After(now),
	}
}
