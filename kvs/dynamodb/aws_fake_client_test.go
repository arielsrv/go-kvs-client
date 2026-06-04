package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
)

// The AWSFakeClient is exercised end-to-end through the LowLevelClient tests,
// but several error branches (defensive type assertions, missing container,
// nil PutRequest, etc.) require crafting raw DynamoDB inputs. This file does
// exactly that to drive those branches and lift the package coverage.

const fakeTableName = "__kvs-test"

func newFake() *dynamodb.AWSFakeClient { return dynamodb.NewAWSFakeClient() }

func TestAWSFakeClient_PutItem_NonStringKey_ReturnsErrConvert(t *testing.T) {
	fake := newFake()

	_, err := fake.PutItem(context.Background(), &awsdynamodb.PutItemInput{
		TableName: aws.String(fakeTableName),
		Item: map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberN{Value: "1"}, // wrong type
			"value": &types.AttributeValueMemberS{Value: "v"},
		},
	})
	require.ErrorIs(t, err, kvs.ErrConvert)
}

func TestAWSFakeClient_PutItem_NonStringValue_ReturnsErrConvert(t *testing.T) {
	fake := newFake()

	_, err := fake.PutItem(context.Background(), &awsdynamodb.PutItemInput{
		TableName: aws.String(fakeTableName),
		Item: map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberS{Value: "k"},
			"value": &types.AttributeValueMemberN{Value: "1"}, // wrong type
		},
	})
	require.ErrorIs(t, err, kvs.ErrConvert)
}

func TestAWSFakeClient_GetItem_NonStringKey_ReturnsErrConvert(t *testing.T) {
	fake := newFake()

	_, err := fake.GetItem(context.Background(), &awsdynamodb.GetItemInput{
		TableName: aws.String(fakeTableName),
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	require.ErrorIs(t, err, kvs.ErrConvert)
}

func TestAWSFakeClient_BatchGetItem_UnknownTable_ReturnsErrNilItem(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchGetItem(context.Background(), &awsdynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			"some-other-table": {Keys: nil},
		},
	})
	require.ErrorIs(t, err, kvs.ErrNilItem)
}

func TestAWSFakeClient_BatchGetItem_MissingKeyAttribute_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchGetItem(context.Background(), &awsdynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			fakeTableName: {
				Keys: []map[string]types.AttributeValue{
					{"not-key": &types.AttributeValueMemberS{Value: "x"}},
				},
			},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestAWSFakeClient_BatchGetItem_NonStringKey_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchGetItem(context.Background(), &awsdynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			fakeTableName: {
				Keys: []map[string]types.AttributeValue{
					{"key": &types.AttributeValueMemberN{Value: "1"}},
				},
			},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestAWSFakeClient_BatchGetItem_MissingKeys_AreSkippedNotErrored(t *testing.T) {
	fake := newFake()

	out, err := fake.BatchGetItem(context.Background(), &awsdynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			fakeTableName: {
				Keys: []map[string]types.AttributeValue{
					{"key": &types.AttributeValueMemberS{Value: "missing"}},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Empty(t, out.Responses[fakeTableName])
}

func TestAWSFakeClient_BatchWriteItem_UnknownTable_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchWriteItem(context.Background(), &awsdynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			"some-other-table": {},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestAWSFakeClient_BatchWriteItem_NilPutRequest_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchWriteItem(context.Background(), &awsdynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			fakeTableName: {{PutRequest: nil}},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestAWSFakeClient_BatchWriteItem_NonStringKey_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchWriteItem(context.Background(), &awsdynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			fakeTableName: {{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"key":   &types.AttributeValueMemberN{Value: "1"},
						"value": &types.AttributeValueMemberS{Value: "v"},
					},
				},
			}},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

func TestAWSFakeClient_BatchWriteItem_NonStringValue_ReturnsErrInternal(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchWriteItem(context.Background(), &awsdynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			fakeTableName: {{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"key":   &types.AttributeValueMemberS{Value: "k"},
						"value": &types.AttributeValueMemberN{Value: "1"},
					},
				},
			}},
		},
	})
	require.ErrorIs(t, err, kvs.ErrInternal)
}

// hugeValue returns a string large enough to be rejected by the freecache
// backend that powers AWSFakeClient (per-entry limit ~= 64KB). 256KB is well
// above any plausible threshold.
func hugeValue() string {
	const size = 256 * 1024
	b := make([]byte, size)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func TestAWSFakeClient_PutItem_CacheSetError_Propagates(t *testing.T) {
	fake := newFake()

	_, err := fake.PutItem(context.Background(), &awsdynamodb.PutItemInput{
		TableName: aws.String(fakeTableName),
		Item: map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberS{Value: "k"},
			"value": &types.AttributeValueMemberS{Value: hugeValue()},
		},
	})
	require.Error(t, err)
}

func TestAWSFakeClient_BatchWriteItem_CacheSetError_Propagates(t *testing.T) {
	fake := newFake()

	_, err := fake.BatchWriteItem(context.Background(), &awsdynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			fakeTableName: {{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"key":   &types.AttributeValueMemberS{Value: "k"},
						"value": &types.AttributeValueMemberS{Value: hugeValue()},
					},
				},
			}},
		},
	})
	require.Error(t, err)
}
