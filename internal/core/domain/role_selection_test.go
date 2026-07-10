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
	if reaction.Stored != "123456789012345678" || reaction.API != "mhcat:123456789012345678" || reaction.CustomEmojiID != "123456789012345678" {
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

func TestIsLegacyUnicodeEmojiMatchesJavaScriptRanges(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "unicode emoji", raw: "✅", want: true},
		{name: "unanchored", raw: "hello✅world", want: true},
		{name: "keycap", raw: "1️⃣", want: true},
		{name: "supplementary non emoji", raw: "𐍈", want: true},
		{name: "malformed class bracket", raw: "[", want: true},
		{name: "custom mention", raw: "<:mhcat:123456789012345678>", want: false},
		{name: "plain text", raw: "not-an-emoji", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsLegacyUnicodeEmoji(test.raw); got != test.want {
				t.Fatalf("IsLegacyUnicodeEmoji(%q) = %t, want %t", test.raw, got, test.want)
			}
		})
	}
}
