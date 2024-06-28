package main

import (
	"strconv"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/infrastructure"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/model"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

func main() {
	kvsClient := infrastructure.NewAWSKVSClient[model.UserDTO](
		dynamodb.NewBuilder(dynamodb.WithTTL(7*24*60*60), // 7 days (hh dd mm ss)
			dynamodb.WithContainerName("users"),
			dynamodb.WithEndpointResolver("http://localhost:4566")).
			Build())

	// get and save a single item
	for i := range 20 {
		key := i + 1
		err := kvsClient.Save(strconv.Itoa(key), &model.UserDTO{
			ID:   key,
			Name: "John Doe",
		})
		if err != nil {
			log.Error(err)
		}

		value, err := kvsClient.Get(strconv.Itoa(key))
		if err != nil {
			log.Error(err)
		}

		log.Info(value)
	}

	// bulk get and save items
	if err := kvsClient.BulkSave([]model.UserDTO{
		{ID: 101, Name: "Jane Doe"},
		{ID: 102, Name: "Alice Doe"},
		{ID: 103, Name: "Bob Doe"},
	},
		func(userDTO model.UserDTO) string {
			return strconv.Itoa(userDTO.ID)
		}); err != nil {
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
