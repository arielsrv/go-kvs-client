package kvs

type Items struct {
	Items []*Item
}

func (r *Items) Add(item *Item) {
	r.Items = append(r.Items, item)
}

func (r *Items) Len() int {
	return len(r.Items)
}

func (r *Items) GetOks() []*Item {
	return r.Items
}
