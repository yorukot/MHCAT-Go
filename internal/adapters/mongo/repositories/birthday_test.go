package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestBirthdayConfigCollectionName(t *testing.T) {
	if BirthdayConfigCollectionName != "birthday_sets" {
		t.Fatalf("birthday config collection = %s, want birthday_sets", BirthdayConfigCollectionName)
	}
}

func TestNewBirthdayConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewBirthdayConfigRepository(nil); err == nil {
		t.Fatal("expected collection validation error")
	}
}

func TestNewBirthdayConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewBirthdayConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database validation error")
	}
}

func TestBirthdayConfigUpdatePreservesLegacyFields(t *testing.T) {
	role := "role-1"
	update, err := birthdayConfigUpdate(documents.BirthdayConfigDocument{
		Guild:                      "guild-1",
		Message:                    "{user} 生日快樂",
		UTCOffset:                  "+08:00",
		Channel:                    "channel-1",
		EveryoneCanSetBirthdayDate: true,
		Role:                       &role,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	for _, tc := range []struct {
		field string
		value any
	}{
		{field: "msg", value: "{user} 生日快樂"},
		{field: "utc", value: "+08:00"},
		{field: "channel", value: "channel-1"},
		{field: "everyone_can_set_birthday_date", value: true},
		{field: "role", value: "role-1"},
	} {
		if !bsonDHas(update, "$set", tc.field, tc.value) {
			t.Fatalf("missing %s set in %#v", tc.field, update)
		}
	}
}

func TestBirthdayConfigInsertSetsGuildOnInsert(t *testing.T) {
	document := documents.BirthdayConfigDocumentFromDomain(domain.BirthdayConfig{
		GuildID:   "guild-1",
		Message:   "happy",
		UTCOffset: "+08:00",
		ChannelID: "channel-1",
	})
	update, err := birthdayConfigInsertUpdate(document)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bsonDHas(update, "$setOnInsert", "guild", "guild-1") || !bsonDHas(update, "$set", "role", nil) {
		t.Fatalf("insert update = %#v", update)
	}
}
