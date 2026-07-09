package mongo

import (
	"context"
	"errors"
	"fmt"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

type TransactionRunner struct {
	client *Client
}

func NewTransactionRunner(client *Client) (*TransactionRunner, error) {
	if client == nil || client.client == nil {
		return nil, errors.New("connected mongo client is required")
	}
	return &TransactionRunner{client: client}, nil
}

func (r *TransactionRunner) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("transaction callback is required")
	}
	return MapError(r.client.client.UseSession(ctx, func(sessionCtx context.Context) error {
		session := drivermongo.SessionFromContext(sessionCtx)
		if session == nil {
			return errors.New("mongo session is not available")
		}
		_, err := session.WithTransaction(sessionCtx, func(txCtx context.Context) (any, error) {
			if err := txCtx.Err(); err != nil {
				return nil, err
			}
			return nil, fn(txCtx)
		})
		if err != nil {
			return fmt.Errorf("run mongo transaction: %w", err)
		}
		return nil
	}))
}
