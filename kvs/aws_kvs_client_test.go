package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
	"github.com/arielsrv/go-kvs-client/kvs/model"
)

// AWSKVSClient was the real implementation up to v1.0.2 and is kept as a plain
// alias of KVSClient so that code written against that API keeps compiling and
// interoperating without conversions. These assignments stop compiling the day
// the alias is turned into a distinct defined type, which is exactly the
// breaking change the deprecation notice defers to the next major release.
var (
	_ *kvs.KVSClient[model.UserDTO]    = (*kvs.AWSKVSClient[model.UserDTO])(nil)
	_ *kvs.AWSKVSClient[model.UserDTO] = (*kvs.KVSClient[model.UserDTO])(nil)
)

func TestNewAWSKVSClient_SaveAndGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "__kvs_test")
	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](lowLevelClient)

	userDTO := model.NewUserDTO("John", "Doe")
	userDTO.ID = 1

	require.NoError(t, kvsClient.Save("1", userDTO))

	got, err := kvsClient.Get("1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 1, got.ID)
	require.Equal(t, "John", got.FirstName)
	require.Equal(t, "Doe", got.LastName)
}
