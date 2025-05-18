package dynamodb_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

func TestBuilder_WithOptions(t *testing.T) {
	builder := dynamodb.NewBuilder(
		dynamodb.WithContainerName("my-service"),
		dynamodb.WithEndpointResolver("http://localhost:8000"),
		dynamodb.WithTTL(5*time.Minute),
	)

	actual := builder.Build(aws.Config{})
	require.NotNil(t, actual)
	require.Equal(t, 5*time.Minute, actual.TTL())
	require.Equal(t, "my-service", actual.TableName())
}

func TestBuilder_WithFunc(t *testing.T) {
	builder := dynamodb.NewBuilder()
	builder.WithContainerName("my-service")
	builder.WithEndpointResolver("http://localhost:8000")
	builder.WithTTL(5 * time.Minute)

	actual := builder.Build(aws.Config{})
	require.NotNil(t, actual)
	require.Equal(t, 5*time.Minute, actual.TTL())
	require.Equal(t, "my-service", actual.TableName())
}
