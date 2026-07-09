package repositories

import "testing"

func TestNewRedeemRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewRedeemRepository(nil, nil); err == nil {
		t.Fatal("expected nil codes collection error")
	}
}

func TestNewRedeemRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewRedeemRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}
