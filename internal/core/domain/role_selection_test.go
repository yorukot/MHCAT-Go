package domain

import (
	"errors"
	"testing"
)

func TestParseDiscordMessageURLAcceptsLegacyHosts(t *testing.T) {
	for _, raw := range []string{
		"https://discord.com/channels/guild-1/channel-1/message-1",
		"https://discordapp.com/channels/guild-1/channel-1/message-1",
		"<https://discord.com/channels/guild-1/channel-1/message-1>",
	} {
		target, err := ParseDiscordMessageURL(raw)
		if err != nil {
			t.Fatalf("ParseDiscordMessageURL(%q): %v", raw, err)
		}
		if target.GuildID != "guild-1" || target.ChannelID != "channel-1" || target.MessageID != "message-1" {
			t.Fatalf("target = %#v", target)
		}
	}
}

func TestParseDiscordMessageURLRejectsInvalidURL(t *testing.T) {
	_, err := ParseDiscordMessageURL("https://example.com/channels/guild/channel/message")
	if !errors.Is(err, ErrInvalidRoleSelectionConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestNormalizeLegacyReaction(t *testing.T) {
	reaction, err := NormalizeLegacyReaction("<:mhcat:123456789012345678>")
	if err != nil {
		t.Fatalf("NormalizeLegacyReaction custom: %v", err)
	}
	if reaction.Stored != "123456789012345678" || reaction.API != "mhcat:123456789012345678" {
		t.Fatalf("custom reaction = %#v", reaction)
	}
	reaction, err = NormalizeLegacyReaction("✅")
	if err != nil {
		t.Fatalf("NormalizeLegacyReaction unicode: %v", err)
	}
	if reaction.Stored != "✅" || reaction.API != "✅" {
		t.Fatalf("unicode reaction = %#v", reaction)
	}
}
