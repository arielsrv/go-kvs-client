package kvs

import (
	"context"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"

	"github.com/pkg/errors"
)

type Client interface {
	Get(key string) (*Item, error)
	BulkGet(keys []string) (*Items, error)
	Save(key string, item *Item) error
	BulkSave(items *Items) error
	GetWithContext(ctx context.Context, key string) (*Item, error)
	SaveWithContext(ctx context.Context, key string, item *Item) error
	BulkGetWithContext(ctx context.Context, key []string) (*Items, error)
	BulkSaveWithContext(ctx context.Context, items *Items) error
	ContainerName() string
}

type LowLevelClientProxy struct {
	lowLevelClient Client
}

func NewLowLevelClientProxy(lowLevelClient Client) LowLevelClientProxy {
	return LowLevelClientProxy{
		lowLevelClient: lowLevelClient,
	}
}

func (r LowLevelClientProxy) Get(key string) (*Item, error) {
	return r.GetWithContext(context.Background(), key)
}

func (r LowLevelClientProxy) BulkGet(keys []string) (*Items, error) {
	return r.BulkGetWithContext(context.Background(), keys)
}

func (r LowLevelClientProxy) Save(key string, item *Item) error {
	return r.SaveWithContext(context.Background(), key, item)
}

func (r LowLevelClientProxy) BulkSave(items *Items) error {
	return r.BulkSaveWithContext(context.Background(), items)
}

func (r LowLevelClientProxy) GetWithContext(ctx context.Context, key string) (*Item, error) {
	metrics.Collector.Prometheus().IncrementCounter("__kvs_operations", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "get",
	})

	start := time.Now()
	value, err := r.lowLevelClient.GetWithContext(ctx, key)

	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "get",
	})

	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
				"client_name": r.lowLevelClient.ContainerName(),
				"stats":       "miss",
			})
		} else {
			metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
				"client_name": r.lowLevelClient.ContainerName(),
				"stats":       "error",
			})
		}
		return nil, err
	}

	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "hit",
	})

	return value, nil
}

func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "save",
	})

	start := time.Now()
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "save",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) BulkGetWithContext(ctx context.Context, key []string) (*Items, error) {
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "bulk_get",
	})

	start := time.Now()
	values, err := r.lowLevelClient.BulkGetWithContext(ctx, key)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "bulk_get",
	})
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (r LowLevelClientProxy) BulkSaveWithContext(ctx context.Context, items *Items) error {
	metrics.Collector.Prometheus().IncrementCounter("__kvs_stats", metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"stats":       "bulk_save",
	})

	start := time.Now()
	err := r.lowLevelClient.BulkSaveWithContext(ctx, items)
	metrics.Collector.Prometheus().RecordExecutionTime("__kvs_connection", time.Since(start), metrics.Tags{
		"client_name": r.lowLevelClient.ContainerName(),
		"type":        "bulk_save",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) ContainerName() string {
	return r.lowLevelClient.ContainerName()
}
