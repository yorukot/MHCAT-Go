package domain

import (
	"errors"
	"strings"
)

var ErrInvalidBalanceQuery = errors.New("invalid balance query")

type Balance struct {
	GuildID string
	Amount  string
}

func (b Balance) Validate() error {
	if strings.TrimSpace(b.GuildID) == "" {
		return ErrInvalidBalanceQuery
	}
	return nil
}
