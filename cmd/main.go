package main

import (
	"strconv"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/model"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/cmd/infrastructure"

	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

func main() {
	kvsClient := infrastructure.NewKVSClient[model.UserDTO]()

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
}
