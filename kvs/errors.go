package kvs

import "errors"

var (
	ErrEmptyKey    = errors.New("[kvs]: key cannot be empty")
	ErrKeyNotFound = errors.New("[kvs]: key not found")
	ErrNilItem     = errors.New("[kvs]: item cannot be nil")
	ErrConvert     = errors.New("[kvs]: failed to convert item")
	ErrTooManyKeys = errors.New("[kvs]: too many keys")
)
