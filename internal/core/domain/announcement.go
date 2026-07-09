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
	if strings.TrimSpace(c.Tag) == "" || strings.TrimSpace(c.Color) == "" || strings.TrimSpace(c.Title) == "" {
		return ErrInvalidAnnouncementConfig
	}
	if !ValidLegacyColor(c.Color) && strings.TrimSpace(c.Color) != "Random" {
		return ErrInvalidAnnouncementConfig
	}
	return nil
}
