package domain

import (
	"errors"
	"strings"
)

var ErrInvalidStatsConfigRequest = errors.New("invalid stats config request")

type StatsConfig struct {
	GuildID  string
	ParentID string
}

func (c StatsConfig) Normalize() StatsConfig {
	return StatsConfig{
		GuildID:  strings.TrimSpace(c.GuildID),
		ParentID: strings.TrimSpace(c.ParentID),
	}
}
