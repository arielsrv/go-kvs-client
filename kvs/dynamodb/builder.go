package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Builder struct {
	containerName string
	rawURL        string
	ttl           int64
}

type BuilderOptions func(f *Builder)

func NewBuilder(opts ...BuilderOptions) *Builder {
	builder := &Builder{}

	for i := range opts {
		opt := opts[i]
		opt(builder)
	}

	return builder
}

func WithTTL(ttl int64) BuilderOptions {
	return func(f *Builder) {
		f.ttl = ttl
	}
}

func WithContainerName(containerName string) BuilderOptions {
	return func(f *Builder) {
		f.containerName = containerName
	}
}

func WithEndpointResolver(rawURL string) BuilderOptions {
	return func(f *Builder) {
		f.rawURL = rawURL
	}
}

func (r *Builder) Build(awsConfig aws.Config) *LowLevelClient {
	return NewLowLevelClient(
		dynamodb.NewFromConfig(awsConfig, func(opts *dynamodb.Options) {
			opts.EndpointResolverV2 = NewResolver(r.rawURL)
		}),
		r.containerName,
		r.ttl,
	)
}
