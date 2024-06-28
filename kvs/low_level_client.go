package kvs

import "context"

type LowLevelClient interface {
	Get(key string) (*Item, error)
	BulkGet(keys []string) (*Items, error)
	Save(key string, item *Item) error
	BulkSave(items *Items) error

	GetWithContext(ctx context.Context, key string) (*Item, error)
	SaveWithContext(ctx context.Context, key string, item *Item) error
	BulkGetWithContext(ctx context.Context, key []string) (*Items, error)
	BulkSaveWithContext(ctx context.Context, items *Items) error
}
