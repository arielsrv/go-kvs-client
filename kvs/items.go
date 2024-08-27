package kvs

import (
	"iter"
	"slices"
)

type Items struct {
	items []*Item
}

func (r *Items) Add(item *Item) {
	r.items = append(r.items, item)
}

func (r *Items) Len() int {
	return len(r.items)
}

func (r *Items) All() iter.Seq[*Item] {
	return slices.Values(r.items)
}
