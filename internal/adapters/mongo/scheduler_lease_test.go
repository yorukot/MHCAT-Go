package mongo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSchedulerLeaseMongoIntegrationLifecycle(t *testing.T) {
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	databaseName := os.Getenv("MHCAT_MONGODB_DATABASE")
	if uri == "" || databaseName == "" {
		t.Fatal("MHCAT_MONGODB_URI and MHCAT_MONGODB_DATABASE are required")
	}
	client, err := NewClient(Options{
		URI:            uri,
		Database:       databaseName,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	database, err := client.Database()
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	if store, err := NewSchedulerLeaseStoreFromDatabase(database); err != nil || store == nil {
		t.Fatalf("new store from database: store=%#v err=%v", store, err)
	}
	collection := database.Collection(fmt.Sprintf("mhcat_scheduler_lease_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = collection.Drop(context.Background()) })
	store, err := NewSchedulerLeaseStore(collection)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	base, err := NewBaseRepository(collection)
	if err != nil {
		t.Fatalf("new base repository: %v", err)
	}
	if base.CollectionName() != collection.Name() {
		t.Fatalf("collection name = %q", base.CollectionName())
	}
	if err := base.Ping(ctx); err != nil {
		t.Fatalf("base repository ping: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	status, err := store.Inspect(ctx, " daily-reset ", now)
	if err != nil || status.Name != "daily-reset" || status.Held {
		t.Fatalf("missing status=%#v err=%v", status, err)
	}
	first, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: " daily-reset ", Owner: " worker-a ", TTL: time.Minute, Now: now})
	if err != nil || !first.Acquired || first.Fence != 1 || first.Owner != "worker-a" {
		t.Fatalf("first lease=%#v err=%v", first, err)
	}
	blocked, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-b", TTL: time.Minute, Now: now.Add(time.Second)})
	if err != nil || blocked.Acquired {
		t.Fatalf("blocked lease=%#v err=%v", blocked, err)
	}
	reacquired, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: now.Add(2 * time.Second)})
	if err != nil || !reacquired.Acquired || reacquired.Fence != 2 {
		t.Fatalf("reacquired lease=%#v err=%v", reacquired, err)
	}
	renewed, err := store.Renew(ctx, reacquired, 2*time.Minute, now.Add(3*time.Second))
	if err != nil || !renewed.Acquired || renewed.Fence != reacquired.Fence || !renewed.ExpiresAt.Equal(now.Add(123*time.Second)) {
		t.Fatalf("renewed lease=%#v err=%v", renewed, err)
	}
	if err := store.Release(ctx, first); !errors.Is(err, domain.ErrSchedulerLeaseNotHeld) {
		t.Fatalf("release stale lease: %v", err)
	}
	if err := store.Release(ctx, renewed); err != nil {
		t.Fatalf("release renewed lease: %v", err)
	}
	status, err = store.Inspect(ctx, "daily-reset", now.Add(4*time.Second))
	if err != nil || status.Held {
		t.Fatalf("released status=%#v err=%v", status, err)
	}
	second, err := store.TryAcquire(ctx, domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-b", TTL: time.Minute, Now: now.Add(5 * time.Minute)})
	if err != nil || !second.Acquired || second.Fence != 3 {
		t.Fatalf("second lease=%#v err=%v", second, err)
	}
}

func TestBaseRepositoryNilContract(t *testing.T) {
	var repository *BaseRepository
	if repository.CollectionName() != "" {
		t.Fatal("nil repository must have an empty collection name")
	}
	if err := repository.Ping(context.Background()); err == nil {
		t.Fatal("nil repository ping must fail")
	}
}

func TestNewSchedulerLeaseStoreRequiresCollection(t *testing.T) {
	if _, err := NewSchedulerLeaseStore(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewSchedulerLeaseStoreFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewSchedulerLeaseStoreFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestSchedulerLocksCollectionName(t *testing.T) {
	if SchedulerLocksCollectionName != "mhcat_scheduler_locks" {
		t.Fatalf("collection = %s", SchedulerLocksCollectionName)
	}
}

func TestSchedulerLeaseAcquireFilter(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	request := domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: now}
	filter := schedulerLeaseAcquireFilter(request)
	if got := stringValue(filter, "_id"); got != "daily-reset" {
		t.Fatalf("_id = %q", got)
	}
	orValue, ok := lookup(filter, "$or").(bson.A)
	if !ok || len(orValue) != 2 {
		t.Fatalf("expected two $or branches, got %#v", lookup(filter, "$or"))
	}
}

func TestSchedulerLeaseAcquireUpdate(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	request := domain.SchedulerLeaseRequest{Name: "daily-reset", Owner: "worker-a", TTL: time.Minute, Now: now}
	update := schedulerLeaseAcquireUpdate(request)
	set, ok := lookup(update, "$set").(bson.D)
	if !ok {
		t.Fatalf("missing $set: %#v", update)
	}
	if got := stringValue(set, "owner"); got != "worker-a" {
		t.Fatalf("owner = %q", got)
	}
	if got := lookup(set, "expires_at").(time.Time); !got.Equal(now.Add(time.Minute)) {
		t.Fatalf("expires_at = %v", got)
	}
	inc, ok := lookup(update, "$inc").(bson.D)
	if !ok || lookup(inc, "fence") != int64(1) {
		t.Fatalf("expected fence increment, got %#v", lookup(update, "$inc"))
	}
}

func TestSchedulerLeaseHeldFilterRequiresFenceAndUnexpired(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	lease := domain.SchedulerLease{Name: "daily-reset", Owner: "worker-a", Fence: 7, Acquired: true, ExpiresAt: now.Add(time.Minute)}
	filter := schedulerLeaseHeldFilter(lease, now)
	if got := stringValue(filter, "_id"); got != "daily-reset" {
		t.Fatalf("_id = %q", got)
	}
	if got := stringValue(filter, "owner"); got != "worker-a" {
		t.Fatalf("owner = %q", got)
	}
	if got := lookup(filter, "fence"); got != int64(7) {
		t.Fatalf("fence = %#v", got)
	}
	expires, ok := lookup(filter, "expires_at").(bson.D)
	if !ok || lookup(expires, "$gt") != now {
		t.Fatalf("expires filter = %#v", lookup(filter, "expires_at"))
	}
}

func TestSchedulerLeaseDocumentFallbackID(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	document := schedulerLeaseDocument{ID: "daily-reset", Owner: "worker-a", Fence: 1, ExpiresAt: now}
	lease := document.toDomain(true)
	if lease.Name != "daily-reset" || !lease.Acquired {
		t.Fatalf("lease = %#v", lease)
	}
}

func TestSchedulerLeaseDocumentStatus(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	document := schedulerLeaseDocument{ID: "daily-reset", LockName: "daily-reset", Owner: "worker-a", Fence: 2, ExpiresAt: now.Add(time.Minute)}
	status := document.toStatus(now)
	if !status.Held || status.Owner != "worker-a" || status.Fence != 2 {
		t.Fatalf("status = %#v", status)
	}
	expired := document.toStatus(now.Add(2 * time.Minute))
	if expired.Held {
		t.Fatalf("expired status should not be held: %#v", expired)
	}
}

func lookup(document bson.D, key string) any {
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	return nil
}

func stringValue(document bson.D, key string) string {
	value, _ := lookup(document, key).(string)
	return value
}
