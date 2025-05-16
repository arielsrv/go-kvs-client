// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// AWSClient is an interface for interacting with AWS DynamoDB.
// It defines methods for basic DynamoDB operations like PutItem, GetItem, BatchGetItem, and BatchWriteItem.
// This interface allows for easier testing by mocking the AWS SDK.
type AWSClient interface {
	// PutItem puts a single item in a DynamoDB table.
	PutItem(
		ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.PutItemOutput, error)

	// GetItem retrieves a single item from a DynamoDB table.
	GetItem(
		ctx context.Context,
		params *dynamodb.GetItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.GetItemOutput, error)

	// BatchGetItem retrieves multiple items from one or more DynamoDB tables.
	BatchGetItem(
		ctx context.Context,
		params *dynamodb.BatchGetItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.BatchGetItemOutput, error)

	// BatchWriteItem puts or deletes multiple items in one or more DynamoDB tables.
	BatchWriteItem(
		ctx context.Context,
		params *dynamodb.BatchWriteItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.BatchWriteItemOutput, error)
}
