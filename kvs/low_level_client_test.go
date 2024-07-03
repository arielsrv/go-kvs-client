package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestLowLevelClientProxy_Get(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	item, err := proxy.Get("key")
	require.Error(t, err)
	require.Equal(t, kvs.ErrKeyNotFound, err)
	require.Nil(t, item)
}

func TestLowLevelClientProxy_Save(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	err := proxy.Save("key", kvs.NewItem("key", kvs.NewItem("key", "value")))
	require.NoError(t, err)

	item, err := proxy.Get("key")
	require.NoError(t, err)
	require.NotNil(t, item)
}

func TestLowLevelClientProxy_BulkGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	items := &kvs.Items{
		Items: []*kvs.Item{
			kvs.NewItem("key1", "value1"),
			kvs.NewItem("key2", "value2"),
		},
	}

	err := proxy.BulkSave(items)
	require.NoError(t, err)

	bulkItems, err := proxy.BulkGet([]string{"key1", "key2", "key3"})
	require.NoError(t, err)
	require.Len(t, bulkItems.Items, 2)
}

func TestLowLevelClientProxy_BulkSave(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	items := &kvs.Items{
		Items: []*kvs.Item{
			kvs.NewItem("key1", "value1"),
			kvs.NewItem("key2", "value2"),
		},
	}

	err := proxy.BulkSave(items)
	require.NoError(t, err)
}
