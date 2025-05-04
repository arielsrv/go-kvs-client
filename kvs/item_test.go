package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

func TestNewItem(t *testing.T) {
	item := kvs.NewItem("key", `{"name": "value"}`)

	var out map[string]interface{}
	err := item.TryGetValueAsObjectType(&out)

	require.NoError(t, err)
	require.Equal(t, "value", out["name"])
	require.Equal(t, "key", item.Key)
	assert.JSONEq(t, `{"name": "value"}`, item.Value.(string))
}

func TestNewItem_Err(t *testing.T) {
	item := kvs.NewItem("key", `invalid`)

	var out map[string]interface{}
	err := item.TryGetValueAsObjectType(&out)

	require.Error(t, err)
	var keyValueError kvs.KeyValueError
	require.ErrorAs(t, err, &keyValueError)
}
