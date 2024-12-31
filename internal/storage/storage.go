package storage

import (
	"fmt"
	"strings"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Value interface{}
}

type Storage interface {
	UpdateMetric(metricType MetricType, metricName string, metricValue interface{}) error
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

func (s *InMemoryStorage) UpdateMetric(metricType MetricType, metricName string, metricValue interface{}) error {

	// it's hack for lowering direct write to store
	metricName = strings.ToLower(metricName)

	switch metricType {
	case Gauge:
		val, ok := metricValue.(float64)
		if !ok {
			return fmt.Errorf("invalid value for gauge metric %T", metricValue)
		}
		s.metrics[metricName] = Metric{Type: Gauge, Value: val}
	case Counter:
		val, ok := metricValue.(int64)
		if !ok {
			return fmt.Errorf("invalid value for counter metric %T", metricValue)
		}
		existing, found := s.metrics[metricName]
		if found {
			val += existing.Value.(int64)
		}
		s.metrics[metricName] = Metric{Type: Counter, Value: val}
	default:
		return fmt.Errorf("invalid metric type")
	}
	return nil
}

func (s *InMemoryStorage) GetMetric(metric string) (Metric, error) {
	metric = strings.ToLower(metric)
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
