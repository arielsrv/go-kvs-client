package kvs_test

import (
	"strconv"
	"testing"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/model"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestKVSClient_SaveAndGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs_test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

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
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

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

func TestKVSClient_GetWithContext(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs_test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	userDTO := model.NewUserDTO("John", "Doe")
	userDTO.ID = 1

	// Save the user
	err := kvsClient.Save("1", userDTO)
	require.NoError(t, err)

	// Test with a valid context
	ctx := t.Context()
	retrievedUser, err := kvsClient.GetWithContext(ctx, "1")
	require.NoError(t, err)
	require.NotNil(t, retrievedUser)
	require.Equal(t, 1, retrievedUser.ID)
	require.Equal(t, "John", retrievedUser.FirstName)
	require.Equal(t, "Doe", retrievedUser.LastName)
	require.Equal(t, "John Doe", retrievedUser.FullName)

	// Note: The fake AWS client doesn't properly handle canceled contexts,
	// so we're not testing that case here. In a real environment,
	// a canceled context would result in an error.

	// Test with a non-existent key
	_, err = kvsClient.GetWithContext(ctx, "non-existent")
	require.Error(t, err)
	require.Equal(t, kvs.ErrKeyNotFound, err)
}

func TestKVSClient_BulkGetWithContext(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	users := []model.UserDTO{
		{ID: 1, FirstName: "John Doe"},
		{ID: 2, FirstName: "Alice Doe"},
	}

	// Save the users
	err := kvsClient.BulkSave(users, func(item model.UserDTO) string {
		return strconv.Itoa(item.ID)
	})
	require.NoError(t, err)

	// Test with a valid context
	ctx := t.Context()
	result, err := kvsClient.BulkGetWithContext(ctx, []string{"1", "2"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)
	require.Equal(t, 1, result[0].ID)
	require.Equal(t, "John Doe", result[0].FirstName)
	require.Equal(t, 2, result[1].ID)
	require.Equal(t, "Alice Doe", result[1].FirstName)

	// Note: The fake AWS client doesn't properly handle canceled contexts,
	// so we're not testing that case here. In a real environment,
	// a canceled context would result in an error.

	// Test with non-existent keys
	result, err = kvsClient.BulkGetWithContext(ctx, []string{"3", "4"})
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestKVSClient_SaveWithContext(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs_test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	userDTO := model.NewUserDTO("John", "Doe")
	userDTO.ID = 1

	// Test with a valid context
	ctx := t.Context()
	err := kvsClient.SaveWithContext(ctx, "1", userDTO)
	require.NoError(t, err)

	// Verify the save worked
	retrievedUser, err := kvsClient.Get("1")
	require.NoError(t, err)
	require.NotNil(t, retrievedUser)
	require.Equal(t, 1, retrievedUser.ID)

	// Note: The fake AWS client doesn't properly handle canceled contexts,
	// so we're not testing that case here. In a real environment,
	// a canceled context would result in an error.

	// Test with TTL
	ttlCtx := t.Context()
	err = kvsClient.SaveWithContext(ttlCtx, "3", userDTO, time.Hour)
	require.NoError(t, err)

	// Verify the save with TTL worked
	retrievedUser, err = kvsClient.Get("3")
	require.NoError(t, err)
	require.NotNil(t, retrievedUser)
	require.Equal(t, 1, retrievedUser.ID)
}

func TestKVSClient_BulkSaveWithContext(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs-test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	users := []model.UserDTO{
		{ID: 1, FirstName: "John Doe"},
		{ID: 2, FirstName: "Alice Doe"},
	}

	keyMapper := func(item model.UserDTO) string {
		return strconv.Itoa(item.ID)
	}

	// Test with a valid context
	ctx := t.Context()
	err := kvsClient.BulkSaveWithContext(ctx, users, keyMapper)
	require.NoError(t, err)

	// Verify the save worked
	result, err := kvsClient.BulkGet([]string{"1", "2"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)

	// Note: The fake AWS client doesn't properly handle canceled contexts,
	// so we're not testing that case here. In a real environment,
	// a canceled context would result in an error.

	// Test with TTL
	ttlCtx := t.Context()
	err = kvsClient.BulkSaveWithContext(ttlCtx, users, keyMapper, time.Hour)
	require.NoError(t, err)

	// Verify the save with TTL worked
	result, err = kvsClient.BulkGet([]string{"1", "2"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)
}
