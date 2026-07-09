package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidEconomyQuery = errors.New("invalid economy query")
var ErrInvalidSignIn = errors.New("invalid sign in")
var ErrInvalidEconomySettings = errors.New("invalid economy settings")
var ErrInvalidCoinAdminCommand = errors.New("invalid coin admin command")
var ErrInvalidCoinRankQuery = errors.New("invalid coin rank query")
var ErrInvalidEconomyProfileQuery = errors.New("invalid economy profile query")

type CoinAdminOperation string

const (
	CoinAdminOperationAdd    CoinAdminOperation = "add"
	CoinAdminOperationReduce CoinAdminOperation = "reduce"
)

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

type CoinAdminCommand struct {
	GuildID   string
	UserID    string
	Operation CoinAdminOperation
	Amount    int64
}

type CoinAdminResult struct {
	Balance CoinBalance
	Delta   int64
	Created bool
}

func (c CoinAdminCommand) Normalize() CoinAdminCommand {
	return CoinAdminCommand{
		GuildID:   strings.TrimSpace(c.GuildID),
		UserID:    strings.TrimSpace(c.UserID),
		Operation: CoinAdminOperation(strings.TrimSpace(string(c.Operation))),
		Amount:    c.Amount,
	}
}

func (c CoinAdminCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.UserID == "" || c.Amount <= 0 {
		return ErrInvalidCoinAdminCommand
	}
	switch c.Operation {
	case CoinAdminOperationAdd, CoinAdminOperationReduce:
		return nil
	default:
		return ErrInvalidCoinAdminCommand
	}
}

func (c CoinAdminCommand) SignedDelta() (int64, error) {
	c = c.Normalize()
	if err := c.Validate(); err != nil {
		return 0, err
	}
	if c.Operation == CoinAdminOperationReduce {
		return -c.Amount, nil
	}
	return c.Amount, nil
}
