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
	valid.Color = "Random"
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate random: %v", err)
	}
	valid.Color = "not-a-color"
	if err := valid.Validate(); !errors.Is(err, ErrInvalidAnnouncementConfig) {
		t.Fatalf("expected invalid color, got %v", err)
	}
}
