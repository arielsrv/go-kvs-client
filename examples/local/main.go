package main

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-relic/otel/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"

	"examples/local/infrastructure"
	"examples/local/model"

	"github.com/aws/aws-sdk-go-v2/config"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

func main() {
	ctx := context.Background()
	app, err := tracing.New(ctx,
		tracing.WithAppName("example"),
		tracing.WithProtocol(tracing.NewGRPCProtocol("localhost:4317")))
	if err != nil {
		log.Fatal(err)
	}

	defer func(app *tracing.App, ctx context.Context) {
		shutdownErr := app.Shutdown(ctx)
		if shutdownErr != nil {
			log.Fatal(err)
		}
	}(app, ctx)

	ctx, txn := tracing.NewTransaction(ctx, "MyService")
	defer txn.End()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	kvsClient := infrastructure.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(dynamodb.WithTTL(7*24*60*60), // 7 days (hh dd mm ss)
			dynamodb.WithContainerName("users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566")).
			Build(cfg))

	// get and save a single item
	for i := range 20 {
		userID := i + 1
		cacheKey := buildCacheKey(userID)
		if kvsErr := kvsClient.SaveWithContext(ctx, cacheKey, &model.UserDTO{ID: userID, FirstName: "John Doe"}, 10); kvsErr != nil {
			log.Error(kvsErr)
		}

		value, kvsErr := kvsClient.Get(cacheKey)
		if kvsErr != nil {
			log.Error(kvsErr)
		}

		log.Infof("Item %s: %+v", cacheKey, value)
	}

	// bulk get and save items
	err = kvsClient.BulkSaveWithContext(ctx, []model.UserDTO{
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

	items, err := kvsClient.BulkGetWithContext(ctx, []string{"101", "102", "103"})
	if err != nil {
		log.Error(err)
	}

	for i := range items {
		log.Infof("Item %d: %+v", i+1, items[i])
	}
}

func buildCacheKey(key int) string {
	return fmt.Sprintf("USER:%d:v1", key)
}
