package kvs

// IsValidKey checks if a key is valid for use with the KVS client.
// A key is considered valid if it is not empty.
func IsValidKey(key string) bool {
	return key != ""
}
