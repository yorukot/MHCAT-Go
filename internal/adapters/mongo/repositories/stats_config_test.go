package repositories

import "testing"

func TestNewStatsConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewStatsConfigRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewStatsConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewStatsConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}
