package kvs

import "encoding/json"

type Item struct {
	Value any
	Key   string
	TTL   int
}

func NewItem(key string, value any) *Item {
	return &Item{
		Key:   key,
		Value: value,
	}
}

func (r Item) TryGetValueAsObjectType(out any) error {
	err := json.Unmarshal([]byte(r.Value.(string)), out)
	if err != nil {
		return err
	}

	return nil
}
