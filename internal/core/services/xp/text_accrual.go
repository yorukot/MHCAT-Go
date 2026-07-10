package xp

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TextAccrualService struct {
	Repository       ports.TextXPAccrualRepository
	RandomMultiplier func() int64
}

type TextAccrualResult struct {
	Profile domain.XPProfile
	Gained  int64
	Leveled bool
}

func (s TextAccrualService) AccrueMessage(ctx context.Context, guildID string, userID string, content string) (TextAccrualResult, error) {
	if s.Repository == nil {
		return TextAccrualResult{}, domain.ErrInvalidXPAdjustment
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return TextAccrualResult{}, domain.ErrInvalidXPAdjustment
	}
	gained := LegacyTextXPForMessage(content, s.multiplier())
	profile, err := s.Repository.GetTextXPProfile(ctx, guildID, userID)
	if err != nil {
		if !errors.Is(err, ports.ErrTextXPProfileMissing) {
			return TextAccrualResult{}, err
		}
		profile = domain.XPProfile{GuildID: guildID, UserID: userID}
	}
	profile, leveled := domain.ApplyTextXPMessage(profile, gained)
	if err := profile.Validate(); err != nil {
		return TextAccrualResult{}, err
	}
	if err := s.Repository.SaveTextXPProfile(ctx, profile); err != nil {
		return TextAccrualResult{}, err
	}
	return TextAccrualResult{Profile: profile, Gained: gained, Leveled: leveled}, ctx.Err()
}

func (s TextAccrualService) multiplier() int64 {
	if s.RandomMultiplier != nil {
		return s.RandomMultiplier()
	}
	value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(701))
	if err != nil {
		return 100
	}
	return 100 + value.Int64()
}

func LegacyTextXPForMessage(content string, multiplierMilli int64) int64 {
	if multiplierMilli < 100 {
		multiplierMilli = 100
	}
	if multiplierMilli > 800 {
		multiplierMilli = 800
	}
	length := LegacyTextXPContentLength(content)
	if length > 50 {
		length = 50
	}
	return int64(math.Round(float64(length*2*multiplierMilli) / 1000))
}

func LegacyTextXPContentLength(content string) int64 {
	total := int64(0)
	for _, r := range content {
		if r <= 0xff {
			total++
			continue
		}
		total += int64(len(utf16.Encode([]rune{r})) * 2)
	}
	return total
}
