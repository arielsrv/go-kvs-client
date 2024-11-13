[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/-/releases)

> This package provides a high-level abstract for KVS (Key Value Store) distributed client (beta)

example
```go
package main

import (
	"strconv"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/infrastructure"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/model"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

func main() {
	kvsClient := infrastructure.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(dynamodb.WithTTL(7*24*60*60), // 7 days (hh dd mm ss)
			dynamodb.WithContainerName("users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566")).
			Build())

	// get and save a single item
	for i := range 20 {
		key := i + 1
		if err := kvsClient.Save(strconv.Itoa(key), &model.UserDTO{
			ID:        key,
			FirstName: "John Doe",
		}); err != nil {
			log.Error(err)
		}

		value, err := kvsClient.Get(strconv.Itoa(key))
		if err != nil {
			log.Error(err)
		}

		log.Infof("Item %d: %+v", key, value)
	}

	// bulk get and save items
	err := kvsClient.BulkSave([]model.UserDTO{
		{
			ID:        101,
			FirstName: "Jane Doe",
		}, {
			ID:        102,
			FirstName: "Alice Doe",
		}, {
			ID:        103,
			FirstName: "Bob Doe",
		},
	}, func(userDTO model.UserDTO) string {
		return strconv.Itoa(userDTO.ID)
	})
	if err != nil {
		log.Error(err)
	}

	items, err := kvsClient.BulkGet([]string{"101", "102", "103"})
	if err != nil {
		log.Error(err)
	}

	for i := range items {
		log.Infof("Item %d: %+v", i+1, items[i])
	}
}

```
