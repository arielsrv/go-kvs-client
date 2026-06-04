package redis_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

func TestBuilder_FluentSetters_AllSettersAreReachable(t *testing.T) {
	// Smoke test that exercises every fluent setter so they are covered and
	// remain side-effect free. We don't assert on the result of Build because
	// we just want a chainable builder that finishes without panicking.
	builder := kvsredis.NewBuilder().
		WithAddresses("127.0.0.1:6379").
		WithKeyPrefix("__kvs:fluent").
		WithTTL(time.Minute).
		WithUsername("alice").
		WithPassword("s3cret").
		WithDB(7).
		WithMasterName("mymaster").
		WithTLS(&tls.Config{MinVersion: tls.VersionTLS12}).
		WithPoolSize(8).
		WithTimeouts(time.Second, 2*time.Second, 3*time.Second).
		WithRouteRandomly(true).
		WithTracing().
		WithMetrics()

	client := builder.FakeBuild()
	require.NotNil(t, client)
	require.Equal(t, "__kvs:fluent", client.KeyPrefix())
	require.Equal(t, time.Minute, client.TTL())
}

func TestFakeClient_MGet_PropagatesNonNotFoundErrors(t *testing.T) {
	// After Close the FakeClient returns ErrInternal for every Get; MGet must
	// propagate that error instead of treating it as a missing key.
	fake := kvsredis.NewFakeClient()
	require.NoError(t, fake.Close())

	results, err := fake.MGet(context.Background(), []string{"k"})
	require.Error(t, err)
	require.Nil(t, results)
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestGoRedisClient_MGet_PropagatesPipelineError(t *testing.T) {
	srv := miniredis.RunT(t)

	_, client := startMiniredisAt(t, srv)

	require.NoError(t, client.Set(context.Background(), "a", "1", 0))

	// Killing the server forces the pipeline Exec to fail with a connection
	// error, exercising the non-redis.Nil error branch of MGet.
	srv.Close()

	_, err := client.MGet(context.Background(), []string{"a", "b"})
	require.Error(t, err)
}

func TestLowLevelClient_BulkSaveWithContext_AllItemsFiltered_IsNoOp(t *testing.T) {
	// All items are invalid (nil / empty key / unmarshalable), so the resulting
	// pair list is empty and BulkSave must return nil without calling MSet.
	client := newClient(t)

	items := new(kvs.Items)
	items.Add(nil)
	items.Add(kvs.NewItem("", "v"))
	items.Add(kvs.NewItem("bad", make(chan int)))
	items.Add(&kvs.Item{
		Key:   "expired",
		Value: "v",
		TTL:   time.Now().Add(-time.Hour).Unix(),
	})

	require.NoError(t, client.BulkSave(items))
}

func TestGoRedisClient_MGet_PropagatesPerKeyError(t *testing.T) {
	// Store one key as a hash so the per-key GET returns WRONGTYPE, which
	// flows through pipeline.Exec as a non-Nil error but the individual
	// cmd.Result() also fails — exercising the `case err != nil` branch of
	// MGet that is otherwise hard to reach.
	srv, client := startMiniredis(t)

	require.NoError(t, client.Set(context.Background(), "ok", "1", 0))
	// HSET makes "wrong" a hash; GET on it returns WRONGTYPE.
	srv.HSet("wrong", "field", "value")

	_, err := client.MGet(context.Background(), []string{"ok", "wrong"})
	require.Error(t, err)
}
