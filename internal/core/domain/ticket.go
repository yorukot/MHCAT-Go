package domain

import (
	"errors"
	"strings"
)

var ErrInvalidTicketConfig = errors.New("invalid ticket config")

type TicketConfig struct {
	GuildID        string
	CategoryID     string
	AdminRoleID    string
	EveryoneRoleID string
}

func (c TicketConfig) ValidateForWrite() error {
	if strings.TrimSpace(c.GuildID) == "" ||
		strings.TrimSpace(c.CategoryID) == "" ||
		strings.TrimSpace(c.AdminRoleID) == "" ||
		strings.TrimSpace(c.EveryoneRoleID) == "" {
		return ErrInvalidTicketConfig
	}
	return nil
}
