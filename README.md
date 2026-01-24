[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/releases)

Go client for a distributed Key-Value Store (KVS) on AWS DynamoDB, with optional in‑memory cache, Prometheus metrics, and OpenTelemetry tracing. Status: beta.

Features
- Generic, typed API (Go generics) for Get, Save, BulkGet, and BulkSave; context-enabled variants are available.
- AWS DynamoDB implementation with an options builder (TTL, table/container, endpoint resolver/LocalStack).
- Optional in‑memory cache via freecache/gocache to improve latency (hit/miss exposed as metrics).
- Built-in Prometheus metrics (operations, connection latencies, hit/miss, errors).
- OpenTelemetry tracing (example with Tempo and otelaws for AWS SDK v2).
- Ready-to-run usage examples (simple and trace).

Requirements
- Go >= 1.25
- Docker and Docker Compose (optional) for the local observability stack.
- AWS credentials or LocalStack for local development.

Installation
```shell
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client@latest
```

Quick start (LocalStack)
1) Spin up local env (LocalStack + Prometheus + Grafana + Tempo):
```shell
task awslocal:start tf:init tf:apply
```
2) Run simple example:
```shell
go run gitlab.com/arielsrv/go-kvs-client/examples/simple@latest
```
3) Inspect DynamoDB data (LocalStack UI):
```shell
open https://app.localstack.cloud/inst/default/resources/dynamodb/tables/__kvs-users-store/items
```

Code example
```go
package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"gitlab.com/arielsrv/go-kvs-client/examples/trace/model"
	"gitlab.com/arielsrv/go-kvs-client/kvs"
	"github.com/aws/aws-sdk-go-v2/config"
	"gitlab.com/arielsrv/go-kvs-client/kvs/dynamodb"
	"gitlab.com/arielsrv/go-logger/log"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil { log.Fatal(err) }

	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(
			dynamodb.WithTTL(24*time.Hour),
			dynamodb.WithContainerName("__kvs-users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566"), // LocalStack
		).Build(cfg),
	)

 // Save and read an item
	for i := 1; i <= 2; i++ {
		key := fmt.Sprintf("USER:%d:v1", i)
		user := &model.UserDTO{ID: i, FirstName: "John", LastName: "Doe", FullName: "John Doe"}
		if err := kvsClient.SaveWithContext(ctx, key, user, 10*time.Second); err != nil { log.Error(err); continue }
		value, err := kvsClient.Get(key)
		if err != nil { log.Error(err); continue }
		log.Infof("Item %s: %+v", key, value)
	}

	// Bulk save + bulk get
	if err = kvsClient.BulkSaveWithContext(ctx, []model.UserDTO{
		{ID: 101, FirstName: "Jane", LastName: "Doe", FullName: "Jane Doe"},
		{ID: 102, FirstName: "Bob", LastName: "Doe", FullName: "Bob Doe"},
		{ID: 103, FirstName: "Alice", LastName: "Doe", FullName: "Alice Doe"},
	}, func(u model.UserDTO) string { return strconv.Itoa(u.ID) }); err != nil { log.Fatal(err) }

	keys := []string{"101", "102", "103"}
	items, err := kvsClient.BulkGetWithContext(ctx, keys)
	if err != nil { log.Fatal(err) }
	for item := range slices.Values(items) { log.Infof("Item: %+v", item) }
}
```

KVS API (summary)
- Get(key string) (*T, error)
- BulkGet(keys []string) ([]T, error)
- Save(key string, item *T, ttl ...time.Duration) error
- BulkSave(items []T, keyMapper KeyMapperFunc[T], ttl ...time.Duration) error
- Context-aware versions: GetWithContext, BulkGetWithContext, SaveWithContext, BulkSaveWithContext

DynamoDB/LocalStack configuration
- Default table in examples: __kvs-users-store
- Change endpoint with WithEndpointResolver for local or real AWS.
- TTL: use WithTTL in the builder for automatic expiration.

Prometheus metrics (indicative names)
```text
__kvs_operations{client_name="client_name", type="get"} 10
__kvs_operations{client_name="client_name", type="save"} 15
__kvs_operations{client_name="client_name", type="bulk_get"} 8
__kvs_operations{client_name="client_name", type="bulk_save"} 5
__kvs_stats{client_name="client_name", stats="hit"} 1
__kvs_stats{client_name="client_name", stats="miss"} 1
__kvs_stats{client_name="client_name", stats="error"} 0
__kvs_connection{client_name="client_name", type="get"} 0.005
__kvs_connection{client_name="client_name", type="save"} 0.002
__kvs_connection{client_name="client_name", type="bulk_get"} 0.010
__kvs_connection{client_name="client_name", type="bulk_save"} 0.012
```

Tracing (Tempo / OpenTelemetry)
- Full example in examples/trace.
- Integration with AWS SDK v2 via otelaws.AppendMiddlewares.
- Screenshots: see img.png.

How to develop and test
- Run tests: `task test` or `go test ./...`
- Lint/coverage (per Taskfile.yml): `task lint`, `task cover`
- Run examples: `go run ./examples/simple` and `go run ./examples/trace`

Local observability stack (Docker Compose)
- Files at resources/setup/docker: docker-compose.yml, prometheus.yaml, grafana-datasources.yaml, tempo.yaml
- Start: `task awslocal:start`
- Sample dashboards under resources/grafana/ (import in Grafana)

Roadmap / Planned
- Support for EC and additional providers is planned (e.g., AWS ElastiCache/MemoryDB, and other cloud KVS/backends from GCP/Azure). Contributions and proposals are welcome.

Compatibility & versioning
- Go 1.25+
- Releases: see the releases badge in GitLab.

License
- IskayPet proprietary. Internal use only.
