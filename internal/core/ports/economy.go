package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrCoinBalanceNotFound  = errors.New("coin balance not found")
	ErrEconomyConfigMissing = errors.New("economy config not found")
	ErrAlreadySignedIn      = errors.New("already signed in")
	ErrCoinLimitExceeded    = errors.New("coin limit exceeded")
	ErrCoinNegativeBalance  = errors.New("coin balance would be negative")
)

type EconomyQueryRepository interface {
	GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error)
	GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error)
}

type EconomyRepository = EconomyQueryRepository

type EconomySignInRepository interface {
	EconomyQueryRepository
	SignIn(ctx context.Context, command domain.SignInCommand) (domain.SignInResult, error)
	GetSignCalendar(ctx context.Context, guildID string, userID string, year string, month string) (domain.SignCalendar, error)
	ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error)
}

type EconomyCoinRankRepository interface {
	ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error)
}

type EconomyProfileRepository interface {
	GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error)
	GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error)
	ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error)
	GetWorkConfig(ctx context.Context, guildID string) (domain.WorkConfig, error)
	GetWorkUser(ctx context.Context, guildID string, userID string) (domain.WorkUserState, error)
	GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error)
	GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error)
}

type EconomySettingsRepository interface {
	SaveEconomyConfig(ctx context.Context, config domain.EconomyConfig) (domain.EconomyConfig, error)
}

type EconomyCoinAdminRepository interface {
	AdjustCoinBalance(ctx context.Context, command domain.CoinAdminCommand) (domain.CoinAdminResult, error)
}

type EconomyRockPaperScissorsRepository interface {
	ApplyRockPaperScissors(ctx context.Context, command domain.RockPaperScissorsCommand) (domain.RockPaperScissorsResult, error)
}
