package repositories

import "testing"

func TestNewStatsConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewStatsConfigRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestStatsRoleConfigCollectionName(t *testing.T) {
	if StatsRoleConfigCollectionName != "role_numbers" {
		t.Fatalf("role stats collection = %s, want role_numbers", StatsRoleConfigCollectionName)
	}
}

func TestNewStatsConfigRepositoryWithRoleNumbersRequiresCollections(t *testing.T) {
	if _, err := NewStatsConfigRepositoryWithRoleNumbers(nil, nil); err == nil {
		t.Fatal("expected nil numbers collection error")
	}
}

func TestNewStatsConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewStatsConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}
