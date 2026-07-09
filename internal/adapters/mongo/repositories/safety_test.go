package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestGoodWebConfigCollectionName(t *testing.T) {
	if GoodWebConfigCollectionName != "good_webs" {
		t.Fatalf("anti-scam config collection = %s, want good_webs", GoodWebConfigCollectionName)
	}
}

func TestNewAntiScamConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewAntiScamConfigRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewAntiScamConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewAntiScamConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestAntiScamConfigUpdatePreservesLegacyFields(t *testing.T) {
	update, err := antiScamConfigUpdate(documents.GoodWebConfigDocument{
		Guild: "guild-1",
		Open:  true,
	}, true)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "open"); value != true {
		t.Fatalf("open = %#v", value)
	}
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if value := documentValue(t, setOnInsert, "guild"); value != "guild-1" {
		t.Fatalf("guild setOnInsert = %#v", value)
	}
}

func TestAntiScamConfigUpdateOmitsSetOnInsertWhenNotUpserting(t *testing.T) {
	update, err := antiScamConfigUpdate(documents.GoodWebConfigDocumentFromDomain(domain.AntiScamConfig{
		GuildID: "guild-1",
		Open:    false,
	}), false)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("non-upsert update should not include setOnInsert: %#v", update)
	}
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "open"); value != false {
		t.Fatalf("open = %#v", value)
	}
}
