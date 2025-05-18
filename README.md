[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/releases)

> This package provides a high-level abstract for KVS (Key Value Store) distributed client (beta)

### Live example

```shell
task awslocal:start tf:init tf:apply
````

```shell
go run gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/examples/simple@latest
```

```shell
open https://app.localstack.cloud/inst/default/resources/dynamodb/tables/__kvs-users-store/items
```

example

```go
package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/examples/trace/model"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"

	"github.com/aws/aws-sdk-go-v2/config"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(
			dynamodb.WithTTL(time.Duration(24)*time.Hour),
			dynamodb.WithContainerName("__kvs-users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566"),
		).Build(cfg),
	)

	// Single item: Save and Get
	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("USER:%d:v1", i)
		user := &model.UserDTO{ID: i, FirstName: "John", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "John", "Doe")}
		if kvsError := kvsClient.SaveWithContext(ctx, key, user, time.Duration(10)*time.Second); kvsError != nil {
			log.Error(kvsError)
			continue
		}
		value, kvsErr := kvsClient.Get(key)
		if kvsErr != nil {
			log.Error(kvsErr)
			continue
		}
		log.Infof("Item %s: %+v", key, value)
	}

	// Bulk save + get
	if err = kvsClient.BulkSaveWithContext(ctx, []model.UserDTO{
		{ID: 101, FirstName: "Jane", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Jane", "Doe")},
		{ID: 102, FirstName: "Bob", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Bob", "Doe")},
		{ID: 103, FirstName: "Alice", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Alice", "Doe")},
	}, func(userDTO model.UserDTO) string {
		return strconv.Itoa(userDTO.ID)
	}); err != nil {
		log.Fatal(err)
	}

	keys := []string{"101", "102", "103"}
	items, err := kvsClient.BulkGetWithContext(ctx, keys)
	if err != nil {
		log.Fatal(err)
	}

	for item := range slices.Values(items) {
		log.Infof("Item: %+v", item)
	}
}

```

## Prometheus

example

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

## Tempo

![img.png](img.png)
