package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrRedeemCodeNotFound = errors.New("redeem code not found")
	ErrRedeemCodeExpired  = errors.New("redeem code expired")
)

type RedeemRepository interface {
	GetRedeemCode(ctx context.Context, code string) (domain.RedeemCode, error)
	ConsumeRedeemCode(ctx context.Context, command domain.RedeemCommand, code domain.RedeemCode) error
}
