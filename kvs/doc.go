// Package kvs provides a generic key-value store client interface and implementation.
//
// The kvs package offers a flexible and type-safe way to interact with key-value stores
// through a generic interface. It supports various operations such as Get, Save, BulkGet,
// and BulkSave with optional context and TTL (Time To Live) settings.
//
// Key Features:
//   - Generic interface that works with any data type
//   - Support for individual and bulk operations
//   - Context support for cancellation and timeouts
//   - Optional TTL for automatic expiration of items
//   - AWS DynamoDB implementation
//
// Basic Usage:
//
//	// Create a new KVS client for a specific type
//	client := kvs.NewAWSKVSClient[MyType](options...)
//
//	// Get an item
//	item, err := client.Get("myKey")
//
//	// Save an item
//	err := client.Save("myKey", &myItem)
//
//	// Get multiple items
//	items, err := client.BulkGet([]string{"key1", "key2"})
//
//	// Save multiple items
//	err := client.BulkSave(items, func(item MyType) string {
//	    return item.ID // Extract key from item
//	})
//
// The package also provides context-aware methods for all operations,
// allowing for proper timeout and cancellation handling.
package kvs
