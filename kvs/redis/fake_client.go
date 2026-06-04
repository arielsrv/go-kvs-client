// Package redis provides a Redis implementation of the KVS client.
package redis

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/arielsrv/go-kvs-client/kvs"
)

// FakeClient is an in-memory implementation of Client used for unit tests
// and for the Builder.FakeBuild constructor. It honours TTL semantics
// (entries are evicted lazily on access) and is safe for concurrent use.
type FakeClient struct {
	now     func() time.Time
	entries map[string]fakeEntry
	mu      sync.RWMutex
	closed  bool
}

type fakeEntry struct {
	expiresAt time.Time // zero means "no expiration"
	value     string
}

// NewFakeClient returns a fresh FakeClient backed by an internal map.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		entries: make(map[string]fakeEntry),
		now:     time.Now,
	}
}

// Get implements Client.
func (r *FakeClient) Get(_ context.Context, key string) (string, error) {
	r.mu.RLock()
	entry, ok := r.entries[key]
	closed := r.closed
	r.mu.RUnlock()

	if closed {
		return "", kvs.ErrInternal
	}
	if !ok || r.expired(entry) {
		// Best-effort lazy eviction.
		if ok {
			r.mu.Lock()
			if e, stillThere := r.entries[key]; stillThere && r.expired(e) {
				delete(r.entries, key)
			}
			r.mu.Unlock()
		}
		return "", kvs.ErrKeyNotFound
	}
	return entry.value, nil
}

// Set implements Client.
func (r *FakeClient) Set(_ context.Context, key, value string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return kvs.ErrInternal
	}

	entry := fakeEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = r.now().Add(ttl)
	}
	r.entries[key] = entry
	return nil
}

// MGet implements Client.
func (r *FakeClient) MGet(ctx context.Context, keys []string) ([]GetResult, error) {
	results := make([]GetResult, len(keys))
	for i, key := range keys {
		value, err := r.Get(ctx, key)
		switch {
		case err == nil:
			results[i] = GetResult{Key: key, Value: value, Found: true}
		case errors.Is(err, kvs.ErrKeyNotFound):
			results[i] = GetResult{Key: key, Found: false}
		default:
			return nil, err
		}
	}
	return results, nil
}

// MSet implements Client.
func (r *FakeClient) MSet(ctx context.Context, pairs []Pair) error {
	for _, pair := range pairs {
		if err := r.Set(ctx, pair.Key, pair.Value, pair.TTL); err != nil {
			return err
		}
	}
	return nil
}

// Close implements Client.
func (r *FakeClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	r.entries = nil
	return nil
}

// Len returns the number of entries currently stored (including expired ones
// that have not yet been evicted). Exposed mainly for tests.
func (r *FakeClient) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Keys returns the set of keys whose prefix matches the provided prefix.
// Useful for testing key prefixing/namespacing.
func (r *FakeClient) Keys(prefix string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.entries))
	for k := range r.entries {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out
}

func (r *FakeClient) expired(entry fakeEntry) bool {
	if entry.expiresAt.IsZero() {
		return false
	}
	return !r.now().Before(entry.expiresAt)
}
