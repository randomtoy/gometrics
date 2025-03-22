package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/randomtoy/gometrics/internal/memorystorage"
	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type FileStorage struct {
	memoryStorage *memorystorage.InMemoryStorage
	filepath      string
	log           *zap.SugaredLogger
}

func NewFileStorage(l *zap.SugaredLogger, memoryStorage *memorystorage.InMemoryStorage, filepath string) *FileStorage {
	return &FileStorage{
		memoryStorage: memoryStorage,
		filepath:      filepath,
		log:           l,
	}
}

func (fs *FileStorage) SaveToFile() error {
	fs.memoryStorage.Mutex.Lock()
	defer fs.memoryStorage.Mutex.Unlock()

	file, err := os.Create(fs.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(&fs.memoryStorage.Metrics)
}

func (fs *FileStorage) LoadFromFile() error {
	fs.memoryStorage.Mutex.Lock()
	defer fs.memoryStorage.Mutex.Unlock()

	file, err := os.Open(fs.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error while opening file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&fs.memoryStorage.Metrics)
}

func (fs *FileStorage) UpdateMetric(ctx context.Context, metric model.Metric) (model.Metric, error) {
	return fs.memoryStorage.UpdateMetric(ctx, metric)
}

func (fs *FileStorage) GetMetric(ctx context.Context, metric string) (model.Metric, error) {
	return fs.memoryStorage.GetMetric(ctx, metric)
}

func (fs *FileStorage) GetAllMetrics(ctx context.Context) (map[string]model.Metric, error) {
	return fs.memoryStorage.GetAllMetrics(ctx)
}

func (fs *FileStorage) UpdateMetricBatch(ctx context.Context, metrics []model.Metric) error {
	return fs.memoryStorage.UpdateMetricBatch(ctx, metrics)
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return fs.memoryStorage.Ping(ctx)
}

func (fs *FileStorage) Close() {
	err := fs.SaveToFile()
	if err != nil {
		fs.log.Infof("error saving metrics: %v", err)
	}
}
