// Package kvs provides a generic key-value store client interface and implementation.
// It supports operations like Get, Save, BulkGet, and BulkSave with optional context and TTL.
package kvs

import (
	"context"
)

// KeyMapperFunc is a function type that extracts a key from an item of type T.
// It is used in bulk operations to determine the key for each item.
type KeyMapperFunc[T any] func(item T) string

// KVSClient is the main interface for interacting with a key-value store.
// It provides methods for getting and saving individual items or collections of items.
// The interface is generic over type T, allowing it to work with any data type.
type KVSClient[T any] interface {
	// Get retrieves an item by its key.
	// Returns a pointer to the item if found, or an error if not found or if retrieval fails.
	Get(key string) (*T, error)

	// BulkGet retrieves multiple items by their keys.
	// Returns a slice of items that were found, or an error if retrieval fails.
	BulkGet(key []string) ([]T, error)

	// Save stores an item with the specified key.
	// Optional TTL (Time To Live) in seconds can be provided to automatically expire the item.
	// Returns an error if the save operation fails.
	Save(key string, item *T, ttl ...int64) error

	// BulkSave stores multiple items.
	// The keyMapper function is used to extract the key from each item.
	// Optional TTL (Time To Live) in seconds can be provided to automatically expire the items.
	// Returns an error if the save operation fails.
	BulkSave(items []T, keyMapper KeyMapperFunc[T], ttl ...int64) error

	// GetWithContext is like Get but with context support for cancellation and timeouts.
	GetWithContext(ctx context.Context, key string) (*T, error)

	// BulkGetWithContext is like BulkGet but with context support for cancellation and timeouts.
	BulkGetWithContext(ctx context.Context, keys []string) ([]T, error)

	// SaveWithContext is like Save but with context support for cancellation and timeouts.
	SaveWithContext(ctx context.Context, key string, item *T, ttl ...int64) error

	// BulkSaveWithContext is like BulkSave but with context support for cancellation and timeouts.
	BulkSaveWithContext(ctx context.Context, items []T, keyMapper KeyMapperFunc[T], ttl ...int64) error
}
