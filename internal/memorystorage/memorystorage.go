package memorystorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type InMemoryStorage struct {
	mutex    sync.Mutex
	metrics  map[string]model.Metric
	filepath string
	log      *zap.Logger
}

func NewInMemoryStorage(l *zap.Logger, path string) *InMemoryStorage {
	return &InMemoryStorage{
		metrics:  make(map[string]model.Metric),
		filepath: path,
		log:      l,
	}
}

func (s *InMemoryStorage) UpdateMetric(metric model.Metric) (model.Metric, error) {
	if metric.Type == model.Counter {
		existing, found := s.metrics[metric.ID]
		if found {
			metric.Summ(existing.Delta)
		}
	}
	s.metrics[metric.ID] = metric
	return s.metrics[metric.ID], nil
}

func (s *InMemoryStorage) GetMetric(metric string) (model.Metric, error) {

	m, ok := s.metrics[metric]
	if !ok {
		return model.Metric{}, fmt.Errorf("can't find metric: %s", metric)
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

func (s *InMemoryStorage) GetAllMetrics() (map[string]model.Metric, error) {
	result := make(map[string]model.Metric, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result, nil
}

func (s *InMemoryStorage) Close() {
	err := s.SaveToFile(s.filepath)
	if err != nil {
		s.log.Sugar().Infof("error saving metrics: %v", err)
	}
}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStorage) UpdateMetricBatch(metrics []model.Metric) error {
	for _, metric := range metrics {
		if metric.Type == model.Counter {
			existing, found := s.metrics[metric.ID]
			if found {
				metric.Summ(existing.Delta)
			}
		}
		s.metrics[metric.ID] = metric
	}
	return nil
}
