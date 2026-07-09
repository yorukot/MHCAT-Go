package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestLoggingConfigDocumentRoundTripPreservesLegacyFields(t *testing.T) {
	config := domain.LoggingConfig{
		GuildID:           "guild-1",
		ChannelID:         "channel-1",
		MessageUpdate:     true,
		MessageDelete:     true,
		ChannelUpdate:     false,
		MemberVoiceUpdate: true,
	}
	document := LoggingConfigDocumentFromDomain(config)
	if document.Guild != "guild-1" || document.ChannelID != "channel-1" || !document.MessageUpdate || !document.MessageDelete || document.ChannelUpdate || !document.MemberVoiceUpdate {
		t.Fatalf("document = %#v", document)
	}
	got := document.ToDomain()
	if got != config {
		t.Fatalf("round trip = %#v want %#v", got, config)
	}
}
