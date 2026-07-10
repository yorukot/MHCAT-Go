package domain

import (
	"errors"
	"strings"
)

var ErrInvalidVoiceRoomConfig = errors.New("invalid voice room config")
var ErrInvalidVoiceRoomLock = errors.New("invalid voice room lock")

type VoiceRoomConfig struct {
	GuildID          string
	TriggerChannelID string
	ParentID         string
	Name             string
	Limit            int
	Lock             bool
}

type VoiceRoomLock struct {
	GuildID         string
	ChannelID       string
	Password        string
	PasswordPresent bool
	OwnerID         string
	TextChannelID   string
	AllowedUserIDs  []string
}

type VoiceRoomState struct {
	GuildID   string
	ChannelID string
}

func (c VoiceRoomConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" ||
		strings.TrimSpace(c.TriggerChannelID) == "" ||
		c.Name == "" {
		return ErrInvalidVoiceRoomConfig
	}
	if c.Limit < 0 || c.Limit > 99 {
		return ErrInvalidVoiceRoomConfig
	}
	return nil
}

func (l VoiceRoomLock) Normalize() VoiceRoomLock {
	l.GuildID = strings.TrimSpace(l.GuildID)
	l.ChannelID = strings.TrimSpace(l.ChannelID)
	l.PasswordPresent = l.PasswordPresent || l.Password != ""
	l.AllowedUserIDs = append([]string(nil), l.AllowedUserIDs...)
	return l
}

func (l VoiceRoomLock) HasPassword() bool {
	return l.PasswordPresent || l.Password != ""
}

func (l VoiceRoomLock) Validate() error {
	l = l.Normalize()
	if l.GuildID == "" || l.ChannelID == "" || l.OwnerID == "" {
		return ErrInvalidVoiceRoomLock
	}
	return nil
}

func (s VoiceRoomState) Validate() error {
	if strings.TrimSpace(s.GuildID) == "" || strings.TrimSpace(s.ChannelID) == "" {
		return ErrInvalidVoiceRoomConfig
	}
	return nil
}
