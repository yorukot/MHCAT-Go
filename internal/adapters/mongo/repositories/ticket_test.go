package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestTicketConfigCollectionName(t *testing.T) {
	if TicketConfigCollectionName != "tickets" {
		t.Fatalf("ticket collection = %s, want tickets", TicketConfigCollectionName)
	}
}

func TestTicketConfigCreationFilterScopesRollbackToInsertedRow(t *testing.T) {
	creationID := bson.NewObjectID()
	filter, err := ticketConfigCreationFilter(ports.TicketConfigCreation{
		GuildID: " guild-1 ",
		ID:      creationID.Hex(),
	})
	if err != nil {
		t.Fatalf("build creation filter: %v", err)
	}
	if got := documentValue(t, filter, "_id"); got != creationID {
		t.Fatalf("_id = %#v, want %#v", got, creationID)
	}
	if got := documentValue(t, filter, "guild"); got != "guild-1" {
		t.Fatalf("guild = %#v, want guild-1", got)
	}
}

func TestNewTicketConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewTicketConfigRepository(nil); err == nil {
		t.Fatal("expected collection validation error")
	}
}

func TestNewTicketConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewTicketConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database validation error")
	}
}

func TestTicketConfigCreateUpdateOnlySetsFieldsOnInsert(t *testing.T) {
	document := documents.TicketConfigDocument{
		Guild:         "guild-1",
		TicketChannel: "category-1",
		AdminID:       "admin-role-1",
		EveryoneID:    "everyone-role-1",
	}
	creationID := bson.NewObjectID()
	update, err := ticketConfigCreateUpdate(document, creationID)
	if err != nil {
		t.Fatalf("build create update: %v", err)
	}
	if hasKey(update, "$set") {
		t.Fatalf("create update must not mutate an existing config: %#v", update)
	}
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if got := documentValue(t, setOnInsert, "_id"); got != creationID {
		t.Fatalf("_id = %#v, want %#v", got, creationID)
	}
	for key, want := range map[string]string{
		"guild":          "guild-1",
		"ticket_channel": "category-1",
		"admin_id":       "admin-role-1",
		"everyone_id":    "everyone-role-1",
	} {
		if got := documentValue(t, setOnInsert, key); got != want {
			t.Fatalf("%s = %#v, want %q", key, got, want)
		}
	}
}
