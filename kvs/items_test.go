package kvs_test

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

func TestItems_Add(t *testing.T) {
	items := new(kvs.Items)

	item1 := kvs.NewItem("key1", "value1")
	items.Add(item1)
	item2 := kvs.NewItem("key2", "value2")
	items.Add(item2)

	require.Equal(t, 2, items.Len())

	current, stop := iter.Pull(items.All())
	defer stop()

	item, hasNext := current()
	require.True(t, hasNext)
	require.Equal(t, "key1", item.Key)
	require.Equal(t, "value1", item.Value)

	item, hasNext = current()
	require.True(t, hasNext)
	require.Equal(t, "key2", item.Key)
	require.Equal(t, "value2", item.Value)
}
