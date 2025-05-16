// Package kvs provides a generic key-value store client interface and implementation.
package kvs

import (
	"encoding/json"
	"time"
)

// Item represents a key-value pair in the store.
// It contains the key, the value, and an optional TTL (Time To Live).
type Item struct {
	// Value is the data stored in the item. It can be of any type.
	Value any
	// Key is the unique identifier for the item in the store.
	Key string
	// TTL is the Unix timestamp when the item will expire.
	// If zero, the item does not expire.
	TTL int64
}

// NewItem creates a new Item with the specified key and value.
// Optional TTL (Time To Live) in seconds can be provided to automatically expire the item.
// The TTL is converted to a Unix timestamp by adding it to the current time.
// Returns a pointer to the new Item.
func NewItem(key string, value any, ttl ...int64) *Item {
	item := &Item{
		Key:   key,
		Value: value,
	}

	if len(ttl) > 0 {
		now := time.Now().Unix() + ttl[0]
		item.TTL = now
	}

	return item
}

// TryGetValueAsObjectType attempts to convert the item's value to the type of the provided output parameter.
// The value is expected to be a JSON string that can be unmarshalled into the output parameter.
// Returns ErrConvert if the value is not a string, or ErrMarshal if unmarshalling fails.
func (r Item) TryGetValueAsObjectType(out any) error {
	value, ok := r.Value.(string)
	if !ok {
		return ErrConvert
	}

	err := json.Unmarshal([]byte(value), out)
	if err != nil {
		return ErrMarshal
	}

	return nil
}
