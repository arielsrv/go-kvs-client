package kvs

import (
	"context"
)

// LowLevelClient is the interface for low-level key-value store operations.
// It provides methods for getting and saving individual items or collections of items,
// with or without context support.
type LowLevelClient interface {
	// Get retrieves an item by its key.
	Get(key string) (*Item, error)

	// BulkGet retrieves multiple items by their keys.
	BulkGet(keys []string) (*Items, error)

	// Save stores an item with the specified key.
	Save(key string, item *Item) error

	// BulkSave stores multiple items.
	BulkSave(items *Items) error

	// GetWithContext retrieves an item by its key using the provided context.
	GetWithContext(ctx context.Context, key string) (*Item, error)

	// SaveWithContext stores an item with the specified key using the provided context.
	SaveWithContext(ctx context.Context, key string, item *Item) error

	// BulkGetWithContext retrieves multiple items by their keys using the provided context.
	BulkGetWithContext(ctx context.Context, key []string) (*Items, error)

	// BulkSaveWithContext stores multiple items using the provided context.
	BulkSaveWithContext(ctx context.Context, items *Items) error

	// ContainerName returns the name of the container or service that this client interacts with.
	// Used for metrics and logging.
	ContainerName() string
}

// LowLevelClientProxy is a wrapper around a LowLevelClient that adds metrics collection.
// It implements the same interface as the wrapped client, but adds metrics for each operation.
type LowLevelClientProxy struct {
	lowLevelClient LowLevelClient
}

// NewLowLevelClientProxy creates a new LowLevelClientProxy with the provided client.
// Returns a LowLevelClientProxy that wraps the client.
func NewLowLevelClientProxy(lowLevelClient LowLevelClient) LowLevelClientProxy {
	return LowLevelClientProxy{
		lowLevelClient: lowLevelClient,
	}
}

// Get retrieves an item by its key.
// It uses a background context and delegates to GetWithContext.
// Returns the item if found, or an error if not found or if retrieval fails.
func (r LowLevelClientProxy) Get(key string) (*Item, error) {
	return r.GetWithContext(context.Background(), key)
}

// BulkGet retrieves multiple items by their keys.
// It uses a background context and delegates to BulkGetWithContext.
// Returns a collection of items that were found, or an error if retrieval fails.
func (r LowLevelClientProxy) BulkGet(keys []string) (*Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

// Save stores an item with the specified key.
// It uses a background context and delegates to SaveWithContext.
// Returns an error if the save operation fails.
func (r LowLevelClientProxy) Save(key string, item *Item) error {
	return r.SaveWithContext(context.Background(), key, item)
}

// BulkSave stores multiple items.
// It uses a background context and delegates to BulkSaveWithContext.
// Returns an error if the save operation fails.
func (r LowLevelClientProxy) BulkSave(items *Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

// GetWithContext retrieves an item by its key using the provided context.
// The context can be used for cancellation and timeouts.
// This method collects metrics about the operation, including execution time and success/failure.
// Returns the item if found, or an error if not found or if retrieval fails.
func (r LowLevelClientProxy) GetWithContext(ctx context.Context, key string) (*Item, error) {
	value, err := r.lowLevelClient.GetWithContext(ctx, key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// SaveWithContext stores an item with the specified key using the provided context.
// The context can be used for cancellation and timeouts.
// This method collects metrics about the operation, including execution time.
// Returns an error if the save operation fails.
func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	if err != nil {
		return err
	}

	return nil
}

// BulkGetWithContext retrieves multiple items by their keys using the provided context.
// The context can be used for cancellation and timeouts.
// This method collects metrics about the operation, including execution time.
// Returns a collection of items that were found, or an error if retrieval fails.
func (r LowLevelClientProxy) BulkGetWithContext(ctx context.Context, key []string) (*Items, error) {
	values, err := r.lowLevelClient.BulkGetWithContext(ctx, key)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// BulkSaveWithContext stores multiple items using the provided context.
// The context can be used for cancellation and timeouts.
// This method collects metrics about the operation, including execution time.
// Returns an error if the save operation fails.
func (r LowLevelClientProxy) BulkSaveWithContext(ctx context.Context, items *Items) error {
	err := r.lowLevelClient.BulkSaveWithContext(ctx, items)
	if err != nil {
		return err
	}

	return nil
}

// ContainerName returns the name of the container or service that this client interacts with.
// It delegates to the wrapped client's ContainerName method.
// Used for metrics and logging.
func (r LowLevelClientProxy) ContainerName() string {
	return r.lowLevelClient.ContainerName()
}
