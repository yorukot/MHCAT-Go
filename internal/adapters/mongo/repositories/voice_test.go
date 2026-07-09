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

func TestVoiceRoomCollectionNames(t *testing.T) {
	if VoiceRoomConfigCollectionName != "voice_channels" {
		t.Fatalf("voice-room config collection = %s", VoiceRoomConfigCollectionName)
	}
	if VoiceRoomLockCollectionName != "lock_channels" {
		t.Fatalf("voice-room lock collection = %s", VoiceRoomLockCollectionName)
	}
}

func TestNewVoiceRoomLockRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewVoiceRoomLockRepository(nil); err == nil {
		t.Fatal("expected collection validation error")
	}
}

func TestNewVoiceRoomLockRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewVoiceRoomLockRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database validation error")
	}
}
