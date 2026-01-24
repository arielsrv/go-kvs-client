package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/arielsrv/go-kvs-client/examples/trace/model"
	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	checkErr(err)
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(
			dynamodb.WithTTL(time.Hour),
			dynamodb.WithContainerName("__kvs-users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566"),
		).Build(cfg),
	)

	// Single item: Save + Get
	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("USER:%d:v1", i)
		user := &model.UserDTO{ID: i, FirstName: "John Doe"}
		if kvsError := kvsClient.SaveWithContext(ctx, key, user, time.Second); kvsError != nil {
			fmt.Println(kvsError)
			continue
		}
		value, kvsErr := kvsClient.Get(key)
		if kvsErr != nil {
			fmt.Println(kvsErr)
			continue
		}
		fmt.Printf("Item %s: %+v\n", key, value)
	}

	// Bulk save + get
	users := []model.UserDTO{
		{ID: 101, FirstName: "Jane Doe"},
		{ID: 102, FirstName: "Alice Doe"},
		{ID: 103, FirstName: "Bob Doe"},
	}
	err = kvsClient.BulkSaveWithContext(ctx, users, func(userDTO model.UserDTO) string {
		return strconv.Itoa(userDTO.ID)
	})
	checkErr(err)

	keys := []string{"101", "102", "103"}
	items, err := kvsClient.BulkGetWithContext(ctx, keys)
	checkErr(err)

	for i, item := range items {
		fmt.Printf("Item %d: %+v\n", i+1, item)
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
