package repositories

import "testing"

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
