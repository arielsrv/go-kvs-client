// Package redis provides a Redis implementation of the KVS client.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/arielsrv/go-kvs-client/kvs"
)

// LowLevelClient is the Redis implementation of kvs.LowLevelClient.
//
// It mirrors the behaviour of the DynamoDB backend so that the high-level
// kvs.AWSKVSClient[T] can be backed by either provider transparently:
//
//   - Values are stored as JSON strings (identical wire format).
//   - TTL is honoured on a per-item basis (item.TTL takes precedence over the
//     builder default; an item whose TTL is already in the past is skipped).
//   - Concurrent reads for the same key are de-duplicated via singleflight.
//   - Bulk operations are capped at MaxBulkKeys keys to preserve a consistent
//     contract with the DynamoDB implementation (kvs.ErrTooManyKeys).
//   - Keys are automatically namespaced with the configured key prefix.
type LowLevelClient struct {
	client    Client
	read      singleflight.Group
	keyPrefix string
	ttl       time.Duration
}

// NewLowLevelClient creates a new LowLevelClient backed by the provided
// Client implementation. The key prefix is applied transparently to every
// read/write operation; it should not contain a trailing separator (one is
// added automatically when joining the prefix with the user-supplied key).
func NewLowLevelClient(client Client, keyPrefix string, ttl ...time.Duration) *LowLevelClient {
	llc := &LowLevelClient{
		client:    client,
		keyPrefix: strings.TrimSuffix(strings.TrimSpace(keyPrefix), ":"),
	}
	if len(ttl) > 0 {
		llc.ttl = ttl[0]
	}
	return llc
}

// KeyPrefix returns the configured key prefix (without trailing separator).
func (r *LowLevelClient) KeyPrefix() string {
	return r.keyPrefix
}

// TTL returns the default TTL applied when an item has no explicit TTL.
func (r *LowLevelClient) TTL() time.Duration {
	return r.ttl
}

// ContainerName implements kvs.LowLevelClient.
// For Redis, the "container" is conceptually the configured key prefix
// (or the literal string "redis" when no prefix is set). It is used for
// metrics and logging only.
func (r *LowLevelClient) ContainerName() string {
	if r.keyPrefix == "" {
		return "redis"
	}
	return r.keyPrefix
}

// Close releases the underlying Redis connection pool.
func (r *LowLevelClient) Close() error {
	return r.client.Close()
}

// Get implements kvs.LowLevelClient.
func (r *LowLevelClient) Get(key string) (*kvs.Item, error) {
	return r.GetWithContext(context.Background(), key)
}

// GetWithContext implements kvs.LowLevelClient.
// Concurrent reads for the same key are de-duplicated via singleflight.
func (r *LowLevelClient) GetWithContext(ctx context.Context, key string) (*kvs.Item, error) {
	if strings.TrimSpace(key) == "" {
		return nil, kvs.ErrEmptyKey
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("redis GetWithContext: %w", err)
	}

	result, err, _ := r.read.Do(key, func() (any, error) {
		value, gErr := r.client.Get(ctx, r.fullKey(key))
		if gErr != nil {
			return nil, gErr
		}
		return &kvs.Item{
			Key:   key,
			Value: value,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*kvs.Item), nil
}

// Save implements kvs.LowLevelClient.
func (r *LowLevelClient) Save(key string, item *kvs.Item) error {
	return r.SaveWithContext(context.Background(), key, item)
}

// SaveWithContext implements kvs.LowLevelClient.
// The value is marshalled to JSON; the resulting string is what gets stored
// in Redis. TTL semantics:
//   - if item.TTL > 0 it is interpreted as a Unix timestamp and converted to
//     the remaining duration; if it has already elapsed, the call is a no-op.
//   - otherwise, the builder default (r.ttl) is used (zero means "no expiration").
func (r *LowLevelClient) SaveWithContext(ctx context.Context, key string, item *kvs.Item) error {
	if strings.TrimSpace(key) == "" {
		return kvs.ErrEmptyKey
	}
	if item == nil {
		return kvs.ErrNilItem
	}

	bytes, err := json.Marshal(item.Value)
	if err != nil {
		return fmt.Errorf("redis SaveWithContext: marshal: %w", err)
	}

	ttl, skip := r.resolveTTL(item.TTL)
	if skip {
		return nil
	}

	return r.client.Set(ctx, r.fullKey(key), string(bytes), ttl)
}

// BulkGet implements kvs.LowLevelClient.
func (r *LowLevelClient) BulkGet(keys []string) (*kvs.Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

// BulkGetWithContext implements kvs.LowLevelClient.
// At most MaxBulkKeys keys are accepted; otherwise kvs.ErrTooManyKeys is returned.
// Missing keys are silently skipped (consistent with the DynamoDB backend).
func (r *LowLevelClient) BulkGetWithContext(ctx context.Context, keys []string) (*kvs.Items, error) {
	if len(keys) > MaxBulkKeys {
		return nil, kvs.ErrTooManyKeys
	}
	if len(keys) == 0 {
		return new(kvs.Items), nil
	}

	prefixed := make([]string, len(keys))
	for i, key := range keys {
		prefixed[i] = r.fullKey(key)
	}

	results, err := r.client.MGet(ctx, prefixed)
	if err != nil {
		return nil, err
	}

	items := new(kvs.Items)
	for i, result := range results {
		if !result.Found {
			continue
		}
		items.Add(&kvs.Item{
			Key:   keys[i],
			Value: result.Value,
		})
	}
	return items, nil
}

// BulkSave implements kvs.LowLevelClient.
func (r *LowLevelClient) BulkSave(items *kvs.Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

// BulkSaveWithContext implements kvs.LowLevelClient.
// Items that fail to marshal or whose TTL is already in the past are skipped.
func (r *LowLevelClient) BulkSaveWithContext(ctx context.Context, kvsItems *kvs.Items) error {
	if kvsItems == nil || kvsItems.Len() == 0 {
		return nil
	}

	pairs := make([]Pair, 0, kvsItems.Len())
	for item := range kvsItems.All() {
		if item == nil || strings.TrimSpace(item.Key) == "" {
			continue
		}

		bytes, err := json.Marshal(item.Value)
		if err != nil {
			// Same behaviour as the DynamoDB backend: skip non-serialisable items.
			continue
		}

		ttl, skip := r.resolveTTL(item.TTL)
		if skip {
			continue
		}

		pairs = append(pairs, Pair{
			Key:   r.fullKey(item.Key),
			Value: string(bytes),
			TTL:   ttl,
		})
	}

	if len(pairs) == 0 {
		return nil
	}

	if err := r.client.MSet(ctx, pairs); err != nil {
		return err
	}
	return nil
}

// fullKey joins the configured prefix and the user-supplied key.
func (r *LowLevelClient) fullKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	return r.keyPrefix + ":" + key
}

// resolveTTL converts a kvs.Item TTL (Unix timestamp) into a duration suitable
// for Redis. The boolean result is true when the item must be skipped because
// its TTL is already in the past.
func (r *LowLevelClient) resolveTTL(itemTTL int64) (time.Duration, bool) {
	if itemTTL > 0 {
		remaining := time.Until(time.Unix(itemTTL, 0))
		if remaining <= 0 {
			return 0, true
		}
		return remaining, false
	}
	return r.ttl, false
}
