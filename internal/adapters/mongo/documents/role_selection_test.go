package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRoleSelectionWriteDocumentsPreserveLegacyStringFields(t *testing.T) {
	reaction := RoleReactionDocumentFromDomain(domain.RoleReactionConfig{
		GuildID:   "guild-1",
		MessageID: "message-1",
		React:     "emoji-1",
		RoleID:    "role-1",
	})
	if reaction.Guild != "guild-1" || reaction.Message != "message-1" || reaction.React != "emoji-1" || reaction.Role != "role-1" {
		t.Fatalf("reaction document = %#v", reaction)
	}
	button := RoleButtonDocumentFromDomain(domain.RoleButtonConfig{GuildID: "guild-1", Number: "number-1", RoleID: "role-1"})
	if button.Guild != "guild-1" || button.Number != "number-1" || button.Role != "role-1" {
		t.Fatalf("button document = %#v", button)
	}
}

func TestRoleSelectionReadDocumentsUseMongooseScalarCoercion(t *testing.T) {
	objectID := bson.NewObjectID()
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(1234)},
		{Key: "message", Value: float64(1e20)},
		{Key: "react", Value: true},
		{Key: "role", Value: objectID},
	})
	if err != nil {
		t.Fatalf("marshal reaction: %v", err)
	}
	var reaction RoleReactionReadDocument
	if err := bson.Unmarshal(raw, &reaction); err != nil {
		t.Fatalf("unmarshal reaction: %v", err)
	}
	if got := reaction.ToDomain(); got != (domain.RoleReactionConfig{
		GuildID:   "1234",
		MessageID: "100000000000000000000",
		React:     "true",
		RoleID:    objectID.Hex(),
	}) {
		t.Fatalf("reaction config = %#v", got)
	}

	raw, err = bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "number", Value: float64(1e21)},
		{Key: "role", Value: int32(5678)},
	})
	if err != nil {
		t.Fatalf("marshal button: %v", err)
	}
	var button RoleButtonReadDocument
	if err := bson.Unmarshal(raw, &button); err != nil {
		t.Fatalf("unmarshal button: %v", err)
	}
	if got := button.ToDomain(); got != (domain.RoleButtonConfig{GuildID: "guild-1", Number: "1e+21", RoleID: "5678"}) {
		t.Fatalf("button config = %#v", got)
	}
}

func TestRoleSelectionReadDocumentsRejectNonMongooseStringShapes(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "message", Value: bson.A{"message-1"}},
		{Key: "react", Value: bson.D{{Key: "name", Value: "emoji"}}},
		{Key: "role", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document RoleReactionReadDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := document.ToDomain(); got != (domain.RoleReactionConfig{GuildID: "guild-1"}) {
		t.Fatalf("reaction config = %#v", got)
	}
}
