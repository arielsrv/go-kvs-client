package kvs

type Items struct {
	Items []*Item
}

func (r Items) GetOks() []*Item {
	return r.Items
}
