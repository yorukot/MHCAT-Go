package repositories

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRoleSelectionCollectionNames(t *testing.T) {
	if RoleReactionCollectionName != "message_reactions" || RoleButtonCollectionName != "btns" {
		t.Fatalf("collections = %q, %q", RoleReactionCollectionName, RoleButtonCollectionName)
	}
}

func TestNewRoleSelectionRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewRoleSelectionRepository(nil, nil); err == nil {
		t.Fatal("expected reactions collection error")
	}
	if _, err := NewRoleSelectionRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database error")
	}
}

func TestRoleReactionConfigFilterPreservesLegacyLogicalKey(t *testing.T) {
	want := bson.D{{Key: "guild", Value: "guild-1"}, {Key: "message", Value: "message-1"}, {Key: "react", Value: "emoji-1"}}
	if got := roleReactionConfigFilter("guild-1", "message-1", "emoji-1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("filter = %#v", got)
	}
}

func TestRoleReactionConfigUpdateAlignsDuplicatesAndUpsertsKeys(t *testing.T) {
	document := documents.RoleReactionDocument{Guild: "guild-1", Message: "message-1", React: "emoji-1", Role: "role-1"}
	update, err := roleReactionConfigUpdate(document, false)
	if err != nil {
		t.Fatalf("build duplicate update: %v", err)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("duplicate update should preserve existing keys: %#v", update)
	}
	if value := documentValue(t, documentValue(t, update, "$set"), "role"); value != "role-1" {
		t.Fatalf("role = %#v", value)
	}

	upsert, err := roleReactionConfigUpdate(document, true)
	if err != nil {
		t.Fatalf("build upsert: %v", err)
	}
	setOnInsert := documentValue(t, upsert, "$setOnInsert")
	for key, want := range map[string]string{"guild": "guild-1", "message": "message-1", "react": "emoji-1"} {
		if got := documentValue(t, setOnInsert, key); got != want {
			t.Fatalf("%s = %#v, want %q", key, got, want)
		}
	}
}

func TestRoleButtonConfigUpdateAlignsDuplicatesAndUpsertsKeys(t *testing.T) {
	document := documents.RoleButtonDocument{Guild: "guild-1", Number: "number-1", Role: "role-1"}
	update, err := roleButtonConfigUpdate(document, false)
	if err != nil {
		t.Fatalf("build duplicate update: %v", err)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("duplicate update should preserve existing keys: %#v", update)
	}
	if value := documentValue(t, documentValue(t, update, "$set"), "role"); value != "role-1" {
		t.Fatalf("role = %#v", value)
	}

	upsert, err := roleButtonConfigUpdate(document, true)
	if err != nil {
		t.Fatalf("build upsert: %v", err)
	}
	setOnInsert := documentValue(t, upsert, "$setOnInsert")
	if guild := documentValue(t, setOnInsert, "guild"); guild != "guild-1" {
		t.Fatalf("guild = %#v", guild)
	}
	if number := documentValue(t, setOnInsert, "number"); number != "number-1" {
		t.Fatalf("number = %#v", number)
	}
}
