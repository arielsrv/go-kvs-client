package dynamodb_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
)

type Test struct {
	Name string
	ID   int
}

func TestClient_SaveAndGet(t *testing.T) {
	t.Skip()
	lowLevelClient := kvs.NewLowLevelClientProxy(dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test"))

	input := struct {
		Key   string
		Value Test
	}{
		Key: "1",
		Value: Test{
			ID:   1,
			Name: "John Doe",
		},
	}

	item := kvs.NewItem(input.Key, input.Value)

	err := lowLevelClient.Save(input.Key, item)
	require.NoError(t, err)

	actual, err := lowLevelClient.Get("1")
	require.NoError(t, err)

	actualValue := new(Test)
	err = actual.TryGetValueAsObjectType(&actualValue)
	require.NoError(t, err)

	require.Equal(t, input.Value.ID, actualValue.ID)
	require.Equal(t, input.Value.Name, actualValue.Name)
	require.Equal(t, input.Key, actual.Key)
}

func TestClient_Get_KeyNotFound(t *testing.T) {
	kvsClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	item, err := kvsClient.Get("1")
	require.Error(t, err)
	require.Equal(t, kvs.ErrKeyNotFound, err)
	require.Nil(t, item)
}

func TestClient_Get_EmptyKey(t *testing.T) {
	kvsClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")

	item, err := kvsClient.Get("")
	require.Error(t, err)
	require.Equal(t, kvs.ErrEmptyKey, err)
	require.Nil(t, item)
}

func TestClient_BulkSave_And_BulkGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")

	input1 := struct {
		Key   string
		Value Test
	}{
		Key: "1",
		Value: Test{
			ID:   1,
			Name: "John Doe",
		},
	}

	item1 := kvs.NewItem(input1.Key, input1.Value)

	input2 := struct {
		Key   string
		Value Test
	}{
		Key: "2",
		Value: Test{
			ID:   2,
			Name: "Alice Doe",
		},
	}

	item2 := kvs.NewItem(input2.Key, input2.Value)

	items := new(kvs.Items)
	items.Add(item1)
	items.Add(item2)

	err := lowLevelClient.BulkSave(items)
	require.NoError(t, err)

	actual, err := lowLevelClient.BulkGet([]string{"1", "2"})
	require.NoError(t, err)

	for current := range actual.All() {
		actualValue := new(Test)
		err = current.TryGetValueAsObjectType(&actualValue)
		require.NoError(t, err)
	}

	require.Len(t, slices.Collect(actual.All()), 2)
}

func TestEmptyKeyInGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")

	item, err := lowLevelClient.Get("")
	require.Error(t, err)
	require.Equal(t, kvs.ErrEmptyKey, err)
	require.Nil(t, item)

	err = lowLevelClient.Save("", kvs.NewItem("", Test{}))
	require.Error(t, err)
	require.Equal(t, kvs.ErrEmptyKey, err)

	err = lowLevelClient.Save("1", nil)
	require.Error(t, err)
	require.Equal(t, kvs.ErrNilItem, err)
}
