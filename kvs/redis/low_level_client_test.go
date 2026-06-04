package redis_test

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

type testUser struct {
	Name string
	ID   int
}

func newClient(t *testing.T, opts ...kvsredis.BuilderOptions) *kvsredis.LowLevelClient {
	t.Helper()
	return kvsredis.NewBuilder(opts...).FakeBuild()
}

func TestLowLevelClient_SaveAndGet(t *testing.T) {
	client := newClient(t, kvsredis.WithKeyPrefix("__kvs:test"))

	user := testUser{ID: 1, Name: "John Doe"}
	item := kvs.NewItem("1", user)

	require.NoError(t, client.Save("1", item))

	got, err := client.Get("1")
	require.NoError(t, err)
	require.Equal(t, "1", got.Key)

	out := new(testUser)
	require.NoError(t, got.TryGetValueAsObjectType(&out))
	require.Equal(t, user.ID, out.ID)
	require.Equal(t, user.Name, out.Name)
}

func TestLowLevelClient_Get_KeyNotFound(t *testing.T) {
	client := newClient(t)

	item, err := client.Get("missing")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, item)
}

func TestLowLevelClient_Get_EmptyKey(t *testing.T) {
	client := newClient(t)

	item, err := client.Get("")
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
	require.Nil(t, item)
}

func TestLowLevelClient_Save_EmptyKey(t *testing.T) {
	client := newClient(t)

	err := client.Save("", kvs.NewItem("", testUser{}))
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
}

func TestLowLevelClient_Save_NilItem(t *testing.T) {
	client := newClient(t)

	err := client.Save("k", nil)
	require.ErrorIs(t, err, kvs.ErrNilItem)
}

func TestLowLevelClient_Save_WithTTL_Expires(t *testing.T) {
	client := newClient(t)

	// NOTE: kvs.Item.TTL is stored as a Unix timestamp with second precision,
	// so we need a TTL of at least 2 seconds to ensure the value is observable
	// before expiration regardless of sub-second clock alignment.
	item := kvs.NewItem("k", testUser{ID: 1, Name: "x"}, 2*time.Second)
	require.NoError(t, client.Save("k", item))

	got, err := client.Get("k")
	require.NoError(t, err)
	require.NotNil(t, got)

	time.Sleep(2500 * time.Millisecond)

	got, err = client.Get("k")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, got)
}

func TestLowLevelClient_Save_TTLAlreadyExpired_IsNoOp(t *testing.T) {
	client := newClient(t)

	// Build an item whose TTL is already in the past.
	item := &kvs.Item{
		Key:   "k",
		Value: testUser{ID: 1, Name: "x"},
		TTL:   time.Now().Add(-time.Hour).Unix(),
	}
	require.NoError(t, client.Save("k", item))

	_, err := client.Get("k")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
}

func TestLowLevelClient_BulkSaveAndBulkGet(t *testing.T) {
	client := newClient(t, kvsredis.WithKeyPrefix("__kvs:bulk"))

	items := new(kvs.Items)
	items.Add(kvs.NewItem("1", testUser{ID: 1, Name: "John"}))
	items.Add(kvs.NewItem("2", testUser{ID: 2, Name: "Alice"}))
	items.Add(kvs.NewItem("3", testUser{ID: 3, Name: "Bob"}))

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

func TestLowLevelClient_BulkGet_TooManyKeys(t *testing.T) {
	client := newClient(t)

	keys := make([]string, kvsredis.MaxBulkKeys+1)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	items, err := client.BulkGet(keys)
	require.ErrorIs(t, err, kvs.ErrTooManyKeys)
	require.Nil(t, items)
}

func TestLowLevelClient_BulkGet_Empty(t *testing.T) {
	client := newClient(t)

	items, err := client.BulkGet(nil)
	require.NoError(t, err)
	require.NotNil(t, items)
	require.Equal(t, 0, items.Len())
}

func TestLowLevelClient_BulkSave_NilOrEmpty_IsNoOp(t *testing.T) {
	client := newClient(t)

	require.NoError(t, client.BulkSave(nil))
	require.NoError(t, client.BulkSave(new(kvs.Items)))
}

func TestLowLevelClient_BulkSave_SkipsInvalidItems(t *testing.T) {
	client := newClient(t)

	items := new(kvs.Items)
	items.Add(kvs.NewItem("", testUser{ID: 0}))              // empty key
	items.Add(nil)                                           // nil item
	items.Add(kvs.NewItem("ok", testUser{ID: 1, Name: "n"})) // valid

	require.NoError(t, client.BulkSave(items))

	got, err := client.Get("ok")
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestLowLevelClient_KeyPrefix_IsApplied(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	client := kvsredis.NewLowLevelClient(fake, "__kvs:scoped")

	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))

	keys := fake.Keys("__kvs:scoped:")
	require.Len(t, keys, 1)
	require.Equal(t, "__kvs:scoped:k", keys[0])
}

func TestLowLevelClient_NoKeyPrefix(t *testing.T) {
	fake := kvsredis.NewFakeClient()
	client := kvsredis.NewLowLevelClient(fake, "")

	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))
	keys := fake.Keys("")
	require.Len(t, keys, 1)
	require.Equal(t, "k", keys[0])
	require.Equal(t, "redis", client.ContainerName())
}

func TestLowLevelClient_ContainerName_WithPrefix(t *testing.T) {
	client := kvsredis.NewLowLevelClient(kvsredis.NewFakeClient(), "__kvs:users:")
	require.Equal(t, "__kvs:users", client.ContainerName())
}

func TestLowLevelClient_GetWithContext_Cancelled(t *testing.T) {
	client := newClient(t)
	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))

	ctx, cancel := contextWithCancel(t)
	cancel()

	got, err := client.GetWithContext(ctx, "k")
	require.Error(t, err)
	require.Nil(t, got)
}

func TestLowLevelClient_Close(t *testing.T) {
	client := newClient(t)
	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))
	require.NoError(t, client.Close())
}

func TestLowLevelClient_IntegratesWithGenericClient(t *testing.T) {
	ll := newClient(t, kvsredis.WithKeyPrefix("__kvs:gen"), kvsredis.WithTTL(time.Hour))
	high := kvs.NewKVSClient[testUser](ll)

	require.NoError(t, high.Save("42", &testUser{ID: 42, Name: "Douglas"}))

	got, err := high.Get("42")
	require.NoError(t, err)
	require.Equal(t, 42, got.ID)
	require.Equal(t, "Douglas", got.Name)

	require.NoError(t, high.BulkSave([]testUser{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}, func(u testUser) string { return strconv.Itoa(u.ID) }))

	values, err := high.BulkGet([]string{"1", "2"})
	require.NoError(t, err)
	require.Len(t, values, 2)
}
