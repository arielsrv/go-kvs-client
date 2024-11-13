package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

type Builder struct {
	containerName string
	rawURL        string
	ttl           int
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

func WithTTL(ttl int) BuilderOptions {
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

func (r *Builder) Build() *LowLevelClient {
	defaultConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return NewLowLevelClient(
		dynamodb.NewFromConfig(defaultConfig, func(opts *dynamodb.Options) {
			opts.EndpointResolverV2 = NewResolver(r.rawURL)
		}),
		r.containerName,
		r.ttl,
	)
}
