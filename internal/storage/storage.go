package storage

import (
	"fmt"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Value *float64
	Delta *int64
}

type Storage interface {
	// UpdateMetric(metricType MetricType, metricName string, metricValue interface{}) (Metric, error)
	UpdateGauge(name string, value *float64) Metric
	UpdateCounter(name string, value *int64) Metric
	GetAllMetrics() map[string]Metric
	GetMetric(metric string) (Metric, error)
}

type InMemoryStorage struct {
	metrics map[string]Metric
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		metrics: make(map[string]Metric),
	}
}

func (s *InMemoryStorage) UpdateGauge(name string, value *float64) Metric {
	s.metrics[name] = Metric{Type: Gauge, Value: value}
	return s.metrics[name]
}

func (s *InMemoryStorage) UpdateCounter(name string, value *int64) Metric {

	existing, found := s.metrics[name]
	if found {
		*value = *value + *existing.Delta
	}
	s.metrics[name] = Metric{Type: Counter, Delta: value}
	return s.metrics[name]
}

func (s *InMemoryStorage) GetMetric(metric string) (Metric, error) {
	// metric = strings.ToLower(metric)
	m, ok := s.metrics[metric]
	if !ok {
		return Metric{}, fmt.Errorf("can't find metric: %s", metric)
	}
	return m, nil
}

func (s *InMemoryStorage) GetAllMetrics() map[string]Metric {
	result := make(map[string]Metric, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}
