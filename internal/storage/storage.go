package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/randomtoy/gometrics/internal/db"
	"github.com/randomtoy/gometrics/internal/memorystorage"
	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateMetric(metric model.Metric) (model.Metric, error)
	UpdateMetricBatch(metrics []model.Metric) error
	GetAllMetrics() (map[string]model.Metric, error)
	GetMetric(metric string) (model.Metric, error)

	Close()
	Ping(ctx context.Context) error
}

func NewStorage(l *zap.Logger, config model.Config) (Storage, error) {
	if config.Server.DatabaseDSN != "" {
		dbconn, err := db.NewDBConnector(config.Server.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to create db connector: %w", err)
		}
		err = dbconn.InitDB()
		if err != nil {
			return nil, fmt.Errorf("failed to init db: %w", err)
		}
		l.Info("using PostgreSQL as default storage")
		return dbconn, nil
	}
	l.Info("Using in-memory storage")
	store := memorystorage.NewInMemoryStorage(l, config.Server.FilePath)

	if config.Server.Restore {
		err := store.LoadFromFile(config.Server.FilePath)
		if err != nil {
			l.Fatal("restoring error: %v", zap.Error(err))
		}
		l.Info("restore success")

	}
	if config.Server.FilePath != "" {
		ticker := time.NewTicker(time.Duration(config.Server.StoreInterval) * time.Second)
		go func() {
			for range ticker.C {
				err := store.SaveToFile(config.Server.FilePath)
				if err != nil {
					l.Sugar().Infof("error saving metrics: %v", err)
				}
			}
		}()
	}

	return store, nil
}
