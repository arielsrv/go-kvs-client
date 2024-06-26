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

type LowLevelMockClient struct {
	cache cache.CacheInterface[[]byte]
}

func NewLowLevelMockClient() *LowLevelMockClient {
	cacheStore := freecachestore.NewFreecache(freecache.NewCache(math.MaxInt32))
	return &LowLevelMockClient{
		cache: cache.New[[]byte](cacheStore),
	}
}

func (r LowLevelMockClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
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

func (r LowLevelMockClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
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
