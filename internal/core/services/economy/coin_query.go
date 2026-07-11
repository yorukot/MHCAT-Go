package economy

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const DefaultGachaCost int64 = 500

type CoinQueryService struct {
	Repository ports.EconomyQueryRepository
}

type CoinQueryResult struct {
	Balance          domain.CoinBalance
	Config           domain.EconomyConfig
	ConfigFound      bool
	GachaCost        int64
	GachaCostText    string
	MissingCoins     int64
	MissingCoinsText string
	CanGacha         bool
}

func (s CoinQueryService) Query(ctx context.Context, guildID string, userID string) (CoinQueryResult, error) {
	if s.Repository == nil {
		return CoinQueryResult{}, domain.ErrInvalidEconomyQuery
	}
	balance, err := s.Repository.GetCoinBalance(ctx, guildID, userID)
	if err != nil {
		return CoinQueryResult{}, err
	}
	config, err := s.Repository.GetEconomyConfig(ctx, guildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return CoinQueryResult{}, err
		}
		config = domain.EconomyConfig{GuildID: guildID, GachaCost: DefaultGachaCost}
		configFound = false
	}
	gachaCost := config.GachaCost
	gachaCostText := config.GachaCostText
	if !configFound {
		gachaCost = DefaultGachaCost
		gachaCostText = strconv.FormatInt(DefaultGachaCost, 10)
	} else if gachaCostText == "" {
		gachaCostText = strconv.FormatInt(gachaCost, 10)
	}
	missing := gachaCost - balance.Coins
	if missing < 0 {
		missing = 0
	}
	missingText, canGacha := legacyCoinDifference(gachaCostText, balance.Coins)
	return CoinQueryResult{
		Balance:          balance,
		Config:           config,
		ConfigFound:      configFound,
		GachaCost:        gachaCost,
		GachaCostText:    gachaCostText,
		MissingCoins:     missing,
		MissingCoinsText: missingText,
		CanGacha:         canGacha,
	}, nil
}

func legacyCoinDifference(gachaCost string, coins int64) (string, bool) {
	var cost float64
	switch gachaCost {
	case "null":
		cost = 0
	case "undefined", "NaN":
		return "", true
	default:
		parsed, err := strconv.ParseFloat(gachaCost, 64)
		if err != nil || math.IsNaN(parsed) {
			return "", true
		}
		cost = parsed
	}
	difference := cost - float64(coins)
	if !(difference > 0) {
		return "", true
	}
	if math.IsInf(difference, 1) {
		return "Infinity", false
	}
	return strconv.FormatFloat(difference, 'f', -1, 64), false
}
