package birthday

import (
	"testing"
	"time"
)

func TestBirthdayAddStateExpiresAtDeadline(t *testing.T) {
	now := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	store := newBirthdayAddStateStore()
	id := store.create(now, pendingBirthdayAdd{OwnerUserID: "user-1", ExpiresAt: now.Add(birthdayAddTimeout)})

	if _, ok := store.get(id, now.Add(birthdayAddTimeout-time.Nanosecond)); !ok {
		t.Fatal("state expired before its deadline")
	}
	if _, ok := store.get(id, now.Add(birthdayAddTimeout)); ok {
		t.Fatal("state remained available at its expiry deadline")
	}
}

func TestBirthdayAddStateCreatePrunesExpiredEntries(t *testing.T) {
	now := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	store := newBirthdayAddStateStore()
	expiredID := store.create(now, pendingBirthdayAdd{OwnerUserID: "user-1", ExpiresAt: now.Add(birthdayAddTimeout)})
	activeID := store.create(now.Add(birthdayAddTimeout), pendingBirthdayAdd{OwnerUserID: "user-2", ExpiresAt: now.Add(2 * birthdayAddTimeout)})

	if _, ok := store.entries[expiredID]; ok {
		t.Fatal("expired state was not pruned")
	}
	if _, ok := store.entries[activeID]; !ok {
		t.Fatal("new active state was not stored")
	}
}
