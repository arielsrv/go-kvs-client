package infrastructure

import (
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
)

type Client[T any] interface {
	Get(key string) (*T, error)
	Save(key string, item *T) error
}

type KVSClient[T any] struct {
	kvsClient *dynamodb.Client
}

func NewKVSClient[T any]() *KVSClient[T] {
	kvsClient := dynamodb.NewBuilder(
		dynamodb.WithTTL(7*24*60*60), // 7 dias (hh dd mm ss)
		dynamodb.WithContainerName("users"),
	).Build()

	return &KVSClient[T]{
		kvsClient: kvsClient,
	}
}

func (r KVSClient[T]) Get(key string) (*T, error) {
	item, err := r.kvsClient.Get(key)
	if err != nil {
		return nil, err
	}

	value := new(T)
	err = item.TryGetValueAsObjectType(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r KVSClient[T]) Save(key string, value *T) error {
	item := kvs.NewItem(key, value)
	err := r.kvsClient.Save(key, item)
	if err != nil {
		return err
	}

	return nil
}
