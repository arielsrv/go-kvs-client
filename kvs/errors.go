package kvs

const (
	ErrKeyNotFound = KeyValueError("[kvs]: key not found")
	ErrEmptyKey    = KeyValueError("[kvs]: key cannot be empty")
	ErrNilItem     = KeyValueError("[kvs]: item cannot be nil")
	ErrMarshal     = KeyValueError("[kvs]: failed to convert item")
	ErrTooManyKeys = KeyValueError("[kvs]: too many keys")
	ErrInternal    = KeyValueError("[kvs]: internal error")
)

type KeyValueError string

func (r KeyValueError) Error() string {
	return string(r)
}
