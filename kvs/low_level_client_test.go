package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
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
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	items := new(kvs.Items)
	items.Add(kvs.NewItem("key1", "value1"))
	items.Add(kvs.NewItem("key2", "value2"))

	err := proxy.BulkSave(items)
	require.NoError(t, err)

	bulkItems, err := proxy.BulkGet([]string{"key1", "key2", "key3"})
	require.NoError(t, err)
	require.Equal(t, 2, bulkItems.Len())
}

func TestLowLevelClientProxy_BulkSave(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")

	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)

	items := new(kvs.Items)
	items.Add(kvs.NewItem("key1", "value1"))
	items.Add(kvs.NewItem("key2", "value2"))

	err := proxy.BulkSave(items)
	require.NoError(t, err)
}

func TestLowLevelClientProxy_ContainerName(t *testing.T) {
	// Arrange
	const container = "__kvs-test-container"
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), container)

	// Act
	proxy := kvs.NewLowLevelClientProxy(lowLevelClient)
	name := proxy.ContainerName()

	// Assert
	require.Equal(t, container, name)
}
