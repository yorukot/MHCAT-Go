package mongo

import (
	"errors"
	"strings"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func (c *Client) Collection(name string) (*drivermongo.Collection, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("mongo collection name is required")
	}
	database, err := c.Database()
	if err != nil {
		return nil, err
	}
	return database.Collection(name), nil
}
