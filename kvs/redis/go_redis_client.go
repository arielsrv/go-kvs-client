// Package redis provides a Redis implementation of the KVS client.
package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/arielsrv/go-kvs-client/kvs"
)

// GoRedisClient is the production Client adapter built on top of
// github.com/redis/go-redis/v9. It transparently supports standalone,
// Sentinel and Cluster deployments through redis.UniversalClient.
//
// All bulk operations are implemented with pipelines so they remain
// correct under Redis Cluster (where MGET/MSET are not allowed across
// hash slots), while still keeping the number of round-trips minimal.
type GoRedisClient struct {
	client goredis.UniversalClient
}

// NewGoRedisClient wraps the provided go-redis UniversalClient as a Client.
// The caller retains ownership of the underlying client; calling Close on
// GoRedisClient will close the wrapped instance.
func NewGoRedisClient(client goredis.UniversalClient) *GoRedisClient {
	return &GoRedisClient{client: client}
}

// Get implements Client.
func (r *GoRedisClient) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", kvs.ErrKeyNotFound
		}
		return "", err
	}
	return value, nil
}

// Set implements Client.
// When ttl is non-positive the entry is stored without expiration.
func (r *GoRedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	expiration := max(ttl, 0)
	return r.client.Set(ctx, key, value, expiration).Err()
}

// MGet implements Client using a single pipeline of GET commands so that the
// operation is correct under Redis Cluster regardless of hash-slot distribution.
func (r *GoRedisClient) MGet(ctx context.Context, keys []string) ([]GetResult, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*goredis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, goredis.Nil) {
		return nil, err
	}

	results := make([]GetResult, len(keys))
	for i, cmd := range cmds {
		value, err := cmd.Result()
		switch {
		case errors.Is(err, goredis.Nil):
			results[i] = GetResult{Key: keys[i], Found: false}
		case err != nil:
			return nil, err
		default:
			results[i] = GetResult{Key: keys[i], Value: value, Found: true}
		}
	}
	return results, nil
}

// MSet implements Client using a pipeline so per-item TTL is honoured.
// Empty input is a no-op.
func (r *GoRedisClient) MSet(ctx context.Context, pairs []Pair) error {
	if len(pairs) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, pair := range pairs {
		expiration := max(pair.TTL, 0)
		pipe.Set(ctx, pair.Key, pair.Value, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Close implements Client.
func (r *GoRedisClient) Close() error {
	return r.client.Close()
}
