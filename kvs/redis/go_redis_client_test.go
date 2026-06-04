package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

// startMiniredis starts an in-process Redis server and returns a configured
// GoRedisClient backed by it. The server is shut down automatically when the
// test ends.
func startMiniredis(t *testing.T) (*miniredis.Miniredis, *kvsredis.GoRedisClient) {
	t.Helper()
	return startMiniredisAt(t, miniredis.RunT(t))
}

func TestGoRedisClient_Set_Get(t *testing.T) {
	_, client := startMiniredis(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "v", 0))

	value, err := client.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v", value)
}

func TestGoRedisClient_Get_NotFound_ReturnsErrKeyNotFound(t *testing.T) {
	_, client := startMiniredis(t)

	_, err := client.Get(context.Background(), "missing")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestGoRedisClient_Set_WithTTL_Expires(t *testing.T) {
	srv, client := startMiniredis(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "v", time.Hour))

	// Fast-forward the in-process clock to trigger expiration deterministically.
	srv.FastForward(2 * time.Hour)

	_, err := client.Get(ctx, "k")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestGoRedisClient_Set_NegativeTTL_IsTreatedAsNoExpiration(t *testing.T) {
	srv, client := startMiniredis(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "v", -time.Hour))

	srv.FastForward(24 * time.Hour)

	value, err := client.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v", value)
}

func TestGoRedisClient_MGet_MixedHitsAndMisses(t *testing.T) {
	_, client := startMiniredis(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "a", "1", 0))
	require.NoError(t, client.Set(ctx, "c", "3", 0))

	results, err := client.MGet(ctx, []string{"a", "b", "c"})
	require.NoError(t, err)
	require.Len(t, results, 3)

	require.True(t, results[0].Found)
	require.Equal(t, "1", results[0].Value)

	require.False(t, results[1].Found)

	require.True(t, results[2].Found)
	require.Equal(t, "3", results[2].Value)
}

func TestGoRedisClient_MGet_Empty_ReturnsNil(t *testing.T) {
	_, client := startMiniredis(t)

	results, err := client.MGet(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, results)
}

func TestGoRedisClient_MSet_PerItemTTL(t *testing.T) {
	srv, client := startMiniredis(t)
	ctx := context.Background()

	require.NoError(t, client.MSet(ctx, []kvsredis.Pair{
		{Key: "persistent", Value: "1"},
		{Key: "ephemeral", Value: "2", TTL: time.Hour},
		{Key: "negative", Value: "3", TTL: -time.Minute},
	}))

	v, err := client.Get(ctx, "persistent")
	require.NoError(t, err)
	require.Equal(t, "1", v)

	srv.FastForward(2 * time.Hour)

	_, err = client.Get(ctx, "ephemeral")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)

	v, err = client.Get(ctx, "negative")
	require.NoError(t, err)
	require.Equal(t, "3", v)
}

func TestGoRedisClient_MSet_Empty_IsNoOp(t *testing.T) {
	_, client := startMiniredis(t)

	require.NoError(t, client.MSet(context.Background(), nil))
}

func TestGoRedisClient_PropagatesErrorsAfterClose(t *testing.T) {
	_, client := startMiniredis(t)
	require.NoError(t, client.Close())

	// All subsequent operations should fail with the closed-pool error from
	// go-redis. We don't pin the exact sentinel because go-redis may change it;
	// the important property is "non-nil, non-kvs error".
	_, err := client.Get(context.Background(), "any")
	require.Error(t, err)
	require.NotErrorIs(t, err, kvs.ErrKeyNotFound)

	err = client.Set(context.Background(), "any", "v", 0)
	require.Error(t, err)

	_, err = client.MGet(context.Background(), []string{"a", "b"})
	require.Error(t, err)

	err = client.MSet(context.Background(), []kvsredis.Pair{{Key: "a", Value: "1"}})
	require.Error(t, err)
}

func TestLowLevelClient_EndToEnd_OnMiniredis(t *testing.T) {
	// Wire the high-level KVSClient[T] all the way to miniredis to verify the
	// full stack (generic -> low-level -> GoRedisClient -> miniredis).
	_, client := startMiniredis(t)
	ll := kvsredis.NewLowLevelClient(client, "__kvs:e2e", time.Hour)
	high := kvs.NewKVSClient[testUser](ll)

	require.NoError(t, high.Save("42", &testUser{ID: 42, Name: "Douglas"}))

	got, err := high.Get("42")
	require.NoError(t, err)
	require.Equal(t, 42, got.ID)
	require.Equal(t, "Douglas", got.Name)

	require.NoError(t, high.BulkSave([]testUser{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}, func(u testUser) string {
		switch u.ID {
		case 1:
			return "u1"
		default:
			return "u2"
		}
	}))

	values, err := high.BulkGet([]string{"u1", "u2", "missing"})
	require.NoError(t, err)
	require.Len(t, values, 2)
}
