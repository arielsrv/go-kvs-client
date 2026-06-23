//go:build integration

package redis_test

import (
	"context"
	"slices"
	"strconv"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

func setupRedisClient(t *testing.T, opts ...kvsredis.BuilderOptions) *kvsredis.LowLevelClient {
	t.Helper()

	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testcontainers.TerminateContainer(container)
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	opt, err := goredis.ParseURL(connStr)
	require.NoError(t, err)

	underlying := goredis.NewClient(opt)
	t.Cleanup(func() { _ = underlying.Close() })

	b := kvsredis.NewBuilder(opts...)
	return b.BuildWithClient(kvsredis.NewGoRedisClient(underlying))
}

func TestIntegration_Redis_SaveAndGet(t *testing.T) {
	client := setupRedisClient(t, kvsredis.WithKeyPrefix("__kvs:integration"))

	user := testUser{ID: 42, Name: "John Doe"}
	item := kvs.NewItem("42", user)

	require.NoError(t, client.Save("42", item))

	got, err := client.Get("42")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "42", got.Key)

	out := new(testUser)
	require.NoError(t, got.TryGetValueAsObjectType(&out))
	require.Equal(t, user.ID, out.ID)
	require.Equal(t, user.Name, out.Name)
}

func TestIntegration_Redis_Get_KeyNotFound(t *testing.T) {
	client := setupRedisClient(t)

	got, err := client.Get("does-not-exist")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, got)
}

func TestIntegration_Redis_Get_EmptyKey(t *testing.T) {
	client := setupRedisClient(t)

	got, err := client.Get("")
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
	require.Nil(t, got)
}

func TestIntegration_Redis_Save_EmptyKey(t *testing.T) {
	client := setupRedisClient(t)

	err := client.Save("", kvs.NewItem("", testUser{}))
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
}

func TestIntegration_Redis_Save_NilItem(t *testing.T) {
	client := setupRedisClient(t)

	err := client.Save("k", nil)
	require.ErrorIs(t, err, kvs.ErrNilItem)
}

func TestIntegration_Redis_Save_WithTTL_Expires(t *testing.T) {
	client := setupRedisClient(t)

	item := kvs.NewItem("expiring", testUser{ID: 1, Name: "TTL"}, 2*time.Second)
	require.NoError(t, client.Save("expiring", item))

	got, err := client.Get("expiring")
	require.NoError(t, err)
	require.NotNil(t, got)

	time.Sleep(3 * time.Second)

	got, err = client.Get("expiring")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, got)
}

func TestIntegration_Redis_Save_TTLAlreadyExpired_IsNoOp(t *testing.T) {
	client := setupRedisClient(t)

	item := &kvs.Item{
		Key:   "expired",
		Value: testUser{ID: 1},
		TTL:   time.Now().Add(-time.Hour).Unix(),
	}
	require.NoError(t, client.Save("expired", item))

	_, err := client.Get("expired")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestIntegration_Redis_BulkSaveAndBulkGet(t *testing.T) {
	client := setupRedisClient(t, kvsredis.WithKeyPrefix("__kvs:bulk:integration"))

	items := new(kvs.Items)
	items.Add(kvs.NewItem("1", testUser{ID: 1, Name: "Alice"}))
	items.Add(kvs.NewItem("2", testUser{ID: 2, Name: "Bob"}))
	items.Add(kvs.NewItem("3", testUser{ID: 3, Name: "Charlie"}))

	require.NoError(t, client.BulkSave(items))

	got, err := client.BulkGet([]string{"1", "2", "3", "missing"})
	require.NoError(t, err)
	require.Len(t, slices.Collect(got.All()), 3)

	for current := range got.All() {
		out := new(testUser)
		require.NoError(t, current.TryGetValueAsObjectType(&out))
		require.NotEmpty(t, out.Name)
	}
}

func TestIntegration_Redis_BulkGet_Empty(t *testing.T) {
	client := setupRedisClient(t)

	got, err := client.BulkGet(nil)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 0, got.Len())
}

func TestIntegration_Redis_BulkGet_TooManyKeys(t *testing.T) {
	client := setupRedisClient(t)

	keys := make([]string, kvsredis.MaxBulkKeys+1)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	got, err := client.BulkGet(keys)
	require.ErrorIs(t, err, kvs.ErrTooManyKeys)
	require.Nil(t, got)
}

func TestIntegration_Redis_BulkSave_NilIsNoOp(t *testing.T) {
	client := setupRedisClient(t)

	require.NoError(t, client.BulkSave(nil))
	require.NoError(t, client.BulkSave(new(kvs.Items)))
}

func TestIntegration_Redis_DefaultTTL(t *testing.T) {
	client := setupRedisClient(t, kvsredis.WithTTL(2*time.Second))

	item := kvs.NewItem("ttl-default", testUser{ID: 99, Name: "TTLDefault"})
	require.NoError(t, client.Save("ttl-default", item))

	got, err := client.Get("ttl-default")
	require.NoError(t, err)
	require.NotNil(t, got)

	time.Sleep(3 * time.Second)

	got, err = client.Get("ttl-default")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, got)
}

func TestIntegration_Redis_ContainerName(t *testing.T) {
	client := setupRedisClient(t, kvsredis.WithKeyPrefix("__kvs:ns"))
	require.Equal(t, "__kvs:ns", client.ContainerName())
}

func TestIntegration_Redis_IntegratesWithGenericClient(t *testing.T) {
	ll := setupRedisClient(t, kvsredis.WithKeyPrefix("__kvs:generic"), kvsredis.WithTTL(time.Hour))
	high := kvs.NewKVSClient[testUser](ll)

	require.NoError(t, high.Save("100", &testUser{ID: 100, Name: "Generic"}))

	got, err := high.Get("100")
	require.NoError(t, err)
	require.Equal(t, 100, got.ID)
	require.Equal(t, "Generic", got.Name)

	require.NoError(t, high.BulkSave([]testUser{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}, func(u testUser) string { return strconv.Itoa(u.ID) }))

	values, err := high.BulkGet([]string{"1", "2"})
	require.NoError(t, err)
	require.Len(t, values, 2)
}
