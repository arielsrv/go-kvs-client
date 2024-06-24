package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

type AWSBuilder struct {
	ttl           int
	containerName string
}

type BuilderOptions func(f *AWSBuilder)

func NewBuilder(options ...BuilderOptions) *AWSBuilder {
	builder := &AWSBuilder{}

	for i := range options {
		opt := options[i]
		opt(builder)
	}

	return builder
}

func WithTTL(ttl int) BuilderOptions {
	return func(f *AWSBuilder) {
		f.ttl = ttl
	}
}

func WithContainerName(containerName string) BuilderOptions {
	return func(f *AWSBuilder) {
		f.containerName = containerName
	}
}

func (r AWSBuilder) Build() *Client {
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
