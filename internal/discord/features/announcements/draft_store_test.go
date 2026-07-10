package announcements

import (
	"errors"
	"testing"
	"time"
)

func TestDraftStoreExpiresAtLegacyDeadline(t *testing.T) {
	now := time.Unix(100, 0)
	store := NewDraftStore()
	store.now = func() time.Time { return now }
	id, err := store.Put(AnnouncementDraft{GuildID: "guild-1", UserID: "user-1"})
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	now = now.Add(defaultDraftTTL)
	if _, err := store.TakeForActor(id, "guild-1", "user-1"); !errors.Is(err, ErrAnnouncementDraftNotFound) {
		t.Fatalf("draft must expire at %v: %v", defaultDraftTTL, err)
	}
}

func TestDraftStoreUnauthorizedActorDoesNotConsumeState(t *testing.T) {
	store := NewDraftStore()
	id, err := store.Put(AnnouncementDraft{GuildID: "guild-1", UserID: "owner-1"})
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if _, err := store.TakeForActor(id, "guild-1", "other-user"); !errors.Is(err, ErrAnnouncementDraftUnauthorized) {
		t.Fatalf("unauthorized take: %v", err)
	}
	if _, err := store.TakeForActor(id, "other-guild", "owner-1"); !errors.Is(err, ErrAnnouncementDraftUnauthorized) {
		t.Fatalf("cross-guild take: %v", err)
	}
	if _, err := store.TakeForActor(id, "guild-1", "owner-1"); err != nil {
		t.Fatalf("owner take after denial: %v", err)
	}
}
