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
	ErrShopItemMissing      = errors.New("shop item not found")
	ErrShopItemExists       = errors.New("shop item already exists")
	ErrShopItemLimit        = errors.New("shop item limit exceeded")
	ErrShopInsufficientCoin = errors.New("shop purchase has insufficient coins")
	ErrShopQuantityInvalid  = errors.New("shop purchase quantity is invalid")
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

type EconomyCoinResetRepository interface {
	ResetCoinBalances(ctx context.Context, command domain.CoinResetCommand) (domain.CoinResetResult, error)
}

type EconomyRockPaperScissorsRepository interface {
	ApplyRockPaperScissors(ctx context.Context, command domain.RockPaperScissorsCommand) (domain.RockPaperScissorsResult, error)
}

type EconomyShopRepository interface {
	ListShopItems(ctx context.Context, guildID string) ([]domain.ShopItem, error)
	GetShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error)
	CreateShopItem(ctx context.Context, item domain.ShopItem) (domain.ShopItem, error)
	DeleteShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error)
	PurchaseShopItem(ctx context.Context, command domain.ShopPurchaseCommand) (domain.ShopPurchaseResult, error)
}
