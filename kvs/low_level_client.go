package kvs

import (
	"context"
	"errors"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// Client is the interface for low-level key-value store operations.
// It provides methods for getting and saving individual items or collections of items,
// with or without context support.
type Client interface {
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

// LowLevelClientProxy is a wrapper around a Client that adds metrics collection.
// It implements the same interface as the wrapped client, but adds metrics for each operation.
type LowLevelClientProxy struct {
	lowLevelClient Client
}

// NewLowLevelClientProxy creates a new LowLevelClientProxy with the provided client.
// Returns a LowLevelClientProxy that wraps the client.
func NewLowLevelClientProxy(lowLevelClient Client) LowLevelClientProxy {
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
	metrics.Collector.Prometheus().IncrementCounter("__kvs_operations", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "get",
	})

	start := time.Now()
	value, err := r.lowLevelClient.GetWithContext(ctx, key)

	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "get",
	})

	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
				"client_name": r.lowLevelClient.ContainerName(),
				"stats":       "miss",
			})
		} else {
			metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
				"client_name": r.lowLevelClient.ContainerName(),
				"stats":       "error",
			})
		}
		return nil, err
	}

	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "hit",
	})

	return value, nil
}

// SaveWithContext stores an item with the specified key using the provided context.
// The context can be used for cancellation and timeouts.
// This method collects metrics about the operation, including execution time.
// Returns an error if the save operation fails.
func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "save",
	})

	start := time.Now()
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "save",
	})
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
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "bulk_get",
	})

	start := time.Now()
	values, err := r.lowLevelClient.BulkGetWithContext(ctx, key)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "bulk_get",
	})
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
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "bulk_save",
	})

	start := time.Now()
	err := r.lowLevelClient.BulkSaveWithContext(ctx, items)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "bulk_save",
	})
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
