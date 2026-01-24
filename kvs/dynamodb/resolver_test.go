package dynamodb_test

import (
	"fmt"
	"net"
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
)

func TestNewResolver(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	addr, ok := listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)

	port := addr.Port
	err = listener.Close()
	require.NoError(t, err)

	address := fmt.Sprintf("http://127.0.0.1:%d", port)
	resolver := dynamodb.NewResolver(address)

	endpoint, err := resolver.ResolveEndpoint(t.Context(), aws.EndpointParameters{})
	require.NoError(t, err)
	require.Equal(t, address, endpoint.URI.String())
}

func TestNewResolver_Err(t *testing.T) {
	resolver := dynamodb.NewResolver("::::")

	endpoint, err := resolver.ResolveEndpoint(t.Context(), aws.EndpointParameters{})
	require.Error(t, err)
	require.Equal(t, smithyendpoints.Endpoint{}, endpoint)
}
