package dynamodb

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

type Resolver struct {
	rawURL string
}

func NewResolver(rawURL string) *Resolver {
	return &Resolver{
		rawURL: rawURL,
	}
}

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
