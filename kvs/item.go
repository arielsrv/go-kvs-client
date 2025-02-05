package kvs

import "encoding/json"

type Item struct {
	Value any
	Key   string
	TTL   int
}

func NewItem(key string, value any, ttl ...int) *Item {
	item := &Item{
		Key:   key,
		Value: value,
	}

	if len(ttl) > 0 {
		item.TTL = ttl[0]
	}

	return item
}

func (r Item) TryGetValueAsObjectType(out any) error {
	value, ok := r.Value.(string)
	if !ok {
		return ErrMarshal
	}

	err := json.Unmarshal([]byte(value), out)
	if err != nil {
		return err
	}

	return nil
}
