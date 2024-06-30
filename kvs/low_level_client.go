package kvs

import (
	"context"
	"time"

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
	start := time.Now()
	item, err := r.lowLevelClient.Get(key)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "get", time.Since(start))
	if err != nil {
		r.collector.IncrementCounter(r.GetContainerName(), "stats", "miss")
		return nil, err
	}

	r.collector.IncrementCounter(r.GetContainerName(), "stats", "hit")

	return item, nil
}

func (r LowLevelClientProxy) BulkGet(keys []string) (*Items, error) {
	start := time.Now()
	values, err := r.lowLevelClient.BulkGet(keys)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_get", time.Since(start))

	if err != nil {
		return nil, err
	}

	return values, nil
}

func (r LowLevelClientProxy) Save(key string, item *Item) error {
	start := time.Now()
	err := r.lowLevelClient.Save(key, item)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "save", time.Since(start))
	if err != nil {
		r.collector.IncrementCounter(r.GetContainerName(), "stats", "save_error")
		return err
	}

	return nil
}

func (r LowLevelClientProxy) BulkSave(items *Items) error {
	start := time.Now()
	err := r.lowLevelClient.BulkSave(items)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_save", time.Since(start))
	if err != nil {
		r.collector.IncrementCounter(r.GetContainerName(), "stats", "bulk_save_error")
		return err
	}

	return nil
}

func (r LowLevelClientProxy) GetWithContext(ctx context.Context, key string) (*Item, error) {
	start := time.Now()
	value, err := r.lowLevelClient.GetWithContext(ctx, key)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "get_context", time.Since(start))
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	start := time.Now()
	err := r.lowLevelClient.SaveWithContext(ctx, key, item)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "save_context", time.Since(start))
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) BulkGetWithContext(ctx context.Context, key []string) (*Items, error) {
	start := time.Now()
	values, err := r.lowLevelClient.BulkGetWithContext(ctx, key)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_get_context", time.Since(start))
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (r LowLevelClientProxy) BulkSaveWithContext(ctx context.Context, items *Items) error {
	start := time.Now()
	err := r.lowLevelClient.BulkSaveWithContext(ctx, items)
	r.collector.RecordExecutionTime(r.GetContainerName(), "connection_time", "bulk_save_context", time.Since(start))
	if err != nil {
		return err
	}

	return nil
}

func (r LowLevelClientProxy) GetContainerName() string {
	return r.lowLevelClient.GetContainerName()
}
