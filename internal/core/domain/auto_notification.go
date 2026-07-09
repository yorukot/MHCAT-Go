package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAutoNotificationSchedule = errors.New("invalid auto-notification schedule")

type AutoNotificationSchedule struct {
	GuildID   string
	ID        string
	Cron      string
	ChannelID string
	Pending   bool
}

func (s AutoNotificationSchedule) Normalized() AutoNotificationSchedule {
	s.GuildID = strings.TrimSpace(s.GuildID)
	s.ID = strings.TrimSpace(s.ID)
	s.Cron = strings.TrimSpace(s.Cron)
	s.ChannelID = strings.TrimSpace(s.ChannelID)
	return s
}

func ValidateAutoNotificationGuildID(guildID string) error {
	if strings.TrimSpace(guildID) == "" {
		return ErrInvalidAutoNotificationSchedule
	}
	return nil
}

func ValidateAutoNotificationDelete(guildID string, id string) error {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(id) == "" {
		return ErrInvalidAutoNotificationSchedule
	}
	return nil
}
