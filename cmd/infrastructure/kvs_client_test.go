package infrastructure_test

import (
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/infrastructure"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/model"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestKVSClient_SaveAndGet(t *testing.T) {
	lowLevelClient := dynamodb.NewLowLevelClient(dynamodb.NewAWSFakeClient(), "kvs-client")
	kvsClient := infrastructure.NewKVSClient[model.UserDTO](lowLevelClient)

	if err := kvsClient.Save("1", &model.UserDTO{ID: 1, Name: "John Doe"}); err != nil {
		require.NoError(t, err)
	}

	userDTO, err := kvsClient.Get("1")
	require.NoError(t, err)
	require.NotNil(t, userDTO)
	require.Equal(t, 1, userDTO.ID)
	require.Equal(t, "John Doe", userDTO.Name)
}
