package kvs_test

import (
	"testing"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
)

func TestNewKeyValueError(t *testing.T) {
	err := kvs.NewKeyValueError("test error, 123")
	if err.Error() != "test error, 123" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
}
