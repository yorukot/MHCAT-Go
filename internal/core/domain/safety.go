package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAntiScamConfig = errors.New("invalid anti-scam config")

type AntiScamConfig struct {
	GuildID string
	Open    bool
}

func (c AntiScamConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" {
		return ErrInvalidAntiScamConfig
	}
	return nil
}
