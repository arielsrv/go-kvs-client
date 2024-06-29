package kvs

import (
	"context"

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
	item, err := r.lowLevelClient.Get(key)
	if err != nil {
		r.collector.IncrementCounter(r.GetContainerName(), "stats", "miss")
		return nil, err
	}

	r.collector.IncrementCounter(r.GetContainerName(), "stats", "hit")

	return item, nil
}

func (r LowLevelClientProxy) BulkGet(keys []string) (*Items, error) {
	return r.lowLevelClient.BulkGet(keys)
}

func (r LowLevelClientProxy) Save(key string, item *Item) error {
	return r.lowLevelClient.Save(key, item)
}

func (r LowLevelClientProxy) BulkSave(items *Items) error {
	return r.lowLevelClient.BulkSave(items)
}

func (r LowLevelClientProxy) GetWithContext(ctx context.Context, key string) (*Item, error) {
	return r.lowLevelClient.GetWithContext(ctx, key)
}

func (r LowLevelClientProxy) SaveWithContext(ctx context.Context, key string, item *Item) error {
	return r.lowLevelClient.SaveWithContext(ctx, key, item)
}

func (r LowLevelClientProxy) BulkGetWithContext(ctx context.Context, key []string) (*Items, error) {
	return r.lowLevelClient.BulkGetWithContext(ctx, key)
}

func (r LowLevelClientProxy) BulkSaveWithContext(ctx context.Context, items *Items) error {
	return r.lowLevelClient.BulkSaveWithContext(ctx, items)
}

func (r LowLevelClientProxy) GetContainerName() string {
	return r.lowLevelClient.GetContainerName()
}
