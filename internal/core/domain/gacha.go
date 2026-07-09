package domain

import "errors"

var ErrInvalidGachaQuery = errors.New("invalid gacha query")

type GachaPrize struct {
	GuildID string
	Name    string
	Chance  float64
	Count   int64
}

type GachaPrizePool struct {
	GuildID     string
	Prizes      []GachaPrize
	Config      EconomyConfig
	ConfigFound bool
}
