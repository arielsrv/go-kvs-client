package infrastructure

import (
	"context"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

type DynamoDBKVSClient[T any] struct {
	lowLevelClient kvs.LowLevelClient
}

func NewDynamoDBKVSClient[T any](lowLevelClient kvs.LowLevelClient) *DynamoDBKVSClient[T] {
	return &DynamoDBKVSClient[T]{
		lowLevelClient: lowLevelClient,
	}
}

func (r DynamoDBKVSClient[T]) Get(key string) (*T, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r DynamoDBKVSClient[T]) BulkGet(key []string) ([]T, error) {
	return r.BulkGetWithContext(context.Background(), key)
}

func (r DynamoDBKVSClient[T]) Save(key string, item *T) error {
	return r.SaveWithContext(context.Background(), key, item)
}

func (r DynamoDBKVSClient[T]) BulkSave(items []T, keyMapper func(item T) string) error {
	return r.BulkSaveWithContext(context.Background(), items, keyMapper)
}

func (r DynamoDBKVSClient[T]) GetWithContext(ctx context.Context, key string) (*T, error) {
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

func (r DynamoDBKVSClient[T]) BulkGetWithContext(ctx context.Context, keys []string) ([]T, error) {
	result := make([]T, 0)

	items, err := r.lowLevelClient.BulkGetWithContext(ctx, keys)
	if err != nil {
		return nil, err
	}

	for i := range items.GetOks() {
		item := items.Items[i]
		value := new(T)
		mErr := item.TryGetValueAsObjectType(&value)
		if mErr != nil {
			continue
		}

		result = append(result, *value)
	}

	return result, nil
}

func (r DynamoDBKVSClient[T]) SaveWithContext(ctx context.Context, key string, value *T) error {
	item := kvs.NewItem(key, value)
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	if err != nil {
		return err
	}

	return nil
}

func (r DynamoDBKVSClient[T]) BulkSaveWithContext(ctx context.Context, items []T, keyMapper func(item T) string) error {
	if len(items) == 0 {
		return nil
	}

	kvsItems := new(kvs.Items)
	for i := range items {
		item := items[i]
		kvsItems.Add(&kvs.Item{
			Key:   keyMapper(item),
			Value: item,
		})
	}

	err := r.lowLevelClient.BulkSaveWithContext(ctx, kvsItems)
	if err != nil {
		return err
	}

	return nil
}
