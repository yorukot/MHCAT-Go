package domain

import "errors"

var ErrInvalidGachaQuery = errors.New("invalid gacha query")
var ErrInvalidGachaPrize = errors.New("invalid gacha prize")
var ErrInvalidGachaDraw = errors.New("invalid gacha draw")

const GachaAirPrizeName = "空氣QQ<:peepoHugMilk:994650902050906234>別氣餒，下一次定是你!!"

const legacyGachaMinimumAirWeight = 0.000000000000000000000000000000000000000000000000000000000000001

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

type GachaDrawCommand struct {
	GuildID string
	UserID  string
	Choice  string
}

type GachaDrawRequest struct {
	GuildID      string
	UserID       string
	PaidDraws    int
	ActualDraws  int
	RandomValues []float64
}

type GachaDrawPrizeResult struct {
	Name     string
	Code     string
	GiveCoin int64
	Air      bool
}

type GachaDrawResult struct {
	GuildID               string
	UserID                string
	PaidDraws             int
	ActualDraws           int
	Cost                  int64
	BalanceBefore         int64
	BalanceAfter          int64
	Config                EconomyConfig
	ConfigFound           bool
	NotificationChannelID string
	Prizes                []GachaDrawPrizeResult
}

func (r GachaDrawResult) PrizeNames() []string {
	names := make([]string, 0, len(r.Prizes))
	for _, prize := range r.Prizes {
		names = append(names, prize.Name)
	}
	return names
}

func (r GachaDrawResult) NonAirPrizes() []GachaDrawPrizeResult {
	prizes := make([]GachaDrawPrizeResult, 0, len(r.Prizes))
	for _, prize := range r.Prizes {
		if !prize.Air {
			prizes = append(prizes, prize)
		}
	}
	return prizes
}

func (r GachaDrawResult) CodePrizes() []GachaDrawPrizeResult {
	prizes := make([]GachaDrawPrizeResult, 0, len(r.Prizes))
	for _, prize := range r.Prizes {
		if !prize.Air && prize.Code != "" {
			prizes = append(prizes, prize)
		}
	}
	return prizes
}

func ResolveGachaDraw(prizes []GachaPrizeConfig, randomValue float64) GachaDrawPrizeResult {
	positiveChanceTotal := 0.0
	for _, prize := range prizes {
		if prize.Chance > 0 {
			positiveChanceTotal += prize.Chance
		}
	}
	airWeight := 100 - positiveChanceTotal
	if airWeight <= 0 {
		airWeight = legacyGachaMinimumAirWeight
	}
	totalWeight := positiveChanceTotal + airWeight
	threshold := normalizedGachaRandom(randomValue) * totalWeight
	running := 0.0
	for _, prize := range prizes {
		if prize.Chance <= 0 {
			continue
		}
		running += prize.Chance
		if threshold < running {
			return GachaDrawPrizeResult{
				Name:     prize.Name,
				Code:     prize.Code,
				GiveCoin: prize.GiveCoin,
			}
		}
	}
	return GachaDrawPrizeResult{Name: GachaAirPrizeName, Air: true}
}

func normalizedGachaRandom(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return 0.9999999999999999
	}
	return value
}
