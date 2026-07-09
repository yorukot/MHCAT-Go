package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestGoodWebConfigDocumentRoundTrip(t *testing.T) {
	document := GoodWebConfigDocumentFromDomain(domain.AntiScamConfig{
		GuildID: "guild-1",
		Open:    true,
	})
	if document.Guild != "guild-1" || !document.Open {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || !config.Open {
		t.Fatalf("config = %#v", config)
	}
}
