// Package kvs provides a generic key-value store client interface and implementation.
package kvs

// Error constants for common key-value store errors.
const (
	// ErrKeyNotFound is returned when a key is not found in the store.
	ErrKeyNotFound = KeyValueError("[kvs]: key not found")
	// ErrEmptyKey is returned when an empty key is provided.
	ErrEmptyKey = KeyValueError("[kvs]: key cannot be empty")
	// ErrNilItem is returned when a nil item is provided.
	ErrNilItem = KeyValueError("[kvs]: item cannot be nil")
	// ErrConvert is returned when an item's value cannot be converted to the expected type.
	ErrConvert = KeyValueError("[kvs]: failed to convert item")
	// ErrMarshal is returned when an item's value cannot be marshalled or unmarshalled.
	ErrMarshal = KeyValueError("[kvs]: failed to marshal item")
	// ErrTooManyKeys is returned when too many keys are provided for a bulk operation.
	ErrTooManyKeys = KeyValueError("[kvs]: too many keys")
	// ErrInternal is returned when an internal error occurs in the key-value store.
	ErrInternal = KeyValueError("[kvs]: internal error")
)

// KeyValueError is a custom error type for key-value store operations.
type KeyValueError string

// NewKeyValueError creates a new KeyValueError with the provided error message.
func NewKeyValueError(err string) KeyValueError {
	return KeyValueError(err)
}

// Error implements the error interface for KeyValueError.
func (r KeyValueError) Error() string {
	return string(r)
}
