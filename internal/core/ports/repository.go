package ports

import "context"

type PageRequest struct {
	Limit  int
	Cursor string
}

type PageResult[T any] struct {
	Items      []T
	NextCursor string
}

type RepositoryHealth interface {
	Ping(ctx context.Context) error
}
