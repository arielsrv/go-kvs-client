package redis_test

import (
	"context"
	"testing"
)

// contextWithCancel returns a context bound to the test lifetime that can be
// cancelled manually.
func contextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithCancel(t.Context())
}
