// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

// Item represents a key-value pair in DynamoDB.
// It is used for marshaling and unmarshalling items to and from DynamoDB.
// The struct tags specify the attribute names in DynamoDB.
type Item struct {
	// Key is the unique identifier for the item in DynamoDB.
	Key string `dynamodbav:"key"`
	// Value is the data stored in the item, serialized as a JSON string.
	Value string `dynamodbav:"value"`
	// TTL is the Unix timestamp when the item will expire.
	// If zero, the item does not expire.
	TTL int64 `dynamodbav:"ttl"`
}
