package domain

import (
	"errors"
	"strings"
)

var ErrInvalidWarningQuery = errors.New("invalid warning query")
var ErrInvalidWarningSettings = errors.New("invalid warning settings")
var ErrInvalidWarningRemoval = errors.New("invalid warning removal")
var ErrInvalidWarningIssue = errors.New("invalid warning issue")

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
	Threshold float64
	Action    string
}

type WarningIssue struct {
	GuildID     string
	UserID      string
	ModeratorID string
	Reason      string
	Time        string
}

type WarningIssueResult struct {
	History WarningHistory
	Created bool
}

type WarningRemoval struct {
	GuildID string
	UserID  string
	Index   int64
}

func (h WarningHistory) ValidateQuery() error {
	if strings.TrimSpace(h.GuildID) == "" || strings.TrimSpace(h.UserID) == "" {
		return ErrInvalidWarningQuery
	}
	return nil
}

func (s WarningSettings) Validate() error {
	if strings.TrimSpace(s.GuildID) == "" {
		return ErrInvalidWarningSettings
	}
	switch strings.TrimSpace(s.Action) {
	case WarningSettingsActionBan, WarningSettingsActionKick:
		return nil
	default:
		return ErrInvalidWarningSettings
	}
}

func (i WarningIssue) Validate() error {
	if strings.TrimSpace(i.GuildID) == "" ||
		strings.TrimSpace(i.UserID) == "" ||
		strings.TrimSpace(i.ModeratorID) == "" ||
		i.Reason == "" ||
		strings.TrimSpace(i.Time) == "" {
		return ErrInvalidWarningIssue
	}
	return nil
}

func (r WarningRemoval) ValidateSingle() error {
	if strings.TrimSpace(r.GuildID) == "" || strings.TrimSpace(r.UserID) == "" {
		return ErrInvalidWarningRemoval
	}
	return nil
}

func (r WarningRemoval) ValidateAll() error {
	if strings.TrimSpace(r.GuildID) == "" || strings.TrimSpace(r.UserID) == "" {
		return ErrInvalidWarningRemoval
	}
	return nil
}
