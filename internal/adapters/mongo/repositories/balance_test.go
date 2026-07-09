package repositories

import "testing"

func TestNewBalanceRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewBalanceRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewBalanceRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewBalanceRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}
