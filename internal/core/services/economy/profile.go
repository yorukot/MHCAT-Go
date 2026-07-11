package economy

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ProfileService struct {
	Repository ports.EconomyProfileRepository
}

type ProfileQuery struct {
	GuildID string
	UserID  string
	Now     time.Time
}

type ProfileResult struct {
	GuildID string
	UserID  string

	CoinBalance domain.CoinBalance
	CoinFound   bool
	CoinRank    int

	Config      domain.EconomyConfig
	ConfigFound bool

	WorkConfig      domain.WorkConfig
	WorkConfigFound bool
	WorkUser        domain.WorkUserState
	WorkUserFound   bool

	TextXP      domain.XPProfile
	TextXPFound bool
	TextRank    int

	VoiceXP      domain.XPProfile
	VoiceXPFound bool
	VoiceRank    int

	SignStatus string
	NowUnix    int64
}

func (s ProfileService) Query(ctx context.Context, query ProfileQuery) (ProfileResult, error) {
	if err := ctx.Err(); err != nil {
		return ProfileResult{}, err
	}
	if s.Repository == nil {
		return ProfileResult{}, domain.ErrInvalidEconomyProfileQuery
	}
	query.GuildID = strings.TrimSpace(query.GuildID)
	query.UserID = strings.TrimSpace(query.UserID)
	if query.GuildID == "" || query.UserID == "" {
		return ProfileResult{}, domain.ErrInvalidEconomyProfileQuery
	}
	if query.Now.IsZero() {
		query.Now = time.Now()
	}
	result := ProfileResult{
		GuildID:    query.GuildID,
		UserID:     query.UserID,
		SignStatus: "沒有資料",
		NowUnix:    query.Now.Unix(),
	}

	balance, err := s.Repository.GetCoinBalance(ctx, query.GuildID, query.UserID)
	switch {
	case err == nil:
		result.CoinBalance = balance
		result.CoinFound = true
	case errors.Is(err, ports.ErrCoinBalanceNotFound):
	default:
		return ProfileResult{}, err
	}

	config, err := s.Repository.GetEconomyConfig(ctx, query.GuildID)
	switch {
	case err == nil:
		result.Config = config
		result.ConfigFound = true
	case errors.Is(err, ports.ErrEconomyConfigMissing):
		result.Config = config
	default:
		return ProfileResult{}, err
	}

	workConfig, err := s.Repository.GetWorkConfig(ctx, query.GuildID)
	switch {
	case err == nil:
		result.WorkConfig = workConfig
		result.WorkConfigFound = true
	case errors.Is(err, ports.ErrWorkConfigMissing):
	default:
		return ProfileResult{}, err
	}

	workUser, err := s.Repository.GetWorkUser(ctx, query.GuildID, query.UserID)
	switch {
	case err == nil:
		result.WorkUser = workUser
		result.WorkUserFound = true
	case errors.Is(err, ports.ErrWorkUserMissing):
	default:
		return ProfileResult{}, err
	}

	textXP, err := s.Repository.GetTextXPProfile(ctx, query.GuildID, query.UserID)
	switch {
	case err == nil:
		result.TextXP = textXP
		result.TextXPFound = true
	case errors.Is(err, ports.ErrTextXPProfileMissing):
	default:
		return ProfileResult{}, err
	}
	voiceXP, err := s.Repository.GetVoiceXPProfile(ctx, query.GuildID, query.UserID)
	switch {
	case err == nil:
		result.VoiceXP = voiceXP
		result.VoiceXPFound = true
	case errors.Is(err, ports.ErrVoiceXPProfileMissing):
	default:
		return ProfileResult{}, err
	}

	if result.CoinFound {
		balances, err := s.Repository.ListCoinBalances(ctx, query.GuildID)
		if err != nil {
			return ProfileResult{}, err
		}
		result.CoinRank = legacyCoinProfileRank(balances, result.CoinBalance)
	}
	if result.TextXPFound {
		profiles, err := s.Repository.ListTextXPProfiles(ctx, query.GuildID)
		if err != nil {
			return ProfileResult{}, err
		}
		result.TextRank = legacyXPProfileRank(profiles, result.TextXP, false)
	}
	if result.VoiceXPFound {
		profiles, err := s.Repository.ListVoiceXPProfiles(ctx, query.GuildID)
		if err != nil {
			return ProfileResult{}, err
		}
		result.VoiceRank = legacyXPProfileRank(profiles, result.VoiceXP, true)
	}
	result.SignStatus = legacyProfileSignStatus(result.CoinBalance, result.CoinFound, result.Config, result.ConfigFound, result.NowUnix)
	return result, nil
}

func legacyCoinProfileRank(balances []domain.CoinBalance, viewer domain.CoinBalance) int {
	viewerCoins, viewerNumeric := legacyCoinRankNumber(viewer)
	rank := 0
	for _, balance := range balances {
		coins, numeric := legacyCoinRankNumber(balance)
		if !numeric || !viewerNumeric || !(coins < viewerCoins) {
			rank++
		}
	}
	return rank
}

func legacyXPProfileRank(profiles []domain.XPProfile, viewer domain.XPProfile, voice bool) int {
	viewerTotal := legacyXPProfileTotal(viewer, voice, true)
	rank := 0
	for _, profile := range profiles {
		if legacyXPProfileTotal(profile, voice, false) >= viewerTotal {
			rank++
		}
	}
	return rank
}

func legacyXPProfileTotal(profile domain.XPProfile, voice bool, viewer bool) int64 {
	divisor := 3.0
	includeZero := true
	if voice {
		divisor = 2.0
	} else if viewer {
		includeZero = false
	}
	return legacyXPAccumulated(profile.Level, divisor, includeZero) + 100 + profile.XP
}

func legacyXPAccumulated(level int64, divisor float64, includeZero bool) int64 {
	total := int64(0)
	for y := level - 1; ; y-- {
		if includeZero {
			if y < 0 {
				break
			}
		} else if y <= 0 {
			break
		}
		total += int64(float64(y)*(float64(y)/divisor))*100 + 100
	}
	return total
}

func LegacyProfileXPRequired(level int64, voice bool) int64 {
	divisor := 3.0
	if voice {
		divisor = 2.0
	}
	return int64((float64(level) * (float64(level) / divisor) * 100) + 100)
}

func legacyProfileSignStatus(balance domain.CoinBalance, found bool, config domain.EconomyConfig, configFound bool, nowUnix int64) string {
	if !found {
		return "沒有資料"
	}
	switch balance.Today {
	case 1:
		return "已簽到"
	case 0:
		return "未簽到"
	default:
		cooldown := int64(86400)
		if configFound && config.ResetMarker > 0 {
			cooldown = config.ResetMarker
		}
		if nowUnix-balance.Today < cooldown {
			return "已簽到"
		}
		return "未簽到"
	}
}

func LegacyProfileAmount(value float64) string {
	switch {
	case math.IsInf(value, -1):
		return "-Infinity"
	case value >= 1_000_000_000:
		return legacyProfileScaled(value, 1_000_000_000, "G")
	case value >= 1_000_000:
		return legacyProfileScaled(value, 1_000_000, "M")
	case value >= 1_000:
		return legacyProfileScaled(value, 1_000, "K")
	default:
		if math.Trunc(value) == value {
			return strconv.FormatInt(int64(value), 10)
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
}

func LegacyProfileCoinAmount(balance domain.CoinBalance) string {
	value, numeric := legacyCoinRankNumber(balance)
	if !numeric {
		return "NaN"
	}
	return LegacyProfileAmount(value)
}

func legacyProfileScaled(value float64, divisor float64, suffix string) string {
	if math.IsInf(value, 1) {
		return "Infinity" + suffix
	}
	rounded := math.Floor((value/divisor)*10+0.5) / 10
	text := fmt.Sprintf("%.1f", rounded)
	text = strings.TrimSuffix(text, ".0")
	return text + suffix
}
