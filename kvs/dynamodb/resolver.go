package dynamodb

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type Resolver struct{}

func (*Resolver) ResolveEndpoint(_ context.Context, _ dynamodb.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	uri, err := url.Parse("http://localhost:4566/")
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}
	return smithyendpoints.Endpoint{
		URI: *uri,
	}, nil
}
