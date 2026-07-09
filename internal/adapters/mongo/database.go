package mongo

import (
	"errors"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func (c *Client) Database() (*drivermongo.Database, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("mongo client is not connected")
	}
	return c.client.Database(c.opts.Database), nil
}
