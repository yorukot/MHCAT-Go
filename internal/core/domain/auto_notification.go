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
	Message   AutoNotificationMessage
	Pending   bool
}

type AutoNotificationSetupDraft struct {
	GuildID   string
	ID        string
	ChannelID string
}

type AutoNotificationSetup struct {
	GuildID string
	ID      string
	Cron    string
	Message AutoNotificationMessage
}

type AutoNotificationMessage struct {
	Content          string
	EmbedTitle       string
	EmbedDescription string
	EmbedColor       string
}

func (s AutoNotificationSchedule) Normalized() AutoNotificationSchedule {
	s.GuildID = strings.TrimSpace(s.GuildID)
	s.ID = strings.TrimSpace(s.ID)
	s.Cron = strings.TrimSpace(s.Cron)
	s.ChannelID = strings.TrimSpace(s.ChannelID)
	s.Message = s.Message.Normalized()
	return s
}

func (d AutoNotificationSetupDraft) Normalized() AutoNotificationSetupDraft {
	d.GuildID = strings.TrimSpace(d.GuildID)
	d.ID = strings.TrimSpace(d.ID)
	d.ChannelID = strings.TrimSpace(d.ChannelID)
	return d
}

func (s AutoNotificationSetup) Normalized() AutoNotificationSetup {
	s.GuildID = strings.TrimSpace(s.GuildID)
	s.ID = strings.TrimSpace(s.ID)
	s.Cron = strings.TrimSpace(s.Cron)
	s.Message = s.Message.Normalized()
	return s
}

func (m AutoNotificationMessage) Normalized() AutoNotificationMessage {
	return m
}

func (m AutoNotificationMessage) HasEmbed() bool {
	return m.EmbedTitle != "" || m.EmbedDescription != ""
}

func (m AutoNotificationMessage) Empty() bool {
	return m.Content == "" && m.EmbedTitle == "" && m.EmbedDescription == ""
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

func ValidateAutoNotificationSetupDraft(draft AutoNotificationSetupDraft) error {
	draft = draft.Normalized()
	if draft.GuildID == "" || draft.ID == "" || draft.ChannelID == "" {
		return ErrInvalidAutoNotificationSchedule
	}
	return nil
}

func ValidateAutoNotificationSetup(setup AutoNotificationSetup) error {
	setup = setup.Normalized()
	if setup.GuildID == "" || setup.ID == "" || setup.Cron == "" || setup.Message.Empty() {
		return ErrInvalidAutoNotificationSchedule
	}
	return nil
}

func ValidateAutoNotificationDelivery(schedule AutoNotificationSchedule) error {
	schedule = schedule.Normalized()
	if schedule.Pending || schedule.GuildID == "" || schedule.ID == "" || schedule.Cron == "" || schedule.ChannelID == "" || schedule.Message.Empty() {
		return ErrInvalidAutoNotificationSchedule
	}
	return nil
}
