package kvs

const (
	ErrKeyNotFound = KeyValueError("[kvs]: key not found")
	ErrEmptyKey    = KeyValueError("[kvs]: key cannot be empty")
	ErrNilItem     = KeyValueError("[kvs]: item cannot be nil")
	ErrMarshal     = KeyValueError("[kvs]: failed to convert item")
	ErrTooManyKeys = KeyValueError("[kvs]: too many keys")
	ErrInternal    = KeyValueError("[kvs]: internal error")
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
