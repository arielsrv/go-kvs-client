package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

type Builder struct {
	ttl           int
	containerName string
}

type BuilderOptions func(f *Builder)

func NewBuilder(options ...BuilderOptions) *Builder {
	builder := &Builder{}

	for i := range options {
		opt := options[i]
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

func (r Builder) Build() *Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return NewClient(
		dynamodb.NewFromConfig(cfg, func(opts *dynamodb.Options) {
			opts.EndpointResolverV2 = new(Resolver)
		}),
		r.containerName,
		r.ttl,
	)
}
