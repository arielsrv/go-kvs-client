// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"golang.org/x/sync/singleflight"
)

// LowLevelClient is a client for interacting with AWS DynamoDB.
// It implements the kvs.LowLevelClient interface, providing methods for getting and saving items.
// The client uses a singleflight.Group to deduplicate concurrent reads for the same key.
type LowLevelClient struct {
	AWSClient // Embedded AWS DynamoDB client

	read      singleflight.Group // Group for deduplicating concurrent reads
	tableName string             // Name of the DynamoDB table
	ttl       time.Duration      // Default Time To Live for items in seconds
}

// NewLowLevelClient creates a new LowLevelClient with the provided AWS client and container name.
// The container name is used as the base for the DynamoDB table name.
// Optional TTL (Time To Live) in seconds can be provided to set a default TTL for items.
// Returns a pointer to the new LowLevelClient.
func NewLowLevelClient(awsClient AWSClient, containerName string, ttl ...time.Duration) *LowLevelClient {
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

func (r *LowLevelClient) TableName() string {
	return r.tableName
}

func (r *LowLevelClient) TTL() time.Duration {
	return r.ttl
}

// Constants for DynamoDB attribute names.
const (
	KeyName   = "key"   // Attribute name for the item's key
	ValueName = "value" // Attribute name for the item's value
	TTLName   = "ttl"   // Attribute name for the item's TTL
)

// getTableName returns the full name of the DynamoDB table.
// The table name is prefixed with "__kvs-" followed by the container name.
func (r *LowLevelClient) getTableName() *string {
	return aws.String(r.tableName)
}

// Get retrieves an item by its key.
// It uses a background context and delegates to GetWithContext.
// Returns the item if found, or an error if not found or if retrieval fails.
func (r *LowLevelClient) Get(key string) (*kvs.Item, error) {
	return r.GetWithContext(context.Background(), key)
}

// GetWithContext retrieves an item by its key using the provided context.
// The context can be used for cancellation and timeouts.
// This method uses a singleflight to deduplicate concurrent reads for the same key.
// Returns the item if found, or an error if not found or if retrieval fails.
func (r *LowLevelClient) GetWithContext(ctx context.Context, key string) (*kvs.Item, error) {
	if strings.TrimSpace(key) == "" {
		return nil, kvs.ErrEmptyKey
	}

	result, err, _ := r.read.Do(key, func() (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("GetWithContext: operation cancelled or timed out: %w", ctx.Err())
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

// Save stores an item with the specified key.
// It uses a background context and delegates to SaveWithContext.
// Returns an error if the save operation fails.
func (r *LowLevelClient) Save(key string, kvsItem *kvs.Item) error {
	return r.SaveWithContext(context.Background(), key, kvsItem)
}

// SaveWithContext stores an item with the specified key using the provided context.
// The context can be used for cancellation and timeouts.
// If the item has no TTL and the client has a default TTL, the default TTL is applied.
// The item is marshaled to JSON and stored in DynamoDB.
// Returns an error if the key is empty, the item is nil, marshaling fails, or the save operation fails.
func (r *LowLevelClient) SaveWithContext(ctx context.Context, key string, item *kvs.Item) error {
	if strings.TrimSpace(key) == "" {
		return kvs.ErrEmptyKey
	}

	if item == nil {
		return kvs.ErrNilItem
	}

	if r.ttl > 0 && item.TTL == 0 {
		item.TTL = time.Now().Add(r.ttl).Unix()
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

// BulkGet retrieves multiple items by their keys.
// It uses a background context and delegates to BulkGetWithContext.
// Returns a collection of items that were found, or an error if retrieval fails.
func (r *LowLevelClient) BulkGet(keys []string) (*kvs.Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

// BulkGetWithContext retrieves multiple items by their keys using the provided context.
// The context can be used for cancellation and timeouts.
// This method has a limit of 100 keys per request.
// Returns a collection of items that were found, or an error if retrieval fails.
// If more than 100 keys are provided, ErrTooManyKeys is returned.
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

// BulkSave stores multiple items.
// It uses a background context and delegates to BulkSaveWithContext.
// Returns an error if the save operation fails.
func (r *LowLevelClient) BulkSave(items *kvs.Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

// BulkSaveWithContext stores multiple items using the provided context.
// The context can be used for cancellation and timeouts.
// Each item is marshalled to JSON and stored in DynamoDB.
// If marshalling of an individual item fails, it is skipped and an error is logged.
// Returns an error if the batch write operation fails.
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

// ContainerName returns the name of the container or service that this client interacts with.
// Used for metrics and logging.
func (r *LowLevelClient) ContainerName() string {
	return r.tableName
}

// newItem creates a new DynamoDB item from a KVS item and its marshalled value.
// The item is represented as a map of attribute names to attribute values.
// The key, value, and TTL are stored as attributes.
func (r *LowLevelClient) newItem(item *kvs.Item, bytes []byte) map[string]types.AttributeValue {
	attributes := map[string]types.AttributeValue{}
	attributes[KeyName] = &types.AttributeValueMemberS{Value: item.Key}
	attributes[ValueName] = &types.AttributeValueMemberS{Value: string(bytes)}
	if item.TTL > 0 {
		attributes[TTLName] = &types.AttributeValueMemberN{Value: strconv.FormatInt(item.TTL, 10)}
	}
	return attributes
}
