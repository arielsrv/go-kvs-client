package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

type Test struct {
	ID   int
	Name string
}

func TestClient_SaveAndGet(t *testing.T) {
	kvsClient := dynamodb.NewClient(dynamodb.NewLowLevelMockClient(), "test", 60)

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

	err := kvsClient.Save(input.Key, item)
	require.NoError(t, err)

	actual, err := kvsClient.Get("1")
	require.NoError(t, err)

	actualValue := new(Test)
	err = actual.TryGetValueAsObjectType(&actualValue)
	require.NoError(t, err)

	require.Equal(t, input.Value.ID, actualValue.ID)
	require.Equal(t, input.Value.Name, actualValue.Name)
	require.Equal(t, input.Key, actual.Key)
}

func TestClient_Get_ErrKeyNotFound(t *testing.T) {
	kvsClient := dynamodb.NewClient(dynamodb.NewLowLevelMockClient(), "test")

	item, err := kvsClient.Get("1")
	require.Error(t, err)
	require.Equal(t, kvs.ErrKeyNotFound, err)
	require.Nil(t, item)
}

func TestClient_Get_ErrEmptyKey(t *testing.T) {
	kvsClient := dynamodb.NewClient(dynamodb.NewLowLevelMockClient(), "test")

	item, err := kvsClient.Get("")
	require.Error(t, err)
	require.Equal(t, kvs.ErrEmptyKey, err)
	require.Nil(t, item)
}
