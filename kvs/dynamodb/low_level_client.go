package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

type LowLevelClient struct {
	AWSClient
	containerName string
	ttl           int
}

const (
	KeyName   = "key"
	ValueName = "value"
	TTLName   = "ttl"
)

func NewLowLevelClient(awsClient AWSClient, containerName string, ttl ...int) *LowLevelClient {
	lowLevelClient := &LowLevelClient{
		containerName: containerName,
		AWSClient:     awsClient,
	}

	log.Debugf("[kvs]: setting container name to %s", containerName)

	if len(ttl) > 0 {
		log.Debugf("[kvs]: setting TTL to %d", ttl[0])
		lowLevelClient.ttl = ttl[0]
	}

	return lowLevelClient
}

func (r LowLevelClient) getTableName() *string {
	return aws.String(fmt.Sprintf("__kvs-%s", r.containerName))
}

func (r LowLevelClient) Get(key string) (*kvs.Item, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r LowLevelClient) GetWithContext(ctx context.Context, key string) (*kvs.Item, error) {
	if strings.TrimSpace(key) == "" {
		return nil, kvs.ErrEmptyKey
	}

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
	}, nil
}

func (r LowLevelClient) Save(key string, kvsItem *kvs.Item) error {
	return r.SaveWithContext(context.Background(), key, kvsItem)
}

func (r LowLevelClient) SaveWithContext(ctx context.Context, key string, item *kvs.Item) error {
	if strings.TrimSpace(key) == "" {
		return kvs.ErrEmptyKey
	}

	if item == nil {
		return kvs.ErrNilItem
	}

	if r.ttl > 0 {
		item.TTL = r.ttl
	}

	bytes, err := json.Marshal(item.Value)
	if err != nil {
		return err
	}

	putItemOutput, err := r.AWSClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.getTableName(),
		Item:      r.createItem(item, bytes),
	})

	if err != nil {
		return err
	}

	log.Debugf("[kvs]: putItemOutput: %+v", putItemOutput)

	return nil
}

func (r LowLevelClient) BulkGet(keys []string) (*kvs.Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

func (r LowLevelClient) BulkGetWithContext(ctx context.Context, keys []string) (*kvs.Items, error) {
	if len(keys) > 100 {
		return nil, kvs.ErrTooManyKeys
	}

	inputKeys := make([]map[string]types.AttributeValue, 0)
	for i := range keys {
		inputKeys = append(inputKeys, map[string]types.AttributeValue{
			KeyName: &types.AttributeValueMemberS{
				Value: keys[i],
			},
		})
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

		var found []Item
		err = attributevalue.UnmarshalListOfMaps(value, &found)
		if err != nil {
			log.Errorf("[kvs]: error unmarshalling value: %v", err)
			return nil, err
		}

		for i := range found {
			items.Add(&kvs.Item{
				Key:   found[i].Key,
				Value: found[i].Value,
			})
		}
	}

	log.Debugf("[kvs]: batchGetItemOutput: %+v", batchGetItemOutput)

	return items, nil
}

func (r LowLevelClient) BulkSave(items *kvs.Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

func (r LowLevelClient) BulkSaveWithContext(ctx context.Context, kvsItems *kvs.Items) error {
	items := make([]types.WriteRequest, 0, kvsItems.Len())

	for i := range kvsItems.Items {
		item := kvsItems.Items[i]
		bytes, err := json.Marshal(item.Value)
		if err != nil {
			log.Errorf("[kvs]: error marshalling value: %v", err)
			continue
		}

		items = append(items, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: r.createItem(item, bytes),
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

func (r LowLevelClient) createItem(item *kvs.Item, bytes []byte) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		KeyName: &types.AttributeValueMemberS{
			Value: item.Key,
		},
		ValueName: &types.AttributeValueMemberS{
			Value: string(bytes),
		},
		TTLName: &types.AttributeValueMemberN{
			Value: strconv.Itoa(item.TTL),
		},
	}
}
