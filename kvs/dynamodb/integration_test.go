//go:build integration

package dynamodb_test

import (
	"context"
	"net"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"

	"github.com/arielsrv/go-kvs-client/kvs"
	kvsdynamo "github.com/arielsrv/go-kvs-client/kvs/dynamodb"
)

const integrationTableName = "kvs-integration-test"

func setupLocalStackDynamoDB(t *testing.T) *kvsdynamo.LowLevelClient {
	t.Helper()

	ctx := context.Background()

	container, err := localstack.Run(ctx, "localstack/localstack:3.8.1")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testcontainers.TerminateContainer(container)
	})

	mappedPort, err := container.MappedPort(ctx, "4566/tcp")
	require.NoError(t, err)

	provider, err := testcontainers.NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	require.NoError(t, err)

	endpoint := "http://" + net.JoinHostPort(host, mappedPort.Port())

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	)
	require.NoError(t, err)

	dynamoClient := awsdynamodb.NewFromConfig(awsCfg, func(o *awsdynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	_, err = dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
		TableName: aws.String(integrationTableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String(kvsdynamo.KeyName), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String(kvsdynamo.KeyName), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	require.NoError(t, err)

	return kvsdynamo.NewLowLevelClient(dynamoClient, integrationTableName)
}

func TestIntegration_DynamoDB_SaveAndGet(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	type payload struct {
		ID   int
		Name string
	}

	in := payload{ID: 1, Name: "John Doe"}
	item := kvs.NewItem("1", in)

	err := client.Save("1", item)
	require.NoError(t, err)

	got, err := client.Get("1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "1", got.Key)

	out := new(payload)
	require.NoError(t, got.TryGetValueAsObjectType(&out))
	require.Equal(t, in.ID, out.ID)
	require.Equal(t, in.Name, out.Name)
}

func TestIntegration_DynamoDB_Get_KeyNotFound(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	got, err := client.Get("does-not-exist")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, got)
}

func TestIntegration_DynamoDB_Get_EmptyKey(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	got, err := client.Get("")
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
	require.Nil(t, got)
}

func TestIntegration_DynamoDB_Save_EmptyKey(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	err := client.Save("", kvs.NewItem("", struct{}{}))
	require.ErrorIs(t, err, kvs.ErrEmptyKey)
}

func TestIntegration_DynamoDB_Save_NilItem(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	err := client.Save("k", nil)
	require.ErrorIs(t, err, kvs.ErrNilItem)
}

func TestIntegration_DynamoDB_BulkSaveAndBulkGet(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	type payload struct {
		ID   int
		Name string
	}

	items := new(kvs.Items)
	items.Add(kvs.NewItem("10", payload{ID: 10, Name: "Alice"}))
	items.Add(kvs.NewItem("11", payload{ID: 11, Name: "Bob"}))
	items.Add(kvs.NewItem("12", payload{ID: 12, Name: "Charlie"}))

	require.NoError(t, client.BulkSave(items))

	got, err := client.BulkGet([]string{"10", "11", "12"})
	require.NoError(t, err)
	require.Len(t, slices.Collect(got.All()), 3)

	for current := range got.All() {
		out := new(payload)
		require.NoError(t, current.TryGetValueAsObjectType(&out))
		require.NotEmpty(t, out.Name)
	}
}

func TestIntegration_DynamoDB_BulkGet_MissingKeysSkipped(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	type payload struct{ ID int }

	items := new(kvs.Items)
	items.Add(kvs.NewItem("20", payload{ID: 20}))
	require.NoError(t, client.BulkSave(items))

	got, err := client.BulkGet([]string{"20", "missing-key"})
	require.NoError(t, err)
	require.Len(t, slices.Collect(got.All()), 1)
}

func TestIntegration_DynamoDB_BulkGet_TooManyKeys(t *testing.T) {
	client := setupLocalStackDynamoDB(t)

	keys := make([]string, 101)
	for i := range keys {
		keys[i] = "k"
	}

	got, err := client.BulkGet(keys)
	require.ErrorIs(t, err, kvs.ErrTooManyKeys)
	require.Nil(t, got)
}

func TestIntegration_DynamoDB_ContainerName(t *testing.T) {
	client := setupLocalStackDynamoDB(t)
	require.Equal(t, integrationTableName, client.ContainerName())
}

func TestIntegration_DynamoDB_TableName(t *testing.T) {
	client := setupLocalStackDynamoDB(t)
	require.Equal(t, integrationTableName, client.TableName())
}
