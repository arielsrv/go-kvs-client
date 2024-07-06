package kvs

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/metrics"
)

type LowLevelClient interface {
	Get(key string) (*Item, error)
	BulkGet(keys []string) (*Items, error)
	Save(key string, item *Item) error
	BulkSave(items *Items) error
	GetWithContext(ctx context.Context, key string) (*Item, error)
	SaveWithContext(ctx context.Context, key string, item *Item) error
	BulkGetWithContext(ctx context.Context, key []string) (*Items, error)
	BulkSaveWithContext(ctx context.Context, items *Items) error
	GetContainerName() string
}

type LowLevelClientProxy struct {
	lowLevelClient LowLevelClient
	collector      *metrics.Collector
}

func NewLowLevelClientProxy(lowLevelClient LowLevelClient) LowLevelClientProxy {
	return LowLevelClientProxy{
		lowLevelClient: lowLevelClient,
		collector:      metrics.ProvideMetricCollector(),
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
	r.collector.IncrementCounter(r.GetContainerName(), "stats", "get")

	start := time.Now()
	value, err := r.lowLevelClient.GetWithContext(ctx, key)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "get", time.Since(start))
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			r.collector.IncrementCounter(r.GetContainerName(), "stats", "miss")
		} else {
			r.collector.IncrementCounter(r.GetContainerName(), "stats", "get_error")
		}
		return nil, err
	}

	r.collector.IncrementCounter(r.GetContainerName(), "stats", "hit")

	return value, nil
}

func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	r.collector.IncrementCounter(r.GetContainerName(), "stats", "save")

	start := time.Now()
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "save", time.Since(start))
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) BulkGetWithContext(ctx context.Context, key []string) (*Items, error) {
	r.collector.IncrementCounter(r.GetContainerName(), "stats", "get")

	start := time.Now()
	values, err := r.lowLevelClient.BulkGetWithContext(ctx, key)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_get", time.Since(start))
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (r LowLevelClientProxy) BulkSaveWithContext(ctx context.Context, items *Items) error {
	r.collector.IncrementCounter(r.GetContainerName(), "stats", "save")

	start := time.Now()
	err := r.lowLevelClient.BulkSaveWithContext(ctx, items)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_save", time.Since(start))
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) GetContainerName() string {
	return r.lowLevelClient.GetContainerName()
}
