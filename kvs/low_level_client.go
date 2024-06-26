package kvs

import "context"

type LowLevelClient interface {
	Get(key string) (*Item, error)
	GetWithContext(ctx context.Context, key string) (*Item, error)
	Save(key string, item *Item) error
	SaveWithContext(ctx context.Context, key string, item *Item) error
	BulkGet(keys []string) (*Items, error)
	BulkGetWithContext(ctx context.Context, key []string) (*Items, error)
	BulkSave(items *Items) error
	BulkSaveWithContext(ctx context.Context, items *Items) error
}
