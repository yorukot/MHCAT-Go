package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

func TestStatsConfigUpdatePreservesLegacyInitialDocumentShape(t *testing.T) {
	update := statsConfigUpdate(domain.StatsConfig{
		GuildID:          "guild-1",
		ParentID:         "parent-1",
		MemberNumberID:   "member-1",
		MemberNumberName: "10",
		UserNumberID:     "user-1",
		UserNumberName:   "8",
		BotNumberID:      "bot-1",
		BotNumberName:    "2",
	}, true)

	set := statsConfigUpdateDocument(t, update, "$set")
	wantValues := map[string]string{
		"parent":            "parent-1",
		"memberNumber":      "member-1",
		"memberNumber_name": "10",
		"userNumber":        "user-1",
		"userNumber_name":   "8",
		"BotNumber":         "bot-1",
		"BotNumber_name":    "2",
	}
	nullFields := []string{
		"channelnumber", "channelnumber_name",
		"textnumber", "textnumber_name",
		"voicenumber", "voicenumber_name",
		"categoriesnumber", "categoriesnumber_name",
		"rolesnumber", "rolesnumber_name",
		"rolenumber", "rolenumber_name",
		"norolenumber", "norolenumber_name",
		"emojisnumber", "emojisnumber_name",
		"staticnumber", "staticnumber_name",
		"gifnumber", "gifnumber_name",
		"stickersnumber", "stickersnumber_name",
		"boostsnumber", "boostsnumber_name",
		"tier", "tier_name",
		"onlinenumber", "onlinenumber_name",
		"dndnumber", "dndnumber_name",
		"idlenumber", "idlenumber_name",
		"offlinenumber", "offlinenumber_name",
		"onlinebotnumber", "onlinebotnumber_name",
		"statusnumber", "statusnumber_name",
	}
	if len(set) != len(wantValues)+len(nullFields) {
		t.Fatalf("$set field count = %d, want %d: %#v", len(set), len(wantValues)+len(nullFields), set)
	}
	for field, want := range wantValues {
		if got := statsConfigDocumentValue(t, set, field); got != want {
			t.Fatalf("%s = %#v, want %q", field, got, want)
		}
	}
	for _, field := range nullFields {
		if got := statsConfigDocumentValue(t, set, field); got != nil {
			t.Fatalf("%s = %#v, want BSON null", field, got)
		}
	}

	setOnInsert := statsConfigUpdateDocument(t, update, "$setOnInsert")
	if len(setOnInsert) != 1 || statsConfigDocumentValue(t, setOnInsert, "guild") != "guild-1" {
		t.Fatalf("$setOnInsert = %#v", setOnInsert)
	}
}

func statsConfigUpdateDocument(t *testing.T, update bson.D, key string) bson.D {
	t.Helper()
	value := statsConfigDocumentValue(t, update, key)
	document, ok := value.(bson.D)
	if !ok {
		t.Fatalf("%s = %#v, want bson.D", key, value)
	}
	return document
}

func statsConfigDocumentValue(t *testing.T, document bson.D, key string) any {
	t.Helper()
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	t.Fatalf("missing BSON field %q in %#v", key, document)
	return nil
}
