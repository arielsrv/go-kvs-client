package infrastructure_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/examples/infrastructure"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/examples/model"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestKVSClient_SaveAndGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "kvs-client")
	kvsClient := infrastructure.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	userDTO := model.NewUserDTO("John", "Doe")
	userDTO.ID = 1

	if err := kvsClient.Save("1", userDTO); err != nil {
		require.NoError(t, err)
	}

	userDTO, err := kvsClient.Get("1")
	require.NoError(t, err)
	require.NotNil(t, userDTO)
	require.Equal(t, 1, userDTO.ID)
	require.Equal(t, "John", userDTO.FirstName)
	require.Equal(t, "Doe", userDTO.LastName)
	require.Equal(t, "John Doe", userDTO.FullName)
}

func TestKVSClient_BulkSaveAndBulkGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "test")
	kvsClient := infrastructure.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	users := []model.UserDTO{
		{ID: 1, FirstName: "John Doe"},
		{ID: 2, FirstName: "Alice Doe"},
	}

	if err := kvsClient.BulkSave(users, func(item model.UserDTO) string {
		return strconv.Itoa(item.ID)
	}); err != nil {
		require.NoError(t, err)
	}

	result, err := kvsClient.BulkGet([]string{"1", "2"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)
	require.Equal(t, 1, result[0].ID)
	require.Equal(t, "John Doe", result[0].FirstName)
	require.Equal(t, 2, result[1].ID)
	require.Equal(t, "Alice Doe", result[1].FirstName)
}
