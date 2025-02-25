package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sync/singleflight"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

type LowLevelClient struct {
	AWSClient
	read      singleflight.Group
	tableName string
	ttl       int64
}

const (
	KeyName   = "key"
	ValueName = "value"
	TTLName   = "ttl"
)

func NewLowLevelClient(awsClient AWSClient, containerName string, ttl ...int64) *LowLevelClient {
	lowLevelClient := &LowLevelClient{
		tableName: containerName,
		AWSClient: awsClient,
	}

	log.Debugf("[kvs]: setting container name to %s", containerName)

	if len(ttl) > 0 {
		log.Debugf("[kvs]: setting TTL to %d", ttl[0])
		lowLevelClient.ttl = ttl[0]
	}

	return lowLevelClient
}

func (r *LowLevelClient) getTableName() *string {
	return aws.String(fmt.Sprintf("__kvs-%s", r.tableName))
}

func (r *LowLevelClient) Get(key string) (*kvs.Item, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r *LowLevelClient) GetWithContext(ctx context.Context, key string) (*kvs.Item, error) {
	if strings.TrimSpace(key) == "" {
		return nil, kvs.ErrEmptyKey
	}

	result, err, _ := r.read.Do(key, func() (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			input := &dynamodb.GetItemInput{
				TableName: r.getTableName(),
				Key: map[string]types.AttributeValue{
					KeyName: &types.AttributeValueMemberS{
						Value: key,
					},
				},
			}

			getItemOutput, err := r.AWSClient.GetItem(ctx, input)
			if err != nil {
				return nil, err
			} else if getItemOutput.Item == nil {
				return nil, kvs.ErrKeyNotFound
			}

			var item Item
			err = attributevalue.UnmarshalMap(getItemOutput.Item, &item)
			if err != nil {
				return nil, err
			}

			return &kvs.Item{
				Key:   item.Key,
				Value: item.Value,
				TTL:   item.TTL,
			}, nil
		}
	})
	if err != nil {
		return nil, err
	}

	return result.(*kvs.Item), nil
}

func (r *LowLevelClient) Save(key string, kvsItem *kvs.Item) error {
	return r.SaveWithContext(context.Background(), key, kvsItem)
}

func (r *LowLevelClient) SaveWithContext(ctx context.Context, key string, item *kvs.Item) error {
	if strings.TrimSpace(key) == "" {
		return kvs.ErrEmptyKey
	}

	if item == nil {
		return kvs.ErrNilItem
	}

	if r.ttl > 0 && item.TTL == 0 {
		item.TTL = r.ttl
	}

	bytes, err := json.Marshal(item.Value)
	if err != nil {
		return err
	}

	putItemOutput, err := r.AWSClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.getTableName(),
		Item:      r.newItem(item, bytes),
	})
	if err != nil {
		return err
	}

	log.Debugf("[kvs]: putItemOutput: %+v", putItemOutput)

	return nil
}

func (r *LowLevelClient) BulkGet(keys []string) (*kvs.Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

func (r *LowLevelClient) BulkGetWithContext(ctx context.Context, keys []string) (*kvs.Items, error) {
	if len(keys) > 100 {
		return nil, kvs.ErrTooManyKeys
	}

	inputKeys := make([]map[string]types.AttributeValue, len(keys))
	for i := range keys {
		inputKeys[i] = map[string]types.AttributeValue{
			KeyName: &types.AttributeValueMemberS{
				Value: keys[i],
			},
		}
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			aws.ToString(r.getTableName()): {
				Keys: inputKeys,
			},
		},
	}

	batchGetItemOutput, err := r.AWSClient.BatchGetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	items := new(kvs.Items)

	for key, value := range batchGetItemOutput.Responses {
		log.Debugf("[kvs]: key: %v", key)

		var item []Item
		err = attributevalue.UnmarshalListOfMaps(value, &item)
		if err != nil {
			log.Errorf("[kvs]: error unmarshalling value: %v", err)
			return nil, err
		}

		for i := range item {
			items.Add(&kvs.Item{
				Key:   item[i].Key,
				Value: item[i].Value,
				TTL:   item[i].TTL,
			})
		}
	}

	log.Debugf("[kvs]: batchGetItemOutput: %+v", batchGetItemOutput)

	return items, nil
}

func (r *LowLevelClient) BulkSave(items *kvs.Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

func (r *LowLevelClient) BulkSaveWithContext(ctx context.Context, kvsItems *kvs.Items) error {
	items := make([]types.WriteRequest, 0, kvsItems.Len())

	for item := range kvsItems.All() {
		bytes, err := json.Marshal(item.Value)
		if err != nil {
			log.Errorf("[kvs]: error marshalling value: %v", err)
			continue
		}

		items = append(items, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: r.newItem(item, bytes),
			},
		})
	}

	batchInput := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			aws.ToString(r.getTableName()): items,
		},
	}

	batchWriteItemOutput, err := r.AWSClient.BatchWriteItem(ctx, batchInput)
	if err != nil {
		log.Errorf("[kvs]: error writing kvsItems to DynamoDB: %v", err)
		return err
	}

	log.Debugf("[kvs]: batchWriteItemOutput: %+v", batchWriteItemOutput)

	return nil
}

func (r *LowLevelClient) ContainerName() string {
	return r.tableName
}

func (r *LowLevelClient) newItem(item *kvs.Item, bytes []byte) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		KeyName: &types.AttributeValueMemberS{
			Value: item.Key,
		},
		ValueName: &types.AttributeValueMemberS{
			Value: string(bytes),
		},
		TTLName: &types.AttributeValueMemberN{
			Value: strconv.FormatInt(item.TTL, 10),
		},
	}
}
