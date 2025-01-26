package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	ID    string
	Type  MetricType
	Value *float64
	Delta *int64
}

type Storage interface {
	UpdateGauge(name string, value *float64) Metric
	UpdateCounter(name string, value *int64) Metric
	GetAllMetrics() map[string]Metric
	GetMetric(metric string) (Metric, error)
}

type InMemoryStorage struct {
	mutex   sync.Mutex
	metrics map[string]Metric
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		metrics: make(map[string]Metric),
	}
}

func (m Metric) String() string {
	switch m.Type {
	case Gauge:
		return fmt.Sprintf("%v", *m.Value)
	case Counter:
		return fmt.Sprintf("%v", *m.Delta)
	}
	return ""
}

func (s *InMemoryStorage) UpdateGauge(name string, value *float64) Metric {
	s.metrics[name] = Metric{ID: name, Type: Gauge, Value: value}
	return s.metrics[name]
}

func (s *InMemoryStorage) UpdateCounter(name string, value *int64) Metric {

	existing, found := s.metrics[name]
	if found {
		*value = *value + *existing.Delta
	}
	s.metrics[name] = Metric{ID: name, Type: Counter, Delta: value}
	return s.metrics[name]
}

func (s *InMemoryStorage) GetMetric(metric string) (Metric, error) {

	m, ok := s.metrics[metric]
	if !ok {
		return Metric{}, fmt.Errorf("can't find metric: %s", metric)
	}
	return m, nil
}

func (s *InMemoryStorage) SaveToFile(filepath string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(s.metrics)

}

func (s *InMemoryStorage) LoadFromFile(filepath string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error while opening file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&s.metrics)
}

func (s *InMemoryStorage) GetAllMetrics() map[string]Metric {
	result := make(map[string]Metric, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}
