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
}

type EconomySettingsRepository interface {
	SaveEconomyConfig(ctx context.Context, config domain.EconomyConfig) (domain.EconomyConfig, error)
}
