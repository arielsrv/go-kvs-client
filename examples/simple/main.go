package main

import (
	"context"
	"fmt"
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
	checkErr(err)

	kvsClient := kvs.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(
			dynamodb.WithTTL(time.Duration(24)*time.Hour),
			dynamodb.WithContainerName("users-store"),
			dynamodb.WithEndpointResolver("http://localhost:4566"),
		).Build(cfg),
	)

	// Single item: Save + Get
	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("USER:%d:v1", i)
		user := &model.UserDTO{ID: i, FirstName: "John Doe"}
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
		log.Infof("Item %d: %+v", i+1, item)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
