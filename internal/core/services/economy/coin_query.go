package economy

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"

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
	BalanceText      string
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
	balanceText := balance.CoinsText
	if balanceText == "" {
		balanceText = strconv.FormatInt(balance.Coins, 10)
	}
	missingText, canGacha := legacyCoinDifference(gachaCostText, balanceText)
	return CoinQueryResult{
		Balance:          balance,
		Config:           config,
		ConfigFound:      configFound,
		GachaCost:        gachaCost,
		GachaCostText:    gachaCostText,
		BalanceText:      balanceText,
		MissingCoins:     missing,
		MissingCoinsText: missingText,
		CanGacha:         canGacha,
	}, nil
}

func legacyCoinDifference(gachaCost string, coins string) (string, bool) {
	cost, ok := legacyDisplayedNumber(gachaCost)
	if !ok {
		return "", true
	}
	balance, ok := legacyDisplayedNumber(coins)
	if !ok {
		return "", true
	}
	difference := cost - balance
	if !(difference > 0) {
		return "", true
	}
	if math.IsInf(difference, 1) {
		return "Infinity", false
	}
	return strconv.FormatFloat(difference, 'f', -1, 64), false
}

func legacyDisplayedNumber(value string) (float64, bool) {
	switch value {
	case "null":
		return 0, true
	case "undefined", "NaN":
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil && !math.IsNaN(parsed)
}

// LegacyEconomyNumber exposes Mongoose/JavaScript number coercion to economy adapters.
func LegacyEconomyNumber(value string) (float64, bool) {
	return legacyDisplayedNumber(value)
}

// LegacyEconomyNumberText formats a number like JavaScript String(number).
func LegacyEconomyNumberText(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	}
	absolute := math.Abs(value)
	if absolute >= 1e-6 && absolute < 1e21 {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(value, 'e', -1, 64), "e")
	parsedExponent, _ := strconv.Atoi(exponent)
	if parsedExponent >= 0 {
		return mantissa + "e+" + strconv.Itoa(parsedExponent)
	}
	return mantissa + "e" + strconv.Itoa(parsedExponent)
}
