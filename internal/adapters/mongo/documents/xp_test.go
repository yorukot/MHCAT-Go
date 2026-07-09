package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestTextXPChannelDocumentRoundTrip(t *testing.T) {
	document := TextXPChannelDocumentFromDomain(domain.TextXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "#00ff00",
		Message:   "{user} 升到 {level}",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" || document.Color != "#00ff00" || document.Message == "" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" || config.Color != "#00ff00" || config.Message == "" {
		t.Fatalf("config = %#v", config)
	}
}

func TestVoiceXPChannelDocumentRoundTrip(t *testing.T) {
	document := VoiceXPChannelDocumentFromDomain(domain.VoiceXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "#00ff00",
		Message:   "{user} 升到 {level}",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" || document.Color != "#00ff00" || document.Message == "" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" || config.Color != "#00ff00" || config.Message == "" {
		t.Fatalf("config = %#v", config)
	}
}
