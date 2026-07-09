package ports

import "context"

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
