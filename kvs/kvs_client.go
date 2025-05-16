package kvs

import (
	"context"
)

type KeyMapperFunc[T any] func(item T) string

type KVSClient[T any] interface {
	Get(key string) (*T, error)
	BulkGet(key []string) ([]T, error)
	Save(key string, item *T, ttl ...int64) error
	BulkSave(items []T, keyMapper KeyMapperFunc[T], ttl ...int64) error
	GetWithContext(ctx context.Context, key string) (*T, error)
	BulkGetWithContext(ctx context.Context, keys []string) ([]T, error)
	SaveWithContext(ctx context.Context, key string, item *T, ttl ...int64) error
	BulkSaveWithContext(ctx context.Context, items []T, keyMapper KeyMapperFunc[T], ttl ...int64) error
}
