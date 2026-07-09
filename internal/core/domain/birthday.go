package domain

import (
	"errors"
	"strings"
)

var ErrInvalidBirthdayConfig = errors.New("invalid birthday config")

type BirthdayConfig struct {
	GuildID                    string
	Message                    string
	UTCOffset                  string
	ChannelID                  string
	EveryoneCanSetBirthdayDate bool
	RoleID                     string
}

func (c BirthdayConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" ||
		strings.TrimSpace(c.Message) == "" ||
		strings.TrimSpace(c.UTCOffset) == "" ||
		strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidBirthdayConfig
	}
	if !validLegacyBirthdayUTCOffset(c.UTCOffset) {
		return ErrInvalidBirthdayConfig
	}
	return nil
}

func validLegacyBirthdayUTCOffset(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != len("+00:00") || value[0] != '+' || value[3:] != ":00" {
		return false
	}
	switch value[1:3] {
	case "00", "01", "02", "03", "04", "05", "06", "07", "08", "09",
		"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
		"20", "21", "22", "23":
		return true
	default:
		return false
	}
}
