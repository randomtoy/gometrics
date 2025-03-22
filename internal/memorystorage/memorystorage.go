package memorystorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type InMemoryStorage struct {
	Mutex   sync.Mutex
	Metrics map[string]model.Metric
	log     *zap.SugaredLogger
}

func NewInMemoryStorage(l *zap.SugaredLogger, path string) *InMemoryStorage {
	return &InMemoryStorage{
		Metrics: make(map[string]model.Metric),
		log:     l,
	}
}

func (s *InMemoryStorage) UpdateMetric(ctx context.Context, metric model.Metric) (model.Metric, error) {
	if metric.Type == model.Counter {
		existing, found := s.Metrics[metric.ID]
		if found {
			metric.Summ(existing.Delta)
		}
	}
	s.Metrics[metric.ID] = metric
	return s.Metrics[metric.ID], nil
}

func (s *InMemoryStorage) GetMetric(ctx context.Context, metric string) (model.Metric, error) {

	m, ok := s.Metrics[metric]
	if !ok {
		return model.Metric{}, fmt.Errorf("can't find metric: %s", metric)
	}
	return m, nil
}

func (s *InMemoryStorage) GetAllMetrics(ctx context.Context) (map[string]model.Metric, error) {
	result := make(map[string]model.Metric, len(s.Metrics))
	for k, v := range s.Metrics {
		result[k] = v
	}
	return result, nil
}

func (s *InMemoryStorage) Close() {}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStorage) UpdateMetricBatch(ctx context.Context, metrics []model.Metric) error {
	for _, metric := range metrics {
		if metric.Type == model.Counter {
			existing, found := s.Metrics[metric.ID]
			if found {
				metric.Summ(existing.Delta)
			}
		}
		s.Metrics[metric.ID] = metric
	}
	return nil
}
