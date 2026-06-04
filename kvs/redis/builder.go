// Package redis provides a Redis implementation of the KVS client.
package redis

import (
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

// Builder is a fluent / functional-options builder for the Redis LowLevelClient.
// It supports the three deployment topologies exposed by go-redis through
// UniversalOptions: standalone, Sentinel and Cluster.
type Builder struct {
	tlsConfig      *tls.Config
	username       string
	password       string
	keyPrefix      string
	masterName     string
	addresses      []string
	tracingOpts    []redisotel.TracingOption
	metricsOpts    []redisotel.MetricsOption
	dialTimeout    time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	ttl            time.Duration
	poolSize       int
	db             int
	routeRandom    bool
	tracingEnabled bool
	metricsEnabled bool
}

// BuilderOptions configures a Builder. Used with the functional-options pattern.
type BuilderOptions func(*Builder)

// NewBuilder returns a new Builder configured by the supplied options.
func NewBuilder(opts ...BuilderOptions) *Builder {
	builder := &Builder{}
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}

// ---------------------------------------------------------------------------
// Fluent setters
// ---------------------------------------------------------------------------

// WithAddresses sets the Redis endpoints. Pass a single address for standalone
// mode, several for Cluster or Sentinel mode.
func (r *Builder) WithAddresses(addresses ...string) *Builder {
	r.addresses = addresses
	return r
}

// WithKeyPrefix sets the namespace prepended to every key.
// Example: a prefix "__kvs:users" and a key "42" become "__kvs:users:42".
func (r *Builder) WithKeyPrefix(prefix string) *Builder {
	r.keyPrefix = prefix
	return r
}

// WithTTL sets the default TTL applied to items that do not carry one.
func (r *Builder) WithTTL(ttl time.Duration) *Builder {
	r.ttl = ttl
	return r
}

// WithUsername sets the ACL username (Redis >= 6).
func (r *Builder) WithUsername(username string) *Builder {
	r.username = username
	return r
}

// WithPassword sets the password used for AUTH.
func (r *Builder) WithPassword(password string) *Builder {
	r.password = password
	return r
}

// WithDB selects the Redis logical database (standalone only).
func (r *Builder) WithDB(db int) *Builder {
	r.db = db
	return r
}

// WithMasterName enables Sentinel discovery for the given master name.
func (r *Builder) WithMasterName(name string) *Builder {
	r.masterName = name
	return r
}

// WithTLS enables TLS using the supplied configuration.
func (r *Builder) WithTLS(cfg *tls.Config) *Builder {
	r.tlsConfig = cfg
	return r
}

// WithPoolSize sets the maximum number of socket connections per node.
func (r *Builder) WithPoolSize(size int) *Builder {
	r.poolSize = size
	return r
}

// WithTimeouts sets the dial / read / write timeouts. A zero value means
// "use the go-redis default" for that timeout.
func (r *Builder) WithTimeouts(dial, read, write time.Duration) *Builder {
	r.dialTimeout = dial
	r.readTimeout = read
	r.writeTimeout = write
	return r
}

// WithRouteRandomly distributes read-only commands across replicas at random
// (Cluster only). When false, reads always hit the master.
func (r *Builder) WithRouteRandomly(enabled bool) *Builder {
	r.routeRandom = enabled
	return r
}

// WithTracing enables OpenTelemetry tracing on the underlying Redis driver.
// Each command issued through the client will produce a span describing the
// command name, key(s), DB index and the result status.
//
// Optional redisotel.TracingOption values can be supplied to customise the
// instrumentation (custom tracer provider, attribute filters, etc.).
//
// This is the Redis equivalent of otelaws.AppendMiddlewares for the DynamoDB
// backend; together with WithMetrics it gives the same observability story
// regardless of the chosen provider.
func (r *Builder) WithTracing(opts ...redisotel.TracingOption) *Builder {
	r.tracingEnabled = true
	r.tracingOpts = opts
	return r
}

// WithMetrics enables OpenTelemetry metrics on the underlying Redis driver
// (latency histograms, pool usage, etc.).
//
// Optional redisotel.MetricsOption values can be supplied to customise the
// instrumentation (custom meter provider, attribute filters, etc.).
func (r *Builder) WithMetrics(opts ...redisotel.MetricsOption) *Builder {
	r.metricsEnabled = true
	r.metricsOpts = opts
	return r
}

// ---------------------------------------------------------------------------
// Functional options (mirror of the fluent setters)
// ---------------------------------------------------------------------------

// WithAddresses returns a BuilderOptions that sets the Redis endpoints.
func WithAddresses(addresses ...string) BuilderOptions {
	return func(b *Builder) { b.addresses = addresses }
}

// WithKeyPrefix returns a BuilderOptions that sets the key namespace.
func WithKeyPrefix(prefix string) BuilderOptions {
	return func(b *Builder) { b.keyPrefix = prefix }
}

// WithTTL returns a BuilderOptions that sets the default TTL.
func WithTTL(ttl time.Duration) BuilderOptions {
	return func(b *Builder) { b.ttl = ttl }
}

// WithUsername returns a BuilderOptions that sets the ACL username.
func WithUsername(username string) BuilderOptions {
	return func(b *Builder) { b.username = username }
}

// WithPassword returns a BuilderOptions that sets the AUTH password.
func WithPassword(password string) BuilderOptions {
	return func(b *Builder) { b.password = password }
}

// WithDB returns a BuilderOptions that selects the Redis logical database.
func WithDB(db int) BuilderOptions {
	return func(b *Builder) { b.db = db }
}

// WithMasterName returns a BuilderOptions that enables Sentinel discovery.
func WithMasterName(name string) BuilderOptions {
	return func(b *Builder) { b.masterName = name }
}

// WithTLS returns a BuilderOptions that enables TLS.
func WithTLS(cfg *tls.Config) BuilderOptions {
	return func(b *Builder) { b.tlsConfig = cfg }
}

// WithPoolSize returns a BuilderOptions that sets the connection pool size.
func WithPoolSize(size int) BuilderOptions {
	return func(b *Builder) { b.poolSize = size }
}

// WithTimeouts returns a BuilderOptions that sets the network timeouts.
func WithTimeouts(dial, read, write time.Duration) BuilderOptions {
	return func(b *Builder) {
		b.dialTimeout = dial
		b.readTimeout = read
		b.writeTimeout = write
	}
}

// WithRouteRandomly returns a BuilderOptions that toggles random read routing.
func WithRouteRandomly(enabled bool) BuilderOptions {
	return func(b *Builder) { b.routeRandom = enabled }
}

// WithTracing returns a BuilderOptions that enables OpenTelemetry tracing.
// See Builder.WithTracing for details.
func WithTracing(opts ...redisotel.TracingOption) BuilderOptions {
	return func(b *Builder) {
		b.tracingEnabled = true
		b.tracingOpts = opts
	}
}

// WithMetrics returns a BuilderOptions that enables OpenTelemetry metrics.
// See Builder.WithMetrics for details.
func WithMetrics(opts ...redisotel.MetricsOption) BuilderOptions {
	return func(b *Builder) {
		b.metricsEnabled = true
		b.metricsOpts = opts
	}
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

// Build creates a production LowLevelClient backed by go-redis. The returned
// client transparently supports standalone, Sentinel and Cluster modes
// depending on the configured options.
//
// If no addresses are provided, "127.0.0.1:6379" is used as a sensible default
// for local development.
func (r *Builder) Build() *LowLevelClient {
	addrs := r.addresses
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:6379"}
	}

	universal := goredis.NewUniversalClient(&goredis.UniversalOptions{
		Addrs:                 addrs,
		MasterName:            r.masterName,
		Username:              r.username,
		Password:              r.password,
		DB:                    r.db,
		TLSConfig:             r.tlsConfig,
		PoolSize:              r.poolSize,
		DialTimeout:           r.dialTimeout,
		ReadTimeout:           r.readTimeout,
		WriteTimeout:          r.writeTimeout,
		RouteRandomly:         r.routeRandom,
		ContextTimeoutEnabled: true,
	})

	r.instrument(universal)

	return NewLowLevelClient(NewGoRedisClient(universal), r.keyPrefix, r.ttl)
}

// instrument attaches the requested OpenTelemetry hooks to the given client.
// Errors are ignored on purpose: redisotel only fails when the hook cannot be
// attached (a non-recoverable programming error). Skipping a missing hook is
// strictly better than panicking from inside a constructor.
func (r *Builder) instrument(client goredis.UniversalClient) {
	if r.tracingEnabled {
		_ = redisotel.InstrumentTracing(client, r.tracingOpts...)
	}
	if r.metricsEnabled {
		_ = redisotel.InstrumentMetrics(client, r.metricsOpts...)
	}
}

// BuildWithClient creates a LowLevelClient using the provided Client.
// Useful to inject custom adapters (for instrumentation, testing, etc.).
func (r *Builder) BuildWithClient(client Client) *LowLevelClient {
	return NewLowLevelClient(client, r.keyPrefix, r.ttl)
}

// FakeBuild creates a LowLevelClient backed by an in-memory FakeClient.
// Mirrors dynamodb.Builder.FakeBuild for symmetric ergonomics in tests.
func (r *Builder) FakeBuild() *LowLevelClient {
	return NewLowLevelClient(NewFakeClient(), r.keyPrefix, r.ttl)
}
