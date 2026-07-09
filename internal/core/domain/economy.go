package domain

import (
	"errors"
	"time"
)

var ErrInvalidEconomyQuery = errors.New("invalid economy query")
var ErrInvalidSignIn = errors.New("invalid sign in")
var ErrInvalidEconomySettings = errors.New("invalid economy settings")

type CoinBalance struct {
	GuildID string
	UserID  string
	Coins   int64
	Today   int64
}

type EconomyConfig struct {
	GuildID     string
	GachaCost   int64
	SignCoins   int64
	ChannelID   string
	XPMultiple  float64
	ResetMarker int64
}

func (c EconomyConfig) EffectiveGachaCost() int64 {
	if c.GachaCost <= 0 {
		return 500
	}
	return c.GachaCost
}

type SignCalendar struct {
	GuildID string
	UserID  string
	Date    map[string]map[string][]string
}

func (c SignCalendar) HasDay(year string, month string, day string) bool {
	months, ok := c.Date[year]
	if !ok {
		return false
	}
	days, ok := months[month]
	if !ok {
		return false
	}
	for _, signedDay := range days {
		if signedDay == day {
			return true
		}
	}
	return false
}

type SignInCommand struct {
	GuildID string
	UserID  string
	Now     time.Time
	Year    string
	Month   string
	Day     string
}

type SignInResult struct {
	Balance     CoinBalance
	Config      EconomyConfig
	Calendar    SignCalendar
	Reward      int64
	ConfigFound bool
	SignedAt    time.Time
}

type SignInListEntry struct {
	UserID       string
	SignedAtUnix int64
	ShowSignedAt bool
}

type SignInListResult struct {
	GuildID       string
	ActorUserID   string
	Entries       []SignInListEntry
	RollingWindow bool
}

type EconomySettingsCommand struct {
	GuildID           string
	GachaCost         int64
	SignCooldownHours int64
	SignCoins         int64
	NotificationID    string
	XPMultiple        float64
}
