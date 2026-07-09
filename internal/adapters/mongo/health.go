package mongo

import "context"

type Health struct {
	OK      bool
	Message string
}

func (c *Client) Health(ctx context.Context) Health {
	if err := c.Ping(ctx); err != nil {
		return Health{OK: false, Message: MapError(err).Error()}
	}
	return Health{OK: true, Message: "ok"}
}
