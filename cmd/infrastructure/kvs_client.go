package infrastructure

import (
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

type Client[T any] interface {
	Get(key string) (*T, error)
	BulkGet(key []string) ([]T, error)
	Save(key string, item *T) error
}

type KVSClient[T any] struct {
	lowLevelClient kvs.LowLevelClient
}

func NewKVSClient[T any](lowLevelClient kvs.LowLevelClient) *KVSClient[T] {
	return &KVSClient[T]{
		lowLevelClient: lowLevelClient,
	}
}

func (r KVSClient[T]) Get(key string) (*T, error) {
	item, err := r.lowLevelClient.Get(key)
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

func (r KVSClient[T]) BulkGet(keys []string) ([]T, error) {
	result := make([]T, 0)

	items, err := r.lowLevelClient.BulkGet(keys)
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

func (r KVSClient[T]) Save(key string, value *T) error {
	item := kvs.NewItem(key, value)
	err := r.lowLevelClient.Save(key, item)
	if err != nil {
		return err
	}

	return nil
}
