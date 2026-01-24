package kvs

import (
	"context"
	"time"
)

// AWSKVSClient is an implementation of the Client interface for AWS services.
// It uses a LowLevelClientProxy to interact with the underlying storage.
// The struct is generic over type T, allowing it to work with any data type.
type AWSKVSClient[T any] struct {
	lowLevelClient LowLevelClientProxy
}

// NewAWSKVSClient creates a new AWSKVSClient with the provided low-level client.
// The low-level client is wrapped in a LowLevelClientProxy to add metrics and other functionality.
// Returns a pointer to the new AWSKVSClient.
func NewAWSKVSClient[T any](lowLevelClient LowLevelClient) *AWSKVSClient[T] {
	return &AWSKVSClient[T]{
		lowLevelClient: NewLowLevelClientProxy(lowLevelClient),
	}
}

// Get retrieves an item by its key.
// It uses a background context and delegates to GetWithContext.
// Returns a pointer to the item if found, or an error if not found or if retrieval fails.
func (r AWSKVSClient[T]) Get(key string) (*T, error) {
	return r.GetWithContext(context.Background(), key)
}

// BulkGet retrieves multiple items by their keys.
// It uses a background context and delegates to BulkGetWithContext.
// Returns a slice of items that were found, or an error if retrieval fails.
func (r AWSKVSClient[T]) BulkGet(key []string) ([]T, error) {
	return r.BulkGetWithContext(context.Background(), key)
}

// Save stores an item with the specified key.
// It uses a background context and delegates to SaveWithContext.
// Optional TTL (Time To Live) in seconds can be provided to automatically expire the item.
// Returns an error if the save operation fails.
func (r AWSKVSClient[T]) Save(key string, item *T, ttl ...time.Duration) error {
	return r.SaveWithContext(context.Background(), key, item, ttl...)
}

// BulkSave stores multiple items.
// It uses a background context and delegates to BulkSaveWithContext.
// The keyMapper function is used to extract the key from each item.
// Optional TTL (Time To Live) in seconds can be provided to automatically expire the items.
// Returns an error if the save operation fails.
func (r AWSKVSClient[T]) BulkSave(items []T, keyMapper KeyMapperFunc[T], ttl ...time.Duration) error {
	return r.BulkSaveWithContext(context.Background(), items, keyMapper, ttl...)
}

// GetWithContext retrieves an item by its key using the provided context.
// The context can be used for cancellation and timeouts.
// Returns a pointer to the item if found, or an error if not found or if retrieval fails.
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

// BulkGetWithContext retrieves multiple items by their keys using the provided context.
// The context can be used for cancellation and timeouts.
// Returns a slice of items that were found, or an error if retrieval fails.
// If unmarshalling of an individual item fails, it is skipped and an error is logged.
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
			continue
		}

		result = append(result, *value)
	}

	return result, nil
}

// SaveWithContext stores an item with the specified key using the provided context.
// The context can be used for cancellation and timeouts.
// Optional TTL (Time To Live) in seconds can be provided to automatically expire the item.
// Returns an error if the save operation fails.
func (r AWSKVSClient[T]) SaveWithContext(ctx context.Context, key string, value *T, ttl ...time.Duration) error {
	item := NewItem(key, value, ttl...)
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	if err != nil {
		return err
	}

	return nil
}

// BulkSaveWithContext stores multiple items using the provided context.
// The context can be used for cancellation and timeouts.
// The keyMapper function is used to extract the key from each item.
// Optional TTL (Time To Live) in seconds can be provided to automatically expire the items.
// Returns an error if the save operation fails.
func (r AWSKVSClient[T]) BulkSaveWithContext(
	ctx context.Context,
	items []T,
	keyMapper KeyMapperFunc[T],
	ttl ...time.Duration,
) error {
	kvsItems := new(Items)
	for i := range items {
		item := items[i]
		kvsItems.Add(NewItem(keyMapper(item), &item, ttl...))
	}

	err := r.lowLevelClient.BulkSaveWithContext(ctx, kvsItems)
	if err != nil {
		return err
	}

	return nil
}
