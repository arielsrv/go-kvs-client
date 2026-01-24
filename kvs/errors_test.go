package kvs_test

import (
	"testing"

	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/stretchr/testify/require"
)

func TestNewKeyValueError(t *testing.T) {
	err := kvs.NewKeyValueError("test error, 123")
	if err.Error() != "test error, 123" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}

	var keyValueError kvs.KeyValueError
	require.ErrorAs(t, err, &keyValueError)
}
