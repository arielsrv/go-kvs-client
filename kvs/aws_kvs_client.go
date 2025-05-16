package kvs

import (
	"context"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

type AWSKVSClient[T any] struct {
	lowLevelClient LowLevelClientProxy
}

func NewAWSKVSClient[T any](lowLevelClient Client) *AWSKVSClient[T] {
	return &AWSKVSClient[T]{
		lowLevelClient: NewLowLevelClientProxy(lowLevelClient),
	}
}

func (r AWSKVSClient[T]) Get(key string) (*T, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r AWSKVSClient[T]) BulkGet(key []string) ([]T, error) {
	return r.BulkGetWithContext(context.Background(), key)
}

func (r AWSKVSClient[T]) Save(key string, item *T, ttl ...int64) error {
	return r.SaveWithContext(context.Background(), key, item, ttl...)
}

func (r AWSKVSClient[T]) BulkSave(items []T, keyMapper KeyMapperFunc[T], ttl ...int64) error {
	return r.BulkSaveWithContext(context.Background(), items, keyMapper, ttl...)
}

func (r AWSKVSClient[T]) GetWithContext(ctx context.Context, key string) (*T, error) {
	item, err := r.lowLevelClient.GetWithContext(ctx, key)
	if err != nil {
		return nil, err
	}

	value := new(T)
	err = item.TryGetValueAsObjectType(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r AWSKVSClient[T]) BulkGetWithContext(ctx context.Context, keys []string) ([]T, error) {
	result := make([]T, 0)

	items, err := r.lowLevelClient.BulkGetWithContext(ctx, keys)
	if err != nil {
		return nil, err
	}

	for item := range items.All() {
		value := new(T)
		mErr := item.TryGetValueAsObjectType(&value)
		if mErr != nil {
			log.Errorf("[kvs]: error unmarshalling value for item %v", item)
			continue
		}

		result = append(result, *value)
	}

	return result, nil
}

func (r AWSKVSClient[T]) SaveWithContext(ctx context.Context, key string, value *T, ttl ...int64) error {
	item := NewItem(key, value, ttl...)
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	if err != nil {
		log.Errorf("[kvs]: error saving item %s: %v", key, err)
		return err
	}

	return nil
}

func (r AWSKVSClient[T]) BulkSaveWithContext(
	ctx context.Context,
	items []T,
	keyMapper KeyMapperFunc[T],
	ttl ...int64,
) error {
	kvsItems := new(Items)
	for i := range items {
		item := items[i]
		kvsItems.Add(NewItem(keyMapper(item), &item, ttl...))
	}

	err := r.lowLevelClient.BulkSaveWithContext(ctx, kvsItems)
	if err != nil {
		log.Errorf("[kvs]: error saving items: %v", err)
		return err
	}

	return nil
}
