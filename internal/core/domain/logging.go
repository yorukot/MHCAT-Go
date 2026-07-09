package domain

import (
	"errors"
	"strings"
)

var ErrInvalidLoggingConfig = errors.New("invalid logging config")

type LoggingConfig struct {
	GuildID           string
	ChannelID         string
	MessageUpdate     bool
	MessageDelete     bool
	ChannelUpdate     bool
	MemberVoiceUpdate bool
}

func (c LoggingConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidLoggingConfig
	}
	return nil
}
