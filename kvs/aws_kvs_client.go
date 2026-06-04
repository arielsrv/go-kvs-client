package kvs

// AWSKVSClient is a backward-compatible alias for KVSClient.
//
// Deprecated: the implementation is backend-agnostic (it works with DynamoDB,
// Redis, or any other LowLevelClient). Use KVSClient[T] instead. This alias
// will be removed in a future major release.
type AWSKVSClient[T any] = KVSClient[T]

// NewAWSKVSClient is a backward-compatible constructor for KVSClient.
//
// Deprecated: use NewKVSClient instead. This wrapper will be removed in a
// future major release.
func NewAWSKVSClient[T any](lowLevelClient LowLevelClient) *KVSClient[T] {
	return NewKVSClient[T](lowLevelClient)
}
