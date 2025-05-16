// Package dynamodb provides AWS DynamoDB specific implementation of the KVS client.
package dynamodb

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

// Resolver is a custom endpoint resolver for DynamoDB.
// It allows specifying a custom endpoint URL for DynamoDB operations,
// which is useful for local development or testing with a custom endpoint.
type Resolver struct {
	rawURL string // The raw URL string for the DynamoDB endpoint
}

// NewResolver creates a new Resolver with the provided endpoint URL.
// Returns a pointer to the new Resolver.
func NewResolver(rawURL string) *Resolver {
	return &Resolver{
		rawURL: rawURL,
	}
}

// ResolveEndpoint implements the EndpointResolverV2 interface for the Resolver.
// It parses the raw URL and returns an endpoint with that URI.
// If the URL is invalid, it returns an error.
// The context and endpoint parameters are ignored.
func (r *Resolver) ResolveEndpoint(_ context.Context, _ dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	uri, err := url.Parse(r.rawURL)
	if err != nil {
		log.Warnf("[kvs]: invalid endpoint: %s", r.rawURL)
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *uri,
	}, nil
}
