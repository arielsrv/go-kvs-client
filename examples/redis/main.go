package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/arielsrv/go-kvs-client/examples/trace/model"
	"github.com/arielsrv/go-kvs-client/kvs"
	kvsredis "github.com/arielsrv/go-kvs-client/kvs/redis"
)

func main() {
	ctx := context.Background()

	llClient := kvsredis.NewBuilder(
		kvsredis.WithAddresses("localhost:6379"),
		kvsredis.WithKeyPrefix("__kvs:users"),
		kvsredis.WithTTL(24*time.Hour),
		kvsredis.WithPoolSize(10),
		kvsredis.WithTimeouts(2*time.Second, time.Second, time.Second),
		// Enable OpenTelemetry tracing and metrics on the Redis driver.
		// (Configure a tracer/meter provider elsewhere in the app to actually
		// export the spans/metrics — see examples/trace for the OTLP wiring.)
		kvsredis.WithTracing(),
		kvsredis.WithMetrics(),
	).Build()
	defer func() { _ = llClient.Close() }()

	kvsClient := kvs.NewKVSClient[model.UserDTO](llClient)

	// Single item: Save (with explicit per-call TTL) and Get
	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("USER:%d:v1", i)
		user := &model.UserDTO{
			ID:        i,
			FirstName: "John",
			LastName:  "Doe",
			FullName:  "John Doe",
		}
		if err := kvsClient.SaveWithContext(ctx, key, user, 10*time.Second); err != nil {
			fmt.Printf("save %s: %v\n", key, err)
			continue
		}
		got, err := kvsClient.GetWithContext(ctx, key)
		if err != nil {
			fmt.Printf("get %s: %v\n", key, err)
			continue
		}
		fmt.Printf("Item %s: %+v\n", key, got)
	}

	// Bulk save + bulk get
	if err := kvsClient.BulkSaveWithContext(ctx, []model.UserDTO{
		{ID: 101, FirstName: "Jane", LastName: "Doe", FullName: "Jane Doe"},
		{ID: 102, FirstName: "Bob", LastName: "Doe", FullName: "Bob Doe"},
		{ID: 103, FirstName: "Alice", LastName: "Doe", FullName: "Alice Doe"},
	}, func(u model.UserDTO) string {
		return strconv.Itoa(u.ID)
	}); err != nil {
		panic(err)
	}

	items, err := kvsClient.BulkGetWithContext(ctx, []string{"101", "102", "103"})
	if err != nil {
		panic(err)
	}
	for it := range slices.Values(items) {
		fmt.Printf("Bulk item: %+v\n", it)
	}
}
