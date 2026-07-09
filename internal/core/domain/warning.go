package domain

import (
	"errors"
	"strings"
)

var ErrInvalidWarningQuery = errors.New("invalid warning query")
var ErrInvalidWarningSettings = errors.New("invalid warning settings")

const (
	WarningSettingsActionBan  = "停權"
	WarningSettingsActionKick = "踢出"
)

type WarningEntry struct {
	ModeratorID string
	Reason      string
	Time        string
}

type WarningHistory struct {
	GuildID string
	UserID  string
	Entries []WarningEntry
}

type WarningSettings struct {
	GuildID   string
	Threshold int64
	Action    string
}

func (h WarningHistory) ValidateQuery() error {
	if strings.TrimSpace(h.GuildID) == "" || strings.TrimSpace(h.UserID) == "" {
		return ErrInvalidWarningQuery
	}
	return nil
}

func (s WarningSettings) Validate() error {
	if strings.TrimSpace(s.GuildID) == "" || s.Threshold <= 0 {
		return ErrInvalidWarningSettings
	}
	switch strings.TrimSpace(s.Action) {
	case WarningSettingsActionBan, WarningSettingsActionKick:
		return nil
	default:
		return ErrInvalidWarningSettings
	}
}
