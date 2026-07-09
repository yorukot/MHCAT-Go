package domain

import (
	"errors"
	"testing"
)

func TestAutoChatConfigValidateRequiresGuildAndChannel(t *testing.T) {
	if err := (AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}).Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if err := (AutoChatConfig{GuildID: "guild-1"}).Validate(); !errors.Is(err, ErrInvalidAutoChatConfig) {
		t.Fatalf("missing channel err = %v", err)
	}
	if err := (AutoChatConfig{ChannelID: "channel-1"}).Validate(); !errors.Is(err, ErrInvalidAutoChatConfig) {
		t.Fatalf("missing guild err = %v", err)
	}
}
