package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
)

func TestTicketConfigCollectionName(t *testing.T) {
	if TicketConfigCollectionName != "tickets" {
		t.Fatalf("ticket collection = %s, want tickets", TicketConfigCollectionName)
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

func TestTicketConfigUpdateAlignsDuplicatesAndUpsertsGuild(t *testing.T) {
	document := documents.TicketConfigDocument{
		Guild:         "guild-1",
		TicketChannel: "category-1",
		AdminID:       "admin-role-1",
		EveryoneID:    "everyone-role-1",
	}
	update, err := ticketConfigUpdate(document, false)
	if err != nil {
		t.Fatalf("build duplicate update: %v", err)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("duplicate update should preserve existing guild keys: %#v", update)
	}
	set := documentValue(t, update, "$set")
	for key, want := range map[string]string{
		"ticket_channel": "category-1",
		"admin_id":       "admin-role-1",
		"everyone_id":    "everyone-role-1",
	} {
		if got := documentValue(t, set, key); got != want {
			t.Fatalf("%s = %#v, want %q", key, got, want)
		}
	}

	upsert, err := ticketConfigUpdate(document, true)
	if err != nil {
		t.Fatalf("build upsert: %v", err)
	}
	if guild := documentValue(t, documentValue(t, upsert, "$setOnInsert"), "guild"); guild != "guild-1" {
		t.Fatalf("upsert guild = %#v, want guild-1", guild)
	}
}
