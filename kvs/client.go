package kvs

import "context"

type Client interface {
	Get(key string) (*Item, error)
	GetWithContext(ctx context.Context, key string) (*Item, error)
	Save(key string, item *Item) error
	SaveWithContext(ctx context.Context, key string, item *Item) error
}
