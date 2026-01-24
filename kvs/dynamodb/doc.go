// Package dynamodb provides an AWS DynamoDB implementation of the kvs.KVSClient interface.
//
// This package contains the necessary components to interact with AWS DynamoDB as a key-value store.
// It implements the generic interface defined in the parent kvs package, allowing for seamless
// integration with DynamoDB while maintaining the type-safety and convenience of the KVS client.
//
// Key Components:
//   - LowLevelClient: Handles direct interactions with DynamoDB
//   - AWSClient: Interface for AWS SDK operations
//   - Builder: Constructs DynamoDB-specific requests
//   - Resolver: Processes DynamoDB responses into application objects
//
// The package supports all operations defined in the kvs.KVSClient interface, including:
//   - Individual and bulk item retrieval
//   - Individual and bulk item storage
//   - Context-aware operations for proper timeout and cancellation handling
//   - TTL (Time To Live) settings for automatic item expiration
//
// Usage:
//
// To use the DynamoDB implementation, you typically create a client through the parent kvs package:
//
//	import (
//	    "github.com/aws/aws-sdk-go-v2/config"
//	    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
//	    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
//	)
//
//	// Load AWS configuration
//	cfg, err := config.LoadDefaultConfig(context.TODO())
//	if err != nil {
//	    // Handle error
//	}
//
//	// Create DynamoDB client
//	dynamoClient := dynamodb.NewFromConfig(cfg)
//
//	// Create KVS client with DynamoDB backend
//	client := kvs.NewAWSKVSClient[MyType](
//	    kvs.WithAWSClient(dynamoClient),
//	    kvs.WithTableName("my-table"),
//	)
//
// The package also provides testing utilities like AWS fake clients for unit testing.
// Cambiar referencia de import en doc.go
package dynamodb
