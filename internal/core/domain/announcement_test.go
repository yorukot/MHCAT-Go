package domain

import (
	"errors"
	"testing"
)

func TestAnnouncementChannelConfigValidate(t *testing.T) {
	if err := (AnnouncementChannelConfig{GuildID: "guild", ChannelID: "channel"}).Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if err := (AnnouncementChannelConfig{GuildID: "", ChannelID: "channel"}).Validate(); !errors.Is(err, ErrInvalidAnnouncementConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestBoundAnnouncementConfigValidate(t *testing.T) {
	valid := BoundAnnouncementConfig{
		GuildID:   "guild",
		ChannelID: "channel",
		Tag:       "@here",
		Color:     "#00ff19",
		Title:     "公告",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	for _, color := range []string{"Random", "#fff", "AliceBlue", "rgb(0 0 0)"} {
		valid.Color = color
		if err := valid.Validate(); err != nil {
			t.Fatalf("validate color %q: %v", color, err)
		}
	}
	for _, color := range []string{"RANDOM", "53FF53", " #53FF53", "not-a-color"} {
		valid.Color = color
		if err := valid.Validate(); !errors.Is(err, ErrInvalidAnnouncementConfig) {
			t.Fatalf("expected invalid color %q, got %v", color, err)
		}
	}
}

func TestBoundAnnouncementConfigPreservesNonemptyWhitespaceFields(t *testing.T) {
	config := BoundAnnouncementConfig{GuildID: "guild", ChannelID: "channel", Tag: " ", Color: "#53FF53", Title: "\t"}
	if err := config.Validate(); err != nil {
		t.Fatalf("legacy required options accept nonempty whitespace: %v", err)
	}
}

func TestParseLegacyAnnouncementSendColorMatchesValidatorAndDiscordIntersection(t *testing.T) {
	for value, want := range map[string]int{"#53FF53": 0x53FF53, "Red": 0xED4245, "Green": 0x57F287} {
		got, ok := ParseLegacyAnnouncementSendColor(value)
		if !ok || got != want {
			t.Fatalf("color %q = %x/%t, want %x", value, got, ok, want)
		}
	}
	for _, value := range []string{"53FF53", "Random", "#fff", "red", "AliceBlue", "transparent"} {
		if _, ok := ParseLegacyAnnouncementSendColor(value); ok {
			t.Fatalf("legacy modal/discord.js combination should reject %q", value)
		}
	}
}
