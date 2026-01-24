// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

import (
	"context"
	"errors"
	"math"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/coocood/freecache"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	freecachestore "github.com/eko/gocache/store/freecache/v4"

	"github.com/arielsrv/go-kvs-client/kvs"
)

// AWSFakeClient is a fake implementation of the AWSClient interface for testing.
// Instead of interacting with actual DynamoDB, it uses an in-memory cache.
// This allows for testing without requiring a real DynamoDB instance.
type AWSFakeClient struct {
	cache cache.CacheInterface[[]byte] // In-memory cache for storing key-value pairs
}

// NewAWSFakeClient creates a new AWSFakeClient with an in-memory cache.
// The cache is initialized with the maximum possible size to avoid evictions.
// Returns a pointer to the new AWSFakeClient.
func NewAWSFakeClient() *AWSFakeClient {
	cacheStore := freecachestore.NewFreecache(freecache.NewCache(math.MaxInt8))

	return &AWSFakeClient{
		cache: cache.New[[]byte](cacheStore),
	}
}

// PutItem implements the AWSClient interface for storing a single item.
// It extracts the key and value from the input parameters and stores them in the cache.
// Returns an error if the key or value cannot be converted to the expected type,
// or if the cache operation fails.
func (r AWSFakeClient) PutItem(
	ctx context.Context,
	params *dynamodb.PutItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.PutItemOutput, error) {
	keyMember, convert := params.Item[KeyName].(*types.AttributeValueMemberS)
	if !convert {
		return nil, kvs.ErrConvert
	}

	valueMember, convert := params.Item[ValueName].(*types.AttributeValueMemberS)
	if !convert {
		return nil, kvs.ErrConvert
	}

	if err := r.cache.Set(ctx, keyMember.Value, []byte(valueMember.Value)); err != nil {
		return nil, err
	}

	return &dynamodb.PutItemOutput{}, nil
}

// GetItem implements the AWSClient interface for retrieving a single item.
// It extracts the key from the input parameters and retrieves the corresponding value from the cache.
// Returns the item if found, or an error if the key cannot be converted to the expected type,
// the key is not found, or the cache operation fails.
func (r AWSFakeClient) GetItem(
	ctx context.Context,
	params *dynamodb.GetItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.GetItemOutput, error) {
	keyMember, convert := params.Key[KeyName].(*types.AttributeValueMemberS)
	if !convert {
		return nil, kvs.ErrConvert
	}

	value, err := r.cache.Get(ctx, keyMember.Value)
	if err != nil {
		if errors.Is(err, &store.NotFound{}) {
			return nil, kvs.ErrKeyNotFound
		}
		return nil, err
	}

	return &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			KeyName: &types.AttributeValueMemberS{
				Value: keyMember.Value,
			},
			ValueName: &types.AttributeValueMemberS{
				Value: string(value),
			},
		},
	}, nil
}

// BatchGetItem implements the AWSClient interface for retrieving multiple items.
// It extracts the keys from the input parameters and retrieves the corresponding values from the cache.
// Returns a collection of items that were found, or an error if the keys cannot be found in the request,
// a key cannot be converted to the expected type, or a cache operation fails.
// If a key is not found in the cache, it is skipped without returning an error.
func (r AWSFakeClient) BatchGetItem(
	ctx context.Context,
	params *dynamodb.BatchGetItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.BatchGetItemOutput, error) {
	records, found := params.RequestItems[r.getContainerName()]
	if !found {
		return nil, kvs.ErrNilItem
	}

	batchGetItemOutput := new(dynamodb.BatchGetItemOutput)
	batchGetItemOutput.Responses = make(map[string][]map[string]types.AttributeValue, len(records.Keys))
	batchGetItemOutput.Responses[r.getContainerName()] = []map[string]types.AttributeValue{}

	for i := range records.Keys {
		key, exist := records.Keys[i][KeyName]
		if !exist {
			return nil, kvs.ErrInternal
		}
		keyValueMember, ok := key.(*types.AttributeValueMemberS)
		if !ok {
			return nil, kvs.ErrInternal
		}

		value, err := r.cache.Get(ctx, keyValueMember.Value)
		if err != nil {
			if errors.Is(err, &store.NotFound{}) {
				continue
			}

			return nil, err
		}

		batchGetItemOutput.Responses[r.getContainerName()] = append(
			batchGetItemOutput.Responses[r.getContainerName()],
			[]map[string]types.AttributeValue{
				{
					KeyName: &types.AttributeValueMemberS{
						Value: keyValueMember.Value,
					},
					ValueName: &types.AttributeValueMemberS{
						Value: string(value),
					},
				},
			}...)
	}
	return batchGetItemOutput, nil
}

// getContainerName returns the name of the container or service that this client interacts with.
// For the fake client, this is always "__kvs-test".
// Used for metrics and logging.
func (r AWSFakeClient) getContainerName() string {
	return "__kvs-test"
}

// BatchWriteItem implements the AWSClient interface for storing multiple items.
// It extracts the keys and values from the input parameters and stores them in the cache.
// Returns an error if the items cannot be found in the request, a key or value cannot be
// converted to the expected type, or a cache operation fails.
// Only PutRequest operations are supported; DeleteRequest operations will return an error.
func (r AWSFakeClient) BatchWriteItem(
	ctx context.Context,
	params *dynamodb.BatchWriteItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.BatchWriteItemOutput, error) {
	records, found := params.RequestItems[r.getContainerName()]
	if !found {
		return nil, kvs.ErrInternal
	}

	for i := range records {
		record := records[i]
		if record.PutRequest == nil {
			return nil, kvs.ErrInternal
		}

		keyMember, convert := record.PutRequest.Item[KeyName].(*types.AttributeValueMemberS)
		if !convert {
			return nil, kvs.ErrInternal
		}

		keyValue, convert := record.PutRequest.Item[ValueName].(*types.AttributeValueMemberS)
		if !convert {
			return nil, kvs.ErrInternal
		}

		if err := r.cache.Set(ctx, keyMember.Value, []byte(keyValue.Value)); err != nil {
			return nil, err
		}
	}

	return &dynamodb.BatchWriteItemOutput{}, nil
}
