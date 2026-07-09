package domain

import (
	"errors"
	"strings"
)

var ErrInvalidVoiceRoomConfig = errors.New("invalid voice room config")

type VoiceRoomConfig struct {
	GuildID          string
	TriggerChannelID string
	ParentID         string
	Name             string
	Limit            int
	Lock             bool
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
