package repositories

import "testing"

func TestNewEconomyProfileRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewEconomyProfileRepository(nil, nil, nil, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil collections")
	}
}

func TestNewEconomyProfileRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewEconomyProfileRepositoryFromDatabase(nil); err == nil {
		t.Fatalf("expected error for nil database")
	}
}

func TestEconomyProfileCollectionNames(t *testing.T) {
	if TextXPCollectionName != "text_xps" {
		t.Fatalf("text xp collection = %s", TextXPCollectionName)
	}
	if VoiceXPCollectionName != "voice_xps" {
		t.Fatalf("voice xp collection = %s", VoiceXPCollectionName)
	}
}
