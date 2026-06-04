package redis_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

// erroringClient is a minimal Client implementation that returns the same
// error from every operation, useful to drive the error branches of
// LowLevelClient without depending on the real Redis driver.
type erroringClient struct{ err error }

func (c *erroringClient) Get(_ context.Context, _ string) (string, error) {
	return "", c.err
}

func (c *erroringClient) Set(_ context.Context, _, _ string, _ time.Duration) error {
	return c.err
}

func (c *erroringClient) MGet(_ context.Context, _ []string) ([]kvsredis.GetResult, error) {
	return nil, c.err
}

func (c *erroringClient) MSet(_ context.Context, _ []kvsredis.Pair) error {
	return c.err
}

func (c *erroringClient) Close() error { return nil }

func TestLowLevelClient_Get_PropagatesClientError(t *testing.T) {
	want := errors.New("boom")
	client := kvsredis.NewLowLevelClient(&erroringClient{err: want}, "p")

	_, err := client.Get("k")
	require.ErrorIs(t, err, want)
}

func TestLowLevelClient_BulkGet_PropagatesClientError(t *testing.T) {
	want := errors.New("boom")
	client := kvsredis.NewLowLevelClient(&erroringClient{err: want}, "p")

	items, err := client.BulkGet([]string{"a", "b"})
	require.ErrorIs(t, err, want)
	require.Nil(t, items)
}

func TestLowLevelClient_BulkSave_PropagatesClientError(t *testing.T) {
	want := errors.New("boom")
	client := kvsredis.NewLowLevelClient(&erroringClient{err: want}, "p")

	items := new(kvs.Items)
	items.Add(kvs.NewItem("k", "v"))

	err := client.BulkSave(items)
	require.ErrorIs(t, err, want)
}

func TestLowLevelClient_Save_MarshalError_IsReportedAsKVSError(t *testing.T) {
	client := newClient(t)

	// A function value cannot be marshalled to JSON.
	item := kvs.NewItem("k", func() {})
	err := client.Save("k", item)
	require.Error(t, err)
	require.Contains(t, err.Error(), "marshal")
}

func TestLowLevelClient_BulkSave_SkipsUnmarshalableItems(t *testing.T) {
	// One item is non-marshalable (a chan), the other is fine; only the
	// good one must be persisted.
	client := newClient(t)

	items := new(kvs.Items)
	items.Add(kvs.NewItem("bad", make(chan int)))
	items.Add(kvs.NewItem("good", testUser{ID: 7, Name: "ok"}))

	require.NoError(t, client.BulkSave(items))

	got, err := client.Get("good")
	require.NoError(t, err)
	require.Equal(t, "good", got.Key)

	_, err = client.Get("bad")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestFakeClient_MSet_PropagatesSetError(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	require.NoError(t, fake.Close())

	err := fake.MSet(context.Background(), []kvsredis.Pair{{Key: "k", Value: "v"}})
	require.Error(t, err)
}

func TestFakeClient_Keys_FiltersByPrefix(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	ctx := context.Background()

	require.NoError(t, fake.Set(ctx, "users:1", "a", 0))
	require.NoError(t, fake.Set(ctx, "users:2", "b", 0))
	require.NoError(t, fake.Set(ctx, "carts:1", "c", 0))

	got := fake.Keys("users:")
	require.Len(t, got, 2)
	for _, k := range got {
		require.True(t, strings.HasPrefix(k, "users:"))
	}

	require.Len(t, fake.Keys(""), 3)
}
