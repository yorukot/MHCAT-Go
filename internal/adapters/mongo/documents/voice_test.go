package documents_test

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestVoiceRoomConfigDocumentRoundTrip(t *testing.T) {
	config := domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "voice-1",
		ParentID:         "category-1",
		Name:             "{name} 的包廂",
		Limit:            8,
		Lock:             true,
	}
	document := documents.VoiceRoomConfigDocumentFromDomain(config)
	if document.Guild != "guild-1" || document.TicketChannel != "voice-1" || document.Parent != "category-1" || document.Name != "{name} 的包廂" || document.Limit != 8 || !document.Lock {
		t.Fatalf("document = %#v", document)
	}
	if got := document.ToDomain(); got != config {
		t.Fatalf("round trip = %#v", got)
	}
}

func TestVoiceRoomLockDocumentRoundTripPreservesLegacyFields(t *testing.T) {
	lock := domain.VoiceRoomLock{
		GuildID:        " guild-1 ",
		ChannelID:      " voice-1 ",
		Password:       " secret ",
		OwnerID:        " owner-1 ",
		TextChannelID:  " text-1 ",
		AllowedUserIDs: []string{" user-2 ", "user-3"},
	}
	document := documents.VoiceRoomLockDocumentFromDomain(lock)
	if document.Guild != "guild-1" ||
		document.ChannelID != "voice-1" ||
		document.LockAnswer == nil ||
		*document.LockAnswer != "secret" ||
		document.Owner != "owner-1" ||
		document.TextChannel != "text-1" ||
		!reflect.DeepEqual(document.AllowedUsers, []string{"user-2", "user-3"}) {
		t.Fatalf("document = %#v", document)
	}
	document.AllowedUsers[0] = "mutated"
	got := document.ToDomain()
	if got.GuildID != "guild-1" ||
		got.ChannelID != "voice-1" ||
		got.Password != "secret" ||
		got.OwnerID != "owner-1" ||
		got.TextChannelID != "text-1" ||
		!reflect.DeepEqual(got.AllowedUserIDs, []string{"mutated", "user-3"}) {
		t.Fatalf("domain = %#v", got)
	}
}

func TestVoiceRoomLockDocumentUsesNullForEmptyPassword(t *testing.T) {
	document := documents.VoiceRoomLockDocumentFromDomain(domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	})
	if document.LockAnswer != nil {
		t.Fatalf("empty password should use null lock_anser, got %#v", document.LockAnswer)
	}

	got := documents.VoiceRoomLockDocument{
		Guild:       "guild-1",
		ChannelID:   "voice-1",
		LockAnswer:  nil,
		Owner:       "owner-1",
		TextChannel: "text-1",
	}.ToDomain()
	if got.Password != "" {
		t.Fatalf("null lock_anser should map to empty password, got %#v", got)
	}
}
