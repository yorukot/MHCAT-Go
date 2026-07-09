package repositories

import "testing"

func TestEconomyCollectionNames(t *testing.T) {
	if CoinCollectionName != "coins" {
		t.Fatalf("coin collection = %s", CoinCollectionName)
	}
	if GiftChangeCollectionName != "gift_changes" {
		t.Fatalf("gift_change collection = %s", GiftChangeCollectionName)
	}
	if SignListCollectionName != "sign_lists" {
		t.Fatalf("sign_list collection = %s", SignListCollectionName)
	}
}

func TestNewEconomyRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewEconomyRepository(nil, nil, nil); err == nil {
		t.Fatalf("expected nil collection error")
	}
}

func TestNewEconomyRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewEconomyRepositoryFromDatabase(nil); err == nil {
		t.Fatalf("expected nil database error")
	}
}
