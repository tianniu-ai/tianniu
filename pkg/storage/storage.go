package storage

import "context"

type Storage interface {
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value string) error
}
