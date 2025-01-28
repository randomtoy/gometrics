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
	ID    string     `json:"id"`
	Type  MetricType `json:"type"`
	Value *float64   `json:"value,omitempty"`
	Delta *int64     `json:"delta,omitempty"`
}

type Storage interface {
	UpdateMetric(metric Metric) Metric
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

func (s *InMemoryStorage) UpdateMetric(metric Metric) Metric {
	if metric.Type == Counter {
		existing, found := s.metrics[metric.ID]
		if found {
			*metric.Delta += *existing.Delta
		}
	}
	s.metrics[metric.ID] = metric
	return s.metrics[metric.ID]
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
