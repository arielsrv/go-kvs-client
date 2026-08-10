module github.com/arielsrv/go-kvs-client/examples

go 1.26.5

// Always resolve the parent module from the local checkout so that
// `go mod tidy` (run via `task download`) picks up unreleased packages
// such as kvs/redis. Production users consume the published module via
// the normal `require` statement and never hit this replace.
replace github.com/arielsrv/go-kvs-client => ../

require (
	github.com/arielsrv/go-kvs-client v1.3.4
	github.com/aws/aws-sdk-go-v2/config v1.32.36
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.70.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.43.5 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.35 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.60 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.63.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.36.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.42.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.46.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.5 // indirect
	github.com/aws/smithy-go v1.27.7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/coocood/freecache v1.2.7 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/eko/gocache/lib/v4 v4.2.4 // indirect
	github.com/eko/gocache/store/freecache/v4 v4.2.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/magiconair/properties v1.18.11 // indirect
	github.com/mdelapenya/tlscert v0.2.0 // indirect
	github.com/moby/go-archive v0.3.3 // indirect
	github.com/moby/moby/api v1.55.0 // indirect
	github.com/moby/moby/client v0.5.1 // indirect
	github.com/moby/patternmatcher v0.6.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.22.0 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.22.0 // indirect
	github.com/redis/go-redis/v9 v9.22.0 // indirect
	github.com/shirou/gopsutil/v4 v4.26.7 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/yuin/gopher-lua v1.1.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
