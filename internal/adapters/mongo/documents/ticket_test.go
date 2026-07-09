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
	var document TicketConfigDocument
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
	var document TicketConfigDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode partial bson: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.CategoryID != "" || config.AdminRoleID != "" || config.EveryoneRoleID != "" {
		t.Fatalf("partial config = %#v", config)
	}
}
