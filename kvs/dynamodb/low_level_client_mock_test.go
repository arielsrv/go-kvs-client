package dynamodb_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
	mockdb "github.com/arielsrv/go-kvs-client/resources/mocks/kvs/dynamodb"
)

// Sentinel error returned by mocked AWS calls to drive the error branches of
// the LowLevelClient.
var errBoom = errors.New("boom")

// matchAny is a convenience expecter matcher.
func matchAny() any { return mock.Anything }

func TestLowLevelClient_ContainerName_ReturnsTableName(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	client := dynamodb.NewLowLevelClient(awsMock, "my-table")
	require.Equal(t, "my-table", client.ContainerName())
}

func TestLowLevelClient_GetWithContext_GetItemError_Propagates(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		GetItem(matchAny(), matchAny()).
		Return(nil, errBoom).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	item, err := client.Get("k")
	require.ErrorIs(t, err, errBoom)
	require.Nil(t, item)
}

func TestLowLevelClient_GetWithContext_ContextCancelled_ReturnsCtxError(t *testing.T) {
	// The select{} inside GetWithContext checks ctx.Done() before issuing the
	// GetItem request, so we never expect the mock to be called.
	awsMock := mockdb.NewMockAWSClient(t)
	client := dynamodb.NewLowLevelClient(awsMock, "t")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	item, err := client.GetWithContext(ctx, "k")
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, item)
}

func TestLowLevelClient_SaveWithContext_PutItemError_Propagates(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		PutItem(matchAny(), matchAny()).
		Return(nil, errBoom).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	err := client.Save("k", kvs.NewItem("k", "v"))
	require.ErrorIs(t, err, errBoom)
}

func TestLowLevelClient_SaveWithContext_MarshalError(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	// PutItem should never be invoked because json.Marshal fails first.
	client := dynamodb.NewLowLevelClient(awsMock, "t")

	// Channels cannot be marshalled as JSON.
	err := client.Save("k", kvs.NewItem("k", make(chan int)))
	require.Error(t, err)
}

func TestLowLevelClient_SaveWithContext_AppliesDefaultTTL(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		PutItem(matchAny(), mock.MatchedBy(func(in *awsdynamodb.PutItemInput) bool {
			// The default TTL configured in the LowLevelClient must propagate
			// into a "ttl" attribute on the persisted item.
			ttl, ok := in.Item["ttl"].(*types.AttributeValueMemberN)
			if !ok {
				return false
			}
			expiresAt, parseErr := strconv.ParseInt(ttl.Value, 10, 64)
			if parseErr != nil {
				return false
			}
			// The TTL must be in the future (allow some slack).
			return expiresAt > time.Now().Unix()
		})).
		Return(&awsdynamodb.PutItemOutput{}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t", time.Hour)
	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))
}

func TestLowLevelClient_SaveWithContext_NoTTL_OmitsTTLAttribute(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		PutItem(matchAny(), mock.MatchedBy(func(in *awsdynamodb.PutItemInput) bool {
			_, hasTTL := in.Item["ttl"]
			return !hasTTL
		})).
		Return(&awsdynamodb.PutItemOutput{}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t") // no default TTL
	require.NoError(t, client.Save("k", kvs.NewItem("k", "v")))
}

func TestLowLevelClient_BulkGetWithContext_BatchGetItemError_Propagates(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		BatchGetItem(matchAny(), matchAny()).
		Return(nil, errBoom).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	items, err := client.BulkGet([]string{"a", "b"})
	require.ErrorIs(t, err, errBoom)
	require.Nil(t, items)
}

func TestLowLevelClient_BulkGetWithContext_TooManyKeys(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	client := dynamodb.NewLowLevelClient(awsMock, "t")

	keys := make([]string, 101)
	for i := range keys {
		keys[i] = strings.Repeat("k", i+1)
	}
	items, err := client.BulkGet(keys)
	require.ErrorIs(t, err, kvs.ErrTooManyKeys)
	require.Nil(t, items)
}

func TestLowLevelClient_BulkSaveWithContext_BatchWriteItemError_Propagates(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		BatchWriteItem(matchAny(), matchAny()).
		Return(nil, errBoom).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	items := new(kvs.Items)
	items.Add(kvs.NewItem("k", "v"))

	err := client.BulkSave(items)
	require.ErrorIs(t, err, errBoom)
}

func TestLowLevelClient_BulkSaveWithContext_SkipsUnmarshalableItems(t *testing.T) {
	// One of the items has a non-marshalable value (a channel). The
	// BatchWriteItem call must therefore be invoked with exactly one item
	// (the good one).
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		BatchWriteItem(matchAny(), mock.MatchedBy(func(in *awsdynamodb.BatchWriteItemInput) bool {
			for _, requests := range in.RequestItems {
				if len(requests) != 1 {
					return false
				}
			}
			return true
		})).
		Return(&awsdynamodb.BatchWriteItemOutput{}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	items := new(kvs.Items)
	items.Add(kvs.NewItem("bad", make(chan int)))
	items.Add(kvs.NewItem("good", "v"))

	require.NoError(t, client.BulkSave(items))
}

func TestLowLevelClient_GetWithContext_NilItem_ReturnsErrKeyNotFound(t *testing.T) {
	// DynamoDB GetItem returns a non-nil output with a nil Item map when the
	// key does not exist. The LowLevelClient must translate that into the
	// canonical kvs.ErrKeyNotFound sentinel.
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		GetItem(matchAny(), matchAny()).
		Return(&awsdynamodb.GetItemOutput{Item: nil}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	item, err := client.Get("k")
	require.ErrorIs(t, err, kvs.ErrKeyNotFound)
	require.Nil(t, item)
}

func TestLowLevelClient_GetWithContext_UnmarshalError_Propagates(t *testing.T) {
	// Returning an Item whose attribute values do not match the target Item
	// struct (e.g. a boolean where a string is expected) forces
	// attributevalue.UnmarshalMap to fail.
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		GetItem(matchAny(), matchAny()).
		Return(&awsdynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"key":   &types.AttributeValueMemberBOOL{Value: true},
				"value": &types.AttributeValueMemberBOOL{Value: false},
			},
		}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	item, err := client.Get("k")
	require.Error(t, err)
	require.Nil(t, item)
}

func TestLowLevelClient_BulkGetWithContext_UnmarshalError_Propagates(t *testing.T) {
	awsMock := mockdb.NewMockAWSClient(t)
	awsMock.EXPECT().
		BatchGetItem(matchAny(), matchAny()).
		Return(&awsdynamodb.BatchGetItemOutput{
			Responses: map[string][]map[string]types.AttributeValue{
				"t": {
					{
						"key":   &types.AttributeValueMemberBOOL{Value: true},
						"value": &types.AttributeValueMemberBOOL{Value: false},
					},
				},
			},
		}, nil).
		Once()

	client := dynamodb.NewLowLevelClient(awsMock, "t")

	items, err := client.BulkGet([]string{"a"})
	require.Error(t, err)
	require.Nil(t, items)
}
