package redis_test

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

func TestBuilder_WithOptions(t *testing.T) {
	builder := kvsredis.NewBuilder(
		kvsredis.WithAddresses("redis-1:6379", "redis-2:6379"),
		kvsredis.WithKeyPrefix("__kvs:users"),
		kvsredis.WithTTL(5*time.Minute),
		kvsredis.WithUsername("default"),
		kvsredis.WithPassword("s3cret"),
		kvsredis.WithDB(2),
		kvsredis.WithPoolSize(10),
		kvsredis.WithTimeouts(time.Second, 2*time.Second, 2*time.Second),
		kvsredis.WithMasterName("mymaster"),
		kvsredis.WithRouteRandomly(true),
		kvsredis.WithTLS(&tls.Config{MinVersion: tls.VersionTLS12}),
	)
	require.NotNil(t, builder)

	// Build with the in-memory fake so the test stays hermetic.
	client := builder.FakeBuild()
	require.NotNil(t, client)
	require.Equal(t, "__kvs:users", client.KeyPrefix())
	require.Equal(t, 5*time.Minute, client.TTL())
	require.Equal(t, "__kvs:users", client.ContainerName())
}

func TestBuilder_WithFluentSetters(t *testing.T) {
	builder := kvsredis.NewBuilder().
		WithAddresses("localhost:6379").
		WithKeyPrefix("__kvs:fluent:").
		WithTTL(time.Minute).
		WithDB(0).
		WithPoolSize(5)

	client := builder.FakeBuild()
	require.NotNil(t, client)
	require.Equal(t, "__kvs:fluent", client.KeyPrefix())
	require.Equal(t, time.Minute, client.TTL())
}

func TestBuilder_BuildWithClient(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	client := kvsredis.NewBuilder(
		kvsredis.WithKeyPrefix("p"),
		kvsredis.WithTTL(time.Hour),
	).BuildWithClient(fake)

	require.NotNil(t, client)
	require.Equal(t, "p", client.KeyPrefix())
	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))

	keys := fake.Keys("p:")
	require.Len(t, keys, 1)
}

func TestBuilder_FakeBuild_RoundTrip(t *testing.T) {
	client := kvsredis.NewBuilder().FakeBuild()

	err := client.SaveWithContext(t.Context(), "key", kvs.NewItem("key", "value"))
	require.NoError(t, err)

	item, err := client.GetWithContext(t.Context(), "key")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, `"value"`, item.Value)
	require.Equal(t, "key", item.Key)
}

func TestBuilder_Build_DefaultsToLocalhost(t *testing.T) {
	// We only exercise the construction path; no network round-trip happens
	// because we never invoke any operation on the returned client.
	builder := kvsredis.NewBuilder()
	client := builder.Build()
	require.NotNil(t, client)
	t.Cleanup(func() { _ = client.Close() })
}

func TestBuilder_WithTracing_AttachesHookWithoutPanic(t *testing.T) {
	// We can't easily assert that spans are emitted (that would require an
	// initialised tracer provider and a test exporter), but we can guarantee
	// the construction path works end-to-end with tracing enabled.
	builder := kvsredis.NewBuilder(
		kvsredis.WithAddresses("127.0.0.1:6379"),
		kvsredis.WithTracing(),
	)
	client := builder.Build()
	require.NotNil(t, client)
	t.Cleanup(func() { _ = client.Close() })
}

func TestBuilder_WithMetrics_AttachesHookWithoutPanic(t *testing.T) {
	builder := kvsredis.NewBuilder(
		kvsredis.WithAddresses("127.0.0.1:6379"),
		kvsredis.WithMetrics(),
	)
	client := builder.Build()
	require.NotNil(t, client)
	t.Cleanup(func() { _ = client.Close() })
}

func TestBuilder_WithTracingAndMetrics_Fluent(t *testing.T) {
	builder := kvsredis.NewBuilder().
		WithAddresses("127.0.0.1:6379").
		WithTracing().
		WithMetrics()
	client := builder.Build()
	require.NotNil(t, client)
	t.Cleanup(func() { _ = client.Close() })
}
