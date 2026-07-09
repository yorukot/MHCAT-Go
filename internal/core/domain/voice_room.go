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
	GuildID        string
	ChannelID      string
	Password       string
	OwnerID        string
	TextChannelID  string
	AllowedUserIDs []string
}

func (c VoiceRoomConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" ||
		strings.TrimSpace(c.TriggerChannelID) == "" ||
		strings.TrimSpace(c.Name) == "" {
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
	l.Password = strings.TrimSpace(l.Password)
	l.OwnerID = strings.TrimSpace(l.OwnerID)
	l.TextChannelID = strings.TrimSpace(l.TextChannelID)
	out := make([]string, 0, len(l.AllowedUserIDs))
	for _, userID := range l.AllowedUserIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			out = append(out, userID)
		}
	}
	l.AllowedUserIDs = out
	return l
}

func (l VoiceRoomLock) Validate() error {
	l = l.Normalize()
	if l.GuildID == "" || l.ChannelID == "" || l.OwnerID == "" || l.TextChannelID == "" {
		return ErrInvalidVoiceRoomLock
	}
	return nil
}
