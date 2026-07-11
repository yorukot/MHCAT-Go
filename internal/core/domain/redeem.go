package domain

import (
	"errors"
	"math"
	"strings"
)

var ErrInvalidRedeemCode = errors.New("invalid redeem code")

type RedeemCode struct {
	Identity        any
	Code            string
	Price           float64
	CreatedAtMillis float64
}

type RedeemCommand struct {
	GuildID string
	Code    string
	NowMS   int64
}

func (c RedeemCode) Validate() error {
	if c.Code == "" || math.IsNaN(c.Price) {
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
