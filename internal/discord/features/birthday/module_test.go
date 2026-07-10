package birthday

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestBirthdayAddStateExpiresAtDeadline(t *testing.T) {
	now := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	store := newBirthdayAddStateStore()
	id, err := store.create(now, pendingBirthdayAdd{OwnerUserID: "user-1", ExpiresAt: now.Add(birthdayAddTimeout)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

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
	ids := []string{"expired", "active"}
	store.randomID = func() (string, error) {
		id := ids[0]
		ids = ids[1:]
		return id, nil
	}
	expiredID, err := store.create(now, pendingBirthdayAdd{OwnerUserID: "user-1", ExpiresAt: now.Add(birthdayAddTimeout)})
	if err != nil {
		t.Fatalf("create expired: %v", err)
	}
	activeID, err := store.create(now.Add(birthdayAddTimeout), pendingBirthdayAdd{OwnerUserID: "user-2", ExpiresAt: now.Add(2 * birthdayAddTimeout)})
	if err != nil {
		t.Fatalf("create active: %v", err)
	}

	if _, ok := store.entries[expiredID]; ok {
		t.Fatal("expired state was not pruned")
	}
	if _, ok := store.entries[activeID]; !ok {
		t.Fatal("new active state was not stored")
	}
}

func TestBirthdayAddStateUsesRandomHexIDs(t *testing.T) {
	id, err := randomBirthdayAddStateID()
	if err != nil {
		t.Fatalf("random id: %v", err)
	}
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 12 {
		t.Fatalf("state id = %q, decoded=%d, err=%v", id, len(decoded), err)
	}
}

func TestBirthdayAddStateRejectsRepeatedIDCollisions(t *testing.T) {
	store := newBirthdayAddStateStore()
	store.entries["collision"] = pendingBirthdayAdd{}
	store.randomID = func() (string, error) { return "collision", nil }

	if _, err := store.create(time.Time{}, pendingBirthdayAdd{}); !errors.Is(err, errBirthdayAddStateID) {
		t.Fatalf("create error = %v", err)
	}
}
