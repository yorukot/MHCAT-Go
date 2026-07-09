package fakemongo

import "context"

type RepositoryHealth struct {
	Err error
}

func (h RepositoryHealth) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return h.Err
}
