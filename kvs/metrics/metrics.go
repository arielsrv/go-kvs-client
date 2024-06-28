package metrics

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

var (
	single    sync.Once
	collector *Collector
)

func ProvideMetricCollector() *Collector {
	single.Do(func() {
		serviceMetricCollector := NewPrometheusServiceMetricCollector()
		collector = NewCollector(Default(), "go-kvs-client", serviceMetricCollector)
	})

	return collector
}

type ClientMetricCollector interface {
	IncrementCounter(clientName string, eventType string, eventSubType string, value ...float64)
	RecordExecutionTime(clientName, eventType string, eventSubType string, elapsedTime time.Duration)
	RecordValue(clientName string, eventType string, eventSubType string, value float64)
	Reset()
}

type Config struct {
	Environment string
	Application string
}

func Default() *Config {
	return &Config{
		Environment: strings.ToLower(os.Getenv("ENV")),
		Application: os.Getenv("APP_NAME"),
	}
}

type Collector struct {
	config      *Config
	collector   ServiceCollector
	serviceType string
}

func NewCollector(config *Config, serviceType string, collector ServiceCollector) *Collector {
	return &Collector{
		config:      config,
		serviceType: serviceType,
		collector:   collector,
	}
}

func (r Collector) IncrementCounter(clientName string, eventType string, eventSubType string, value ...float64) {
	r.collector.IncrementCounter(
		CounterDto{
			metricDto: metricDto{
				serviceType:  r.serviceType,
				environment:  r.config.Environment,
				application:  r.config.Application,
				clientName:   clientName,
				eventType:    eventType,
				eventSubType: eventSubType,
			},
			values: value,
		})
}

func (r Collector) RecordExecutionTime(clientName, eventType string, eventSubType string, elapsedTime time.Duration) {
	r.collector.RecordExecutionTime(
		TimerDto{
			metricDto: metricDto{
				serviceType:  r.serviceType,
				environment:  r.config.Environment,
				application:  r.config.Application,
				clientName:   clientName,
				eventType:    eventType,
				eventSubType: eventSubType,
			},
			elapsedTime: elapsedTime,
		})
}

func (r Collector) RecordValue(clientName string, eventType string, eventSubType string, value float64) {
	r.collector.RecordValue(
		ValueDto{
			metricDto: metricDto{
				serviceType:  r.serviceType,
				environment:  r.config.Environment,
				application:  r.config.Application,
				clientName:   clientName,
				eventType:    eventType,
				eventSubType: eventSubType,
			},
			value: value,
		})
}

func (r Collector) Reset() {
	r.collector.Reset()
}

type ServiceCollector interface {
	IncrementCounter(metric CounterDto)
	RecordExecutionTime(metric TimerDto)
	RecordValue(metric ValueDto)
	Reset()
}

type Mapper interface {
	BuildLabels() []string
}

type metricDto struct {
	serviceType  string
	environment  string
	application  string
	clientName   string
	eventType    string
	eventSubType string
}

func (m *metricDto) BuildLabels() []string {
	return []string{
		m.serviceType,
		m.environment,
		m.application,
		m.clientName,
		m.eventType,
		m.eventSubType,
	}
}

type CounterDto struct {
	values []float64

	metricDto
}

func (c *CounterDto) BuildLabels() []string {
	return c.metricDto.BuildLabels()
}

type TimerDto struct {
	elapsedTime time.Duration

	metricDto
}

func (t *TimerDto) BuildLabels() []string {
	return t.metricDto.BuildLabels()
}

func getNamespace() string {
	return "kvs"
}

func getLabels() []string {
	return []string{
		"service_type",
		"environment",
		"application",
		"client_name",
		"event_type",
		"event_subtype",
	}
}

type ValueDto struct {
	value float64

	metricDto
}

func (t *ValueDto) BuildLabels() []string {
	return t.metricDto.BuildLabels()
}

type PrometheusServiceMetricCollector struct {
	counter *prometheus.CounterVec
	gauge   *prometheus.GaugeVec
	summary *prometheus.SummaryVec
}

func (p *PrometheusServiceMetricCollector) IncrementCounter(counterDto CounterDto) {
	if len(counterDto.values) > 0 {
		p.counter.WithLabelValues(counterDto.BuildLabels()...).Add(counterDto.values[0])
	} else {
		p.counter.WithLabelValues(counterDto.BuildLabels()...).Inc()
	}
}

func (p *PrometheusServiceMetricCollector) RecordExecutionTime(timerDto TimerDto) {
	p.summary.WithLabelValues(timerDto.BuildLabels()...).Observe(float64(timerDto.elapsedTime.Milliseconds()))
}

func (p *PrometheusServiceMetricCollector) RecordValue(valueDto ValueDto) {
	p.gauge.WithLabelValues(valueDto.BuildLabels()...).Set(valueDto.value)
}

func (p *PrometheusServiceMetricCollector) Reset() {
	p.gauge.Reset()
	p.summary.Reset()
	p.counter.Reset()
}

func NewPrometheusServiceMetricCollector() *PrometheusServiceMetricCollector {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: getNamespace(),
			Name:      "counter",
			Help:      "Counter",
		}, getLabels(),
	)

	err := prometheus.Register(counter)
	if err != nil {
		log.Error(err)
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: getNamespace(),
			Name:      "gauge",
			Help:      "Gauge",
		}, getLabels(),
	)

	err = prometheus.Register(gauge)
	if err != nil {
		log.Error(err)
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: getNamespace(),
			Name:      "timer",
			Help:      "Timer",
			Objectives: map[float64]float64{
				0.5:  0.05,  // Average
				0.95: 0.01,  // P95
				0.99: 0.001, // P99
			},
		}, getLabels(),
	)

	err = prometheus.Register(summary)
	if err != nil {
		log.Error(err)
	}

	return &PrometheusServiceMetricCollector{
		counter: counter,
		gauge:   gauge,
		summary: summary,
	}
}
