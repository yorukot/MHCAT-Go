package fakemongo

import (
	"context"
	"strings"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type RedeemRepository struct {
	mu       sync.Mutex
	Codes    map[string]domain.RedeemCode
	Balances map[string]float64
	Err      error
}

func NewRedeemRepository() *RedeemRepository {
	return &RedeemRepository{
		Codes:    map[string]domain.RedeemCode{},
		Balances: map[string]float64{},
	}
}

func (r *RedeemRepository) GetRedeemCode(ctx context.Context, code string) (domain.RedeemCode, error) {
	if err := r.ready(ctx); err != nil {
		return domain.RedeemCode{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	code = strings.TrimSpace(code)
	value, ok := r.Codes[code]
	if !ok {
		return domain.RedeemCode{}, ports.ErrRedeemCodeNotFound
	}
	return value, nil
}

func (r *RedeemRepository) ConsumeRedeemCode(ctx context.Context, command domain.RedeemCommand, price float64) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := command.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Codes[command.Code]; !ok {
		return ports.ErrRedeemCodeNotFound
	}
	delete(r.Codes, command.Code)
	r.Balances[command.GuildID] += price
	return nil
}

func (r *RedeemRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.RedeemRepository = (*RedeemRepository)(nil)
