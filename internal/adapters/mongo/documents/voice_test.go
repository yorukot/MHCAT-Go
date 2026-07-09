package documents_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestVoiceRoomConfigDocumentRoundTrip(t *testing.T) {
	config := domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "voice-1",
		ParentID:         "category-1",
		Name:             "{name} 的包廂",
		Limit:            8,
		Lock:             true,
	}
	document := documents.VoiceRoomConfigDocumentFromDomain(config)
	if document.Guild != "guild-1" || document.TicketChannel != "voice-1" || document.Parent != "category-1" || document.Name != "{name} 的包廂" || document.Limit != 8 || !document.Lock {
		t.Fatalf("document = %#v", document)
	}
	if got := document.ToDomain(); got != config {
		t.Fatalf("round trip = %#v", got)
	}
}
