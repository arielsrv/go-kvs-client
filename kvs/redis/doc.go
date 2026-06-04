// Package redis provides a Redis implementation of the kvs.LowLevelClient interface.
//
// This package contains the necessary components to interact with Redis (standalone,
// Sentinel or Cluster) as a key-value store. It implements the generic interface
// defined in the parent kvs package, allowing for seamless integration while
// maintaining the type-safety and convenience of the KVS client.
//
// Key Components:
//   - LowLevelClient: implements kvs.LowLevelClient on top of Redis.
//   - Client: minimal interface that hides the concrete Redis driver, making the
//     implementation testable and swap-friendly.
//   - GoRedisClient: production adapter built on top of github.com/redis/go-redis/v9.
//     It transparently supports standalone, Sentinel and Cluster deployments via
//     redis.UniversalClient.
//   - FakeClient: in-memory implementation of Client used for unit tests (and
//     exposed through Builder.FakeBuild).
//   - Builder: fluent / functional-options builder that wires everything together.
//
// Usage:
//
//	import (
//	    "github.com/arielsrv/go-kvs-client/kvs"
//	    "github.com/arielsrv/go-kvs-client/kvs/redis"
//	)
//
//	client := kvs.NewKVSClient[MyType](
//	    redis.NewBuilder(
//	        redis.WithAddresses("localhost:6379"),
//	        redis.WithKeyPrefix("__kvs:users"),
//	        redis.WithTTL(24 * time.Hour),
//	    ).Build(),
//	)
//
// The implementation honours the same contract as the DynamoDB backend:
//   - Single-flight de-duplication of concurrent reads for the same key.
//   - Per-item TTL with sensible fallback to the builder default.
//   - Bulk operations capped at 100 keys to match kvs.ErrTooManyKeys semantics.
//   - JSON value serialization so values stored by any backend are interchangeable.
package redis
