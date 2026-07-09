package repositories

import "testing"

func TestNewVoiceRoomConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewVoiceRoomConfigRepository(nil); err == nil {
		t.Fatal("expected collection validation error")
	}
}

func TestNewVoiceRoomConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewVoiceRoomConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database validation error")
	}
}
