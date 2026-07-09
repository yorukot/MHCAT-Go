package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Options struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

type Client struct {
	opts   Options
	client *drivermongo.Client
}

func NewClient(opts Options) (*Client, error) {
	if opts.URI == "" {
		return nil, errors.New("mongo uri is required")
	}
	if opts.Database == "" {
		return nil, errors.New("mongo database is required")
	}
	if opts.ConnectTimeout <= 0 {
		return nil, errors.New("mongo connect timeout must be positive")
	}
	if opts.PingTimeout <= 0 {
		return nil, errors.New("mongo ping timeout must be positive")
	}
	return &Client{opts: opts}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	connectCtx, cancel := context.WithTimeout(ctx, c.opts.ConnectTimeout)
	defer cancel()

	client, err := drivermongo.Connect(options.Client().ApplyURI(c.opts.URI))
	if err != nil {
		return fmt.Errorf("create mongo client: %w", err)
	}
	c.client = client
	if err := c.Ping(connectCtx); err != nil {
		_ = c.Disconnect(context.Background())
		return err
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return errors.New("mongo client is not connected")
	}
	pingCtx, cancel := context.WithTimeout(ctx, c.opts.PingTimeout)
	defer cancel()

	if err := c.client.Ping(pingCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}
	return nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect mongo: %w", err)
	}
	c.client = nil
	return nil
}

func (c *Client) DatabaseName() string {
	return c.opts.Database
}
