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
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

type AWSFakeClient struct {
	cache cache.CacheInterface[[]byte]
}

func NewAWSFakeClient() *AWSFakeClient {
	cacheStore := freecachestore.NewFreecache(freecache.NewCache(math.MaxInt32))

	return &AWSFakeClient{
		cache: cache.New[[]byte](cacheStore),
	}
}

func (r AWSFakeClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
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

func (r AWSFakeClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
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

func (r AWSFakeClient) BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
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

		batchGetItemOutput.Responses[r.getContainerName()] = append(batchGetItemOutput.Responses[r.getContainerName()], []map[string]types.AttributeValue{
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

func (r AWSFakeClient) getContainerName() string {
	return "__kvs-test"
}

func (r AWSFakeClient) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
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
