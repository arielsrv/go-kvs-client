package dynamodb_test

import (
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestNewResolver(t *testing.T) {
	t.Skip("troubleshooting")
	resolver := dynamodb.NewResolver("http://0.0.0.0:4566")

	endpoint, err := resolver.ResolveEndpoint(t.Context(), aws.EndpointParameters{})
	require.NoError(t, err)
	require.Equal(t, "http://0.0.0.0:4566", endpoint.URI.String())
}

func TestNewResolver_Err(t *testing.T) {
	t.Skip("troubleshooting")
	resolver := dynamodb.NewResolver("::::")

	endpoint, err := resolver.ResolveEndpoint(t.Context(), aws.EndpointParameters{})
	require.Error(t, err)
	require.Equal(t, smithyendpoints.Endpoint{}, endpoint)
}
