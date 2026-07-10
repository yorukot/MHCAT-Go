package documents_test

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestVoiceRoomConfigReadDocumentUsesMongooseScalarCoercion(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "ticket_channel", Value: "voice-1"},
		{Key: "limit", Value: "7"},
		{Key: "name", Value: int32(42)},
		{Key: "parent", Value: nil},
		{Key: "lock", Value: "yes"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.VoiceRoomConfigReadDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.TriggerChannelID != "voice-1" || config.Limit != 7 || config.Name != "42" || config.ParentID != "" || !config.Lock {
		t.Fatalf("config = %#v", config)
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
		*document.LockAnswer != " secret " ||
		document.Owner != " owner-1 " ||
		document.TextChannel == nil ||
		*document.TextChannel != " text-1 " ||
		!reflect.DeepEqual(document.AllowedUsers, []string{" user-2 ", "user-3"}) {
		t.Fatalf("document = %#v", document)
	}
	document.AllowedUsers[0] = "mutated"
	got := document.ToDomain()
	if got.GuildID != "guild-1" ||
		got.ChannelID != "voice-1" ||
		got.Password != " secret " ||
		!got.PasswordPresent ||
		got.OwnerID != " owner-1 " ||
		got.TextChannelID != " text-1 " ||
		!reflect.DeepEqual(got.AllowedUserIDs, []string{"mutated", "user-3"}) {
		t.Fatalf("domain = %#v", got)
	}
}

func TestVoiceRoomLockReadDocumentUsesMongooseScalarAndMixedArrayCoercion(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "channel_id", Value: "voice-1"},
		{Key: "lock_anser", Value: int32(1234)},
		{Key: "owner", Value: int64(42)},
		{Key: "text_channel", Value: true},
		{Key: "ok_people", Value: bson.A{"user-1", " user-2 ", int32(3), false}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.VoiceRoomLockReadDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	lock := document.ToDomain()
	if lock.GuildID != "guild-1" || lock.ChannelID != "voice-1" || lock.Password != "1234" || !lock.PasswordPresent || lock.OwnerID != "42" || lock.TextChannelID != "true" || !reflect.DeepEqual(lock.AllowedUserIDs, []string{"user-1", " user-2 "}) {
		t.Fatalf("lock = %#v", lock)
	}
}

func TestVoiceRoomLockReadDocumentWrapsLegacyScalarArrayValue(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "channel_id", Value: "voice-1"},
		{Key: "lock_anser", Value: nil},
		{Key: "owner", Value: "owner-1"},
		{Key: "text_channel", Value: nil},
		{Key: "ok_people", Value: "user-1"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.VoiceRoomLockReadDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	lock := document.ToDomain()
	if lock.PasswordPresent || lock.TextChannelID != "" || !reflect.DeepEqual(lock.AllowedUserIDs, []string{"user-1"}) {
		t.Fatalf("lock = %#v", lock)
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
		TextChannel: ptr("text-1"),
	}.ToDomain()
	if got.Password != "" {
		t.Fatalf("null lock_anser should map to empty password, got %#v", got)
	}
	if got.PasswordPresent {
		t.Fatalf("null lock_anser should remain absent, got %#v", got)
	}

	empty := ""
	explicitEmpty := documents.VoiceRoomLockDocument{
		Guild:       "guild-1",
		ChannelID:   "voice-1",
		LockAnswer:  &empty,
		Owner:       "owner-1",
		TextChannel: ptr("text-1"),
	}.ToDomain()
	if !explicitEmpty.PasswordPresent || !explicitEmpty.HasPassword() {
		t.Fatalf("empty string lock_anser should remain present, got %#v", explicitEmpty)
	}
	encoded := documents.VoiceRoomLockDocumentFromDomain(explicitEmpty)
	if encoded.LockAnswer == nil || *encoded.LockAnswer != "" {
		t.Fatalf("explicit empty password should round trip, got %#v", encoded.LockAnswer)
	}
}

func TestVoiceRoomLockDocumentWritesEmptyAllowedUsersAsBSONArray(t *testing.T) {
	document := documents.VoiceRoomLockDocumentFromDomain(domain.VoiceRoomLock{
		GuildID:   "guild-1",
		ChannelID: "voice-1",
		OwnerID:   "owner-1",
	})
	if document.AllowedUsers == nil || len(document.AllowedUsers) != 0 {
		t.Fatalf("allowed users = %#v", document.AllowedUsers)
	}
	raw, err := bson.Marshal(document)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	array, ok := bson.Raw(raw).Lookup("ok_people").ArrayOK()
	if !ok {
		t.Fatalf("ok_people is not an array: %v", bson.Raw(raw).Lookup("ok_people").Type)
	}
	values, err := array.Values()
	if err != nil || len(values) != 0 {
		t.Fatalf("ok_people values=%#v err=%v", values, err)
	}
}

func TestVoiceRoomLockDocumentUsesNullForEmptyTextChannel(t *testing.T) {
	document := documents.VoiceRoomLockDocumentFromDomain(domain.VoiceRoomLock{
		GuildID:   "guild-1",
		ChannelID: "voice-1",
		OwnerID:   "owner-1",
	})
	if document.TextChannel != nil {
		t.Fatalf("empty text channel should use null, got %#v", document.TextChannel)
	}
	got := documents.VoiceRoomLockDocument{
		Guild:       "guild-1",
		ChannelID:   "voice-1",
		LockAnswer:  nil,
		Owner:       "owner-1",
		TextChannel: nil,
	}.ToDomain()
	if got.TextChannelID != "" {
		t.Fatalf("null text_channel should map to empty text channel, got %#v", got)
	}
}

func TestVoiceRoomStateDocumentRoundTrip(t *testing.T) {
	state := domain.VoiceRoomState{GuildID: "guild-1", ChannelID: "voice-1"}
	document := documents.VoiceRoomStateDocumentFromDomain(state)
	if document.Guild != "guild-1" || document.ChannelID != "voice-1" {
		t.Fatalf("document = %#v", document)
	}
	if got := document.ToDomain(); got != state {
		t.Fatalf("round trip = %#v", got)
	}
}

func ptr(value string) *string {
	return &value
}
