package documents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestTicketConfigDocumentLegacyFixtureDecodes(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "mongo", "ticket_config_legacy.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var raw map[string]string
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("decode json fixture: %v", err)
	}
	bsonPayload, err := bson.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal fixture bson: %v", err)
	}
	var document TicketConfigReadDocument
	if err := bson.Unmarshal(bsonPayload, &document); err != nil {
		t.Fatalf("decode bson fixture: %v", err)
	}
	got := document.ToDomain()
	if got.GuildID != raw["guild"] ||
		got.CategoryID != raw["ticket_channel"] ||
		got.AdminRoleID != raw["admin_id"] ||
		got.EveryoneRoleID != raw["everyone_id"] {
		t.Fatalf("domain config = %#v, raw = %#v", got, raw)
	}
}

func TestTicketConfigDocumentRoundTripDomain(t *testing.T) {
	config := domain.TicketConfig{
		GuildID:        "guild-1",
		CategoryID:     "category-1",
		AdminRoleID:    "admin-role-1",
		EveryoneRoleID: "everyone-role-1",
	}
	document := TicketConfigDocumentFromDomain(config)
	if got := document.ToDomain(); got != config {
		t.Fatalf("round trip = %#v, want %#v", got, config)
	}
}

func TestTicketConfigDocumentMissingFieldsDecodeSafe(t *testing.T) {
	payload, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}})
	if err != nil {
		t.Fatalf("marshal partial bson: %v", err)
	}
	var document TicketConfigReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode partial bson: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.CategoryID != "" || config.AdminRoleID != "" || config.EveryoneRoleID != "" {
		t.Fatalf("partial config = %#v", config)
	}
}

func TestTicketConfigReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	everyoneID := bson.NewObjectID()
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(123)},
		{Key: "ticket_channel", Value: true},
		{Key: "admin_id", Value: 42.5},
		{Key: "everyone_id", Value: everyoneID},
	})
	if err != nil {
		t.Fatalf("marshal mixed ticket config: %v", err)
	}
	var document TicketConfigReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode mixed ticket config: %v", err)
	}
	got := document.ToDomain()
	want := domain.TicketConfig{
		GuildID:        "123",
		CategoryID:     "true",
		AdminRoleID:    "42.5",
		EveryoneRoleID: everyoneID.Hex(),
	}
	if got != want {
		t.Fatalf("mixed ticket config = %#v, want %#v", got, want)
	}
}

func TestTicketConfigReadDocumentRejectsNonMongooseStringShapes(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: bson.A{"guild-1"}},
		{Key: "ticket_channel", Value: bson.D{{Key: "id", Value: "category-1"}}},
	})
	if err != nil {
		t.Fatalf("marshal malformed ticket config: %v", err)
	}
	var document TicketConfigReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode malformed ticket config: %v", err)
	}
	if got := document.ToDomain(); got != (domain.TicketConfig{}) {
		t.Fatalf("malformed ticket config = %#v, want empty", got)
	}
}
