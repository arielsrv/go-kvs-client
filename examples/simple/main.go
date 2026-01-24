package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/arielsrv/go-kvs-client/examples/trace/model"
	"github.com/arielsrv/go-kvs-client/kvs"
	"github.com/arielsrv/go-kvs-client/kvs/dynamodb"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
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
			continue
		}
		value, kvsErr := kvsClient.Get(key)
		if kvsErr != nil {
			continue
		}
		fmt.Printf("Item %s: %+v\n", key, value)
	}

	// Bulk save + get
	if err = kvsClient.BulkSaveWithContext(ctx, []model.UserDTO{
		{ID: 101, FirstName: "Jane", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Jane", "Doe")},
		{ID: 102, FirstName: "Bob", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Bob", "Doe")},
		{ID: 103, FirstName: "Alice", LastName: "Doe", FullName: fmt.Sprintf("%s %s", "Alice", "Doe")},
	}, func(userDTO model.UserDTO) string {
		return strconv.Itoa(userDTO.ID)
	}); err != nil {
		panic(err)
	}

	keys := []string{"101", "102", "103"}
	items, err := kvsClient.BulkGetWithContext(ctx, keys)
	if err != nil {
		panic(err)
	}

	for item := range slices.Values(items) {
		fmt.Printf("Item: %+v\n", item)
	}
}
