// Package kvs provides a generic key-value store client interface and implementation.
package kvs

import (
	"iter"
	"slices"
)

// Iterable is an interface for types that can provide a sequence of items.
type Iterable interface {
	// All returns a sequence of all items in the collection.
	All() iter.Seq[*Item]
}

// List is an interface for types that can store and provide access to a collection of items.
type List interface {
	// Add adds an item to the list.
	Add(item *Item)
	// Len returns the number of items in the list.
	Len() int
	// All returns a sequence of all items in the list.
	All() iter.Seq[*Item]
}

// Items is a collection of Item objects that implements the List and Iterable interfaces.
type Items struct {
	items []*Item
}

// Add adds an item to the collection.
func (r *Items) Add(item *Item) {
	r.items = append(r.items, item)
}

// Len returns the number of items in the collection.
func (r *Items) Len() int {
	return len(r.items)
}

// All returns a sequence of all items in the collection.
// This allows iterating over the items using the iter package.
func (r *Items) All() iter.Seq[*Item] {
	return slices.Values(r.items)
}
