package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestAutoChatConfigDocumentRoundTrip(t *testing.T) {
	document := AutoChatConfigDocumentFromDomain(domain.AutoChatConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" {
		t.Fatalf("config = %#v", config)
	}
}
