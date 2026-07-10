package domain

import (
	"errors"
	"testing"
)

func TestBirthdayConfigValidateAcceptsLegacyUTCChoices(t *testing.T) {
	config := BirthdayConfig{
		GuildID:   "guild-1",
		Message:   "{user} 生日快樂",
		UTCOffset: "+08:00",
		ChannelID: "channel-1",
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestBirthdayConfigValidateAcceptsWhitespaceMessage(t *testing.T) {
	config := BirthdayConfig{
		GuildID:   "guild-1",
		Message:   "   ",
		UTCOffset: "+08:00",
		ChannelID: "channel-1",
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate whitespace message: %v", err)
	}
}

func TestBirthdayConfigValidateRejectsInvalidUTC(t *testing.T) {
	config := BirthdayConfig{
		GuildID:   "guild-1",
		Message:   "{user} 生日快樂",
		UTCOffset: "-01:00",
		ChannelID: "channel-1",
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

func TestBirthdayProfileValidateIdentity(t *testing.T) {
	if err := (BirthdayProfile{GuildID: "guild-1", UserID: "user-1"}).ValidateIdentity(); err != nil {
		t.Fatalf("validate identity: %v", err)
	}
	if err := (BirthdayProfile{GuildID: "guild-1"}).ValidateIdentity(); !errors.Is(err, ErrInvalidBirthdayProfile) {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateBirthdayDatePreservesLegacyBounds(t *testing.T) {
	year := 2000
	if err := ValidateBirthdayDate(&year, 2, 29, 2026); err != nil {
		t.Fatalf("validate date: %v", err)
	}
	zeroYear := 0
	if err := ValidateBirthdayDate(&zeroYear, 2, 29, 2026); err != nil {
		t.Fatalf("legacy treats explicit year zero as unset during validation: %v", err)
	}
	for _, tc := range []struct {
		name  string
		year  *int
		month int
		day   int
		want  error
	}{
		{name: "old year", year: intPtr(1899), month: 1, day: 1, want: ErrInvalidBirthdayYear},
		{name: "future year", year: intPtr(2027), month: 1, day: 1, want: ErrInvalidBirthdayYear},
		{name: "month low", year: &year, month: 0, day: 1, want: ErrInvalidBirthdayMonth},
		{name: "month high", year: &year, month: 13, day: 1, want: ErrInvalidBirthdayMonth},
		{name: "day low", year: &year, month: 1, day: 0, want: ErrInvalidBirthdayDay},
		{name: "day high", year: &year, month: 4, day: 31, want: ErrInvalidBirthdayDay},
		{name: "feb high", year: &year, month: 2, day: 30, want: ErrInvalidBirthdayDay},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateBirthdayDate(tc.year, tc.month, tc.day, 2026); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestBirthdayProfileValidateDateTimeRequiresLegacyMinuteStep(t *testing.T) {
	year, month, day, hour, minute := 2001, 12, 31, 23, 55
	profile := BirthdayProfile{
		GuildID:       "guild-1",
		UserID:        "user-1",
		BirthdayYear:  &year,
		BirthdayMonth: &month,
		BirthdayDay:   &day,
		SendHour:      &hour,
		SendMinute:    &minute,
	}
	if err := profile.ValidateDateTime(); err != nil {
		t.Fatalf("validate date/time: %v", err)
	}
	badMinute := 56
	profile.SendMinute = &badMinute
	if err := profile.ValidateDateTime(); !errors.Is(err, ErrInvalidBirthdayTime) {
		t.Fatalf("err = %v", err)
	}
}

func intPtr(value int) *int {
	return &value
}
