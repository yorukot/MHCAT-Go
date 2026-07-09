package mongo

import (
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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
