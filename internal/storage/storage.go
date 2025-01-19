package storage

import (
	"fmt"
	"strconv"
	"strings"
)

type Metric struct {
	ID    string
	Type  MetricType
	Value any
}

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
	Unknown MetricType = "unknown"
)

type Storage interface {
	UpdateMetric(metricType string, metricName string, metricValue string) (Metric, error)
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

func parseMetricType(s string) (MetricType, error) {
	metricTypes := map[MetricType]struct{}{
		Gauge:   {},
		Counter: {},
		Unknown: {},
	}
	mType := MetricType(s)
	_, ok := metricTypes[mType]
	if !ok {
		return Unknown, fmt.Errorf("invalid status: %q", s)
	}
	return mType, nil
}

func (s *InMemoryStorage) UpdateMetric(metricType string, metricName string, metricValue string) (Metric, error) {

	// it's hack for lowering direct write to store
	metricName = strings.ToLower(metricName)

	mtype, err := parseMetricType(metricType)
	if err != nil {
		return Metric{}, err
	}
	switch mtype {
	case Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return Metric{}, fmt.Errorf("error convertng to float64: %s", err)
		}

		s.metrics[metricName] = Metric{ID: metricName, Type: Gauge, Value: value}
	case Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return Metric{}, fmt.Errorf("error convertng to int64: %s", err)
		}
		existing, found := s.metrics[metricName]
		if found {
			value += existing.Value.(int64)
		}
		s.metrics[metricName] = Metric{ID: metricName, Type: Counter, Value: value}
	default:
		return Metric{}, fmt.Errorf("invalid metric type")
	}
	return s.metrics[metricName], nil
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
