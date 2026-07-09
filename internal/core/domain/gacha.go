package domain

import "errors"

var ErrInvalidGachaQuery = errors.New("invalid gacha query")
var ErrInvalidGachaPrize = errors.New("invalid gacha prize")

type GachaPrize struct {
	GuildID string
	Name    string
	Chance  float64
	Count   int64
}

type GachaPrizeConfig struct {
	GuildID    string
	Name       string
	Code       string
	Chance     float64
	AutoDelete bool
	Count      int64
	GiveCoin   int64
}

type GachaPrizeEdit struct {
	GuildID    string
	Name       string
	Code       string
	Chance     float64
	ChanceSet  bool
	AutoDelete bool
	Count      int64
	GiveCoin   int64
}

type GachaPrizePool struct {
	GuildID     string
	Prizes      []GachaPrize
	Config      EconomyConfig
	ConfigFound bool
}
