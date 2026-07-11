package domain

import (
	"errors"
	"strings"
)

var ErrInvalidRedeemCode = errors.New("invalid redeem code")

type RedeemCode struct {
	Code            string
	Price           float64
	CreatedAtMillis int64
}

type RedeemCommand struct {
	GuildID string
	Code    string
	NowMS   int64
}

func (c RedeemCode) Validate() error {
	if c.Code == "" || c.Price < 0 || c.CreatedAtMillis <= 0 {
		return ErrInvalidRedeemCode
	}
	return nil
}

func (c RedeemCommand) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || c.Code == "" || c.NowMS <= 0 {
		return ErrInvalidRedeemCode
	}
	return nil
}
