package domain

import (
	"errors"
	"testing"
)

func TestBirthdayConfigValidateAcceptsLegacyUTCChoices(t *testing.T) {
	config := BirthdayConfig{
		GuildID:    "guild-1",
		Message:    "{user} 生日快樂",
		UTCOffset:  "+08:00",
		ChannelID:  "channel-1",
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestBirthdayConfigValidateRejectsInvalidUTC(t *testing.T) {
	config := BirthdayConfig{
		GuildID:    "guild-1",
		Message:    "{user} 生日快樂",
		UTCOffset:  "-01:00",
		ChannelID:  "channel-1",
	}
	if err := config.Validate(); !errors.Is(err, ErrInvalidBirthdayConfig) {
		t.Fatalf("err = %v", err)
	}
}

func TestBirthdayConfigValidateRequiresFields(t *testing.T) {
	if err := (BirthdayConfig{GuildID: "guild-1", UTCOffset: "+08:00", ChannelID: "channel-1"}).Validate(); !errors.Is(err, ErrInvalidBirthdayConfig) {
		t.Fatalf("err = %v", err)
	}
}
