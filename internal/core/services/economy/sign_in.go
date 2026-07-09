package economy

import (
	"context"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	DefaultSignCoins       int64 = 25
	MaxLegacyCoinBalance   int64 = 999999999
	DefaultSignCooldownSec int64 = 86400
)

type SignInService struct {
	Repository ports.EconomySignInRepository
	Location   *time.Location
}

func (s SignInService) SignIn(ctx context.Context, guildID string, userID string, now time.Time) (domain.SignInResult, error) {
	if s.Repository == nil {
		return domain.SignInResult{}, domain.ErrInvalidSignIn
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.SignInResult{}, domain.ErrInvalidSignIn
	}
	local := now.In(s.location())
	command := domain.SignInCommand{
		GuildID: guildID,
		UserID:  userID,
		Now:     now,
		Year:    local.Format("2006"),
		Month:   local.Format("01"),
		Day:     strings.TrimLeft(local.Format("02"), "0"),
	}
	if command.Day == "" {
		command.Day = "0"
	}
	return s.Repository.SignIn(ctx, command)
}

func (s SignInService) Calendar(ctx context.Context, guildID string, userID string, year string, month string) (domain.SignCalendar, error) {
	if s.Repository == nil {
		return domain.SignCalendar{}, domain.ErrInvalidSignIn
	}
	return s.Repository.GetSignCalendar(ctx, guildID, userID, year, month)
}

func (s SignInService) location() *time.Location {
	if s.Location != nil {
		return s.Location
	}
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.FixedZone("Asia/Taipei", 8*60*60)
	}
	return location
}
