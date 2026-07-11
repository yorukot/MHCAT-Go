package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGoodWebConfigCollectionName(t *testing.T) {
	if GoodWebConfigCollectionName != "good_webs" {
		t.Fatalf("anti-scam config collection = %s, want good_webs", GoodWebConfigCollectionName)
	}
	if NotAGoodWebCollectionName != "not_a_good_webs" {
		t.Fatalf("scam URL catalog collection = %s, want not_a_good_webs", NotAGoodWebCollectionName)
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

func TestNewScamURLCatalogRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewScamURLCatalogRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewScamURLCatalogRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewScamURLCatalogRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestAntiScamConfigUpdatePreservesLegacyFields(t *testing.T) {
	update, err := antiScamConfigUpdate(documents.GoodWebConfigDocument{
		Guild: "guild-1",
		Open:  true,
	})
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

func TestScamURLContainsFilterEscapesUserInput(t *testing.T) {
	filter := scamURLContainsFilter("https://bad.example/a?x=(.*)")
	web := documentValue(t, filter, "web")
	regexDoc := web.(bson.D)
	if value := documentValue(t, regexDoc, "$regex"); value != `https://bad\.example/a\?x=\(\.\*\)` {
		t.Fatalf("regex = %#v", value)
	}
}

func TestScamURLContainsFilterPreservesLegacyInputWhitespace(t *testing.T) {
	filter := scamURLContainsFilter(" https://bad.example ")
	web := documentValue(t, filter, "web").(bson.D)
	if value := documentValue(t, web, "$regex"); value != ` https://bad\.example ` {
		t.Fatalf("regex = %#v", value)
	}
}
