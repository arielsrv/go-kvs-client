// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Builder is a struct that helps configure and create a LowLevelClient for DynamoDB.
// It uses the builder pattern with functional options to allow for flexible configuration.
type Builder struct {
	containerName string        // Name of the container or service, used for metrics and logging
	rawURL        string        // URL for the DynamoDB endpoint, useful for local development
	ttl           time.Duration // Default Time To Live for items in seconds
}

// BuilderOptions is a function type that configures a Builder.
// It's used with the functional options pattern to provide a flexible API.
type BuilderOptions func(f *Builder)

// NewBuilder creates a new Builder with the provided options.
// Each option is a function that configures some aspect of the Builder.
// Returns a pointer to the configured Builder.
func NewBuilder(opts ...BuilderOptions) *Builder {
	builder := &Builder{}

	for i := range opts {
		opt := opts[i]
		opt(builder)
	}

	return builder
}

// WithEndpointResolver sets the DynamoDB endpoint URL.
// This is useful for local development or testing with a custom endpoint.
// Returns a pointer to the Builder.
func (r *Builder) WithEndpointResolver(rawURL string) *Builder {
	r.rawURL = rawURL
	return r
}

// WithContainerName sets the container name.
// The container name is used for metrics and logging.
// Returns a pointer to the Builder.
func (r *Builder) WithContainerName(containerName string) *Builder {
	r.containerName = containerName
	return r
}

// WithTTL sets the default TTL for items.
// The TTL is specified in seconds.
// Returns a pointer to the Builder.
func (r *Builder) WithTTL(ttl time.Duration) *Builder {
	r.ttl = ttl
	return r
}

// WithTTL returns a BuilderOptions that sets the default TTL for items.
// The TTL is specified in seconds.
func WithTTL(ttl time.Duration) BuilderOptions {
	return func(f *Builder) {
		f.ttl = ttl
	}
}

// WithContainerName returns a BuilderOptions that sets the container name.
// The container name is used for metrics and logging.
func WithContainerName(containerName string) BuilderOptions {
	return func(f *Builder) {
		f.containerName = containerName
	}
}

// WithEndpointResolver returns a BuilderOptions that sets the DynamoDB endpoint URL.
// This is useful for local development or testing with a custom endpoint.
func WithEndpointResolver(rawURL string) BuilderOptions {
	return func(f *Builder) {
		f.rawURL = rawURL
	}
}

// Build creates a new LowLevelClient using the configured options and the provided AWS config.
// It sets up the DynamoDB client with the specified endpoint resolver and other options.
// Returns a pointer to the new LowLevelClient.
func (r *Builder) Build(awsConfig aws.Config) *LowLevelClient {
	return NewLowLevelClient(
		dynamodb.NewFromConfig(awsConfig, func(opts *dynamodb.Options) {
			opts.EndpointResolverV2 = NewResolver(r.rawURL)
		}),
		r.containerName,
		r.ttl,
	)
}
