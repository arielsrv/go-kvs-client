package redis_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"

	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"

	goredis "github.com/redis/go-redis/v9"
)

// contextWithCancel returns a context bound to the test lifetime that can be
// cancelled manually.
func contextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithCancel(t.Context())
}

// startMiniredisAt wraps an already-running miniredis server with a
// GoRedisClient. The underlying go-redis pool is closed at test end.
func startMiniredisAt(t *testing.T, srv *miniredis.Miniredis) (*miniredis.Miniredis, *kvsredis.GoRedisClient) {
	t.Helper()

	universal := goredis.NewUniversalClient(&goredis.UniversalOptions{
		Addrs: []string{srv.Addr()},
	})
	t.Cleanup(func() { _ = universal.Close() })

	return srv, kvsredis.NewGoRedisClient(universal)
}
