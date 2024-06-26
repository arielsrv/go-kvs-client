package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

func TestItems_Add(t *testing.T) {
	items := new(kvs.Items)
	item := kvs.NewItem("key", "value")
	items.Add(item)
	require.Equal(t, 1, items.Len())
	require.Equal(t, item, items.GetOks()[0])
}
