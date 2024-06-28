package infrastructure

import (
	"context"
)

type KVSClient[T any] interface {
	Get(key string) (*T, error)
	BulkGet(key []string) ([]T, error)
	Save(key string, item *T) error
	BulkSave(items []T, keyMapper func(item T) string) error

	GetWithContext(ctx context.Context, key string) (*T, error)
	BulkGetWithContext(ctx context.Context, keys []string) ([]T, error)
	SaveWithContext(ctx context.Context, key string, item *T) error
	BulkSaveWithContext(ctx context.Context, items []T, keyMapper func(item T) string) error
}
