package repositories

import "testing"

func TestNewWarningRepositoriesRequireCollections(t *testing.T) {
	if _, err := NewWarningHistoryRepository(nil); err == nil {
		t.Fatal("expected warning history collection requirement")
	}
	if _, err := NewWarningSettingsRepository(nil); err == nil {
		t.Fatal("expected warning settings collection requirement")
	}
}

func TestNewWarningRepositoriesFromDatabaseRequireDatabase(t *testing.T) {
	if _, err := NewWarningHistoryRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected warning history database requirement")
	}
	if _, err := NewWarningSettingsRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected warning settings database requirement")
	}
}
