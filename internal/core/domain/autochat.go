package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAutoChatConfig = errors.New("invalid autochat config")

type AutoChatConfig struct {
	GuildID   string
	ChannelID string
}

type AutoChatFallbackReply struct {
	Content        string
	UseTypingDelay bool
}

func (c AutoChatConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidAutoChatConfig
	}
	return nil
}
