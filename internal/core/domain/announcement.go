package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAnnouncementConfig = errors.New("invalid announcement config")

type AnnouncementChannelConfig struct {
	GuildID   string
	ChannelID string
}

type BoundAnnouncementConfig struct {
	GuildID   string
	ChannelID string
	Tag       string
	Color     string
	Title     string
}

func (c AnnouncementChannelConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidAnnouncementConfig
	}
	return nil
}

func (c BoundAnnouncementConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidAnnouncementConfig
	}
	if c.Tag == "" || c.Color == "" || c.Title == "" {
		return ErrInvalidAnnouncementConfig
	}
	if !ValidLegacyBoundAnnouncementColor(c.Color) {
		return ErrInvalidAnnouncementConfig
	}
	return nil
}

func ValidLegacyBoundAnnouncementColor(value string) bool {
	return value == "Random" || ValidLegacyXPColor(value)
}

func ParseLegacyAnnouncementSendColor(value string) (int, bool) {
	if !validLegacyAnnouncementModalColor(value) {
		return 0, false
	}
	return ParseLegacyColorValue(value)
}

func validLegacyAnnouncementModalColor(value string) bool {
	switch strings.ToLower(value) {
	case "currentcolor", "inherit", "transparent":
		return false
	default:
		return ValidLegacyXPColor(value)
	}
}
