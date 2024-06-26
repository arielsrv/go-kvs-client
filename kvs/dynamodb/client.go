package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

type Client struct {
	LowLevelClient
	containerName string
	ttl           int
}

const (
	KeyName   = "key"
	ValueName = "value"
)

func NewClient(lowLevelClient LowLevelClient, containerName string, ttl ...int) *Client {
	client := &Client{
		containerName:  containerName,
		LowLevelClient: lowLevelClient,
	}

	log.Debugf("[kvs]: setting container name to %s", containerName)

	if len(ttl) > 0 {
		log.Debugf("[kvs]: setting TTL to %d", ttl[0])
		client.ttl = ttl[0]
	}

	return client
}

func (r Client) getTableName() *string {
	return aws.String(fmt.Sprintf("__kvs-%s", r.containerName))
}

func (r Client) Get(key string) (*kvs.Item, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r Client) GetWithContext(ctx context.Context, key string) (*kvs.Item, error) {
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

	getItemOutput, err := r.LowLevelClient.GetItem(ctx, input)
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

func (r Client) Save(key string, kvsItem *kvs.Item) error {
	return r.SaveWithContext(context.Background(), key, kvsItem)
}

func (r Client) SaveWithContext(ctx context.Context, key string, item *kvs.Item) error {
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

	putItemOutput, err := r.LowLevelClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.getTableName(),
		Item: map[string]types.AttributeValue{
			KeyName: &types.AttributeValueMemberS{
				Value: key,
			},
			ValueName: &types.AttributeValueMemberS{
				Value: string(bytes),
			},
		},
	})

	if err != nil {
		return err
	}

	log.Debugf("[kvs]: putItemOutput: %+v", putItemOutput)

	return nil
}

func (r Client) BulkGet(keys []string) (*kvs.Items, error) {
	// TODO implement me
	panic("implement me")
}

func (r Client) BulkGetWithContext(ctx context.Context, key []string) (*kvs.Items, error) {
	// TODO implement me
	panic("implement me")
}

func (r Client) BulkSave(items *kvs.Items) error {
	// TODO implement me
	panic("implement me")
}

func (r Client) BulkSaveWithContext(ctx context.Context, items *kvs.Items) error {
	// TODO implement me
	panic("implement me")
}
