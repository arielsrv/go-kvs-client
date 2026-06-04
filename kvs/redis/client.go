// Package redis provides a Redis implementation of the KVS client.
package redis

import (
	"context"
	"time"
)

// MaxBulkKeys is the maximum number of keys allowed in a bulk operation.
// It mirrors the limit enforced by the DynamoDB backend so that the public
// kvs.Client[T] contract is consistent across providers.
const MaxBulkKeys = 100

// Pair represents a single key/value pair to be written through MSet.
// TTL is the duration relative to "now" after which the entry must expire;
// a zero or negative TTL means the entry should be persisted without expiration.
type Pair struct {
	Key   string
	Value string
	TTL   time.Duration
}

// GetResult represents the result of fetching a single key in a bulk operation.
// Found is false when the key does not exist; in that case Value is empty.
type GetResult struct {
	Key   string
	Value string
	Found bool
}

// Client is the minimal Redis interface required by LowLevelClient.
//
// Keeping this interface narrow has two benefits:
//   - The production adapter (GoRedisClient, built on top of go-redis/v9)
//     can be swapped for any other driver without touching LowLevelClient.
//   - Unit tests can rely on FakeClient (a fast, deterministic, in-memory
//     implementation) instead of spinning up a real Redis instance.
//
// Implementations MUST be safe for concurrent use.
type Client interface {
	// Get fetches the value for the given key.
	// When the key does not exist, implementations MUST return kvs.ErrKeyNotFound.
	Get(ctx context.Context, key string) (string, error)

	// Set stores the value for the given key with an optional TTL.
	// A non-positive ttl means the entry has no expiration.
	Set(ctx context.Context, key, value string, ttl time.Duration) error

	// MGet fetches multiple keys in a single round-trip when possible.
	// The returned slice has the same length as the input keys and preserves
	// their order; missing keys are reported with Found == false.
	MGet(ctx context.Context, keys []string) ([]GetResult, error)

	// MSet stores multiple pairs in a single round-trip when possible.
	// Implementations SHOULD honour per-item TTL.
	MSet(ctx context.Context, pairs []Pair) error

	// Close releases any resources held by the client.
	// Calling Close on an already closed client is a no-op.
	Close() error
}
