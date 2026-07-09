package fakemongo

import "context"

type TransactionRunner struct {
	Committed  bool
	RolledBack bool
	Calls      int
}

func (r *TransactionRunner) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := ctx.Err(); err != nil {
		r.RolledBack = true
		return err
	}
	r.Calls++
	if err := fn(ctx); err != nil {
		r.RolledBack = true
		return err
	}
	r.Committed = true
	return nil
}
