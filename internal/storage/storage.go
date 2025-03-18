package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/randomtoy/gometrics/internal/db"
	"github.com/randomtoy/gometrics/internal/filestorage"
	"github.com/randomtoy/gometrics/internal/memorystorage"
	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateMetric(ctx context.Context, metric model.Metric) (model.Metric, error)
	UpdateMetricBatch(ctx context.Context, metrics []model.Metric) error
	GetAllMetrics(ctx context.Context) (map[string]model.Metric, error)
	GetMetric(ctx context.Context, metric string) (model.Metric, error)

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

	memstorage := memorystorage.NewInMemoryStorage(l.Sugar(), config.Server.FilePath)

	if config.Server.FilePath != "" {
		store := filestorage.NewFileStorage(l.Sugar(), memstorage, config.Server.FilePath)
		if config.Server.Restore {
			err := store.LoadFromFile()
			if err != nil {
				l.Fatal("restoring error: %v", zap.Error(err))
			}
			l.Info("restore success")

		}
		ticker := time.NewTicker(time.Duration(config.Server.StoreInterval) * time.Second)
		go func() {
			for range ticker.C {
				err := store.SaveToFile()
				if err != nil {
					l.Sugar().Infof("error saving metrics: %v", err)
				}
			}
		}()
		l.Info("Using memorystorage")
		return store, nil
	}
	l.Info("Using memorystorage")
	return memstorage, nil
}
