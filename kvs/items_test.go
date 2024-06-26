package kvs_test

import (
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
	require.Equal(t, item1, items.GetOks()[0])
	require.Equal(t, item2, items.GetOks()[1])
}
