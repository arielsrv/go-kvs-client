package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

func TestFakeClient_Set_Get(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	ctx := context.Background()

	require.NoError(t, fake.Set(ctx, "k", "v", 0))
	value, err := fake.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v", value)
}

func TestFakeClient_Get_NotFound(t *testing.T) {
	fake := kvsredis.NewFakeClient()

	_, err := fake.Get(context.Background(), "missing")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestFakeClient_Set_TTL_Expires(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	ctx := context.Background()

	require.NoError(t, fake.Set(ctx, "k", "v", 30*time.Millisecond))
	require.Equal(t, 1, fake.Len())

	time.Sleep(80 * time.Millisecond)

	_, err := fake.Get(ctx, "k")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Equal(t, 0, fake.Len(), "expired entry must be evicted lazily")
}

func TestFakeClient_MGet_MSet(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	ctx := context.Background()

	require.NoError(t, fake.MSet(ctx, []kvsredis.Pair{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2", TTL: time.Hour},
	}))

	results, err := fake.MGet(ctx, []string{"a", "b", "missing"})
	require.NoError(t, err)
	require.Len(t, results, 3)

	require.True(t, results[0].Found)
	require.Equal(t, "1", results[0].Value)
	require.True(t, results[1].Found)
	require.Equal(t, "2", results[1].Value)
	require.False(t, results[2].Found)
}

func TestFakeClient_Close_RejectsOperations(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	require.NoError(t, fake.Close())

	err := fake.Set(context.Background(), "k", "v", 0)
	require.Error(t, err)

	_, err = fake.Get(context.Background(), "k")
	require.Error(t, err)
}

func TestFakeClient_ConcurrentAccess(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	ctx := context.Background()

	const goroutines = 32
	done := make(chan struct{}, goroutines)

	for range goroutines {
		go func() {
			defer func() { done <- struct{}{} }()
			key := "k"
			_ = fake.Set(ctx, key, "v", 0)
			_, _ = fake.Get(ctx, key)
		}()
	}
	for range goroutines {
		<-done
	}

	value, err := fake.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v", value)
}
