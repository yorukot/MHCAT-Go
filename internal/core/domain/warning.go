package domain

import (
	"errors"
	"strings"
)

var ErrInvalidWarningQuery = errors.New("invalid warning query")

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

func (h WarningHistory) ValidateQuery() error {
	if strings.TrimSpace(h.GuildID) == "" || strings.TrimSpace(h.UserID) == "" {
		return ErrInvalidWarningQuery
	}
	return nil
}
