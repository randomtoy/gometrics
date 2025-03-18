package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	sqlc "github.com/randomtoy/gometrics/internal/db/sqlc"
	"github.com/randomtoy/gometrics/internal/model"
)

type DBStorage struct {
	Queries *sqlc.Queries
	DB      *sql.DB
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.ConnectionException ||
			pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
			pgErr.Code == pgerrcode.ConnectionFailure
	}
	return false
}

func NewDBConnector(dsn string) (*DBStorage, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	dbconn := &DBStorage{
		Queries: sqlc.New(db),
		DB:      db,
	}
	return dbconn, nil
}

func (db DBStorage) InitDB() error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("cat initialise goose dialect: %w", err)
	}
	err = goose.Up(db.DB, "internal/migrations")
	if err != nil {
		return fmt.Errorf("can't apply migrations")
	}
	return nil
}

func (db DBStorage) UpdateMetric(ctx context.Context, metric model.Metric) (model.Metric, error) {
	if metric.Type == model.Counter {
		m, err := db.GetMetric(ctx, metric.ID)
		if err == nil {
			metric.Summ(m.Delta)
		}
	}
	err := db.Queries.InsertOrUpdateMetric(ctx, sqlc.InsertOrUpdateMetricParams{
		ID:    metric.ID,
		Type:  string(metric.Type),
		Value: sql.NullFloat64{Float64: metric.DerefFloat64(metric.Value), Valid: metric.Value != nil},
		Delta: sql.NullInt64{Int64: metric.DerefInt64(metric.Delta), Valid: metric.Delta != nil},
	})

	if err != nil {
		return model.Metric{}, fmt.Errorf("cant write metric: %w", err)
	}

	res, err := db.GetMetric(ctx, metric.ID)
	if err != nil {
		return model.Metric{}, fmt.Errorf("cant get metric after writing: %w", err)
	}
	return res, nil
}

func (db DBStorage) GetMetric(ctx context.Context, id string) (model.Metric, error) {
	m, err := db.Queries.GetMetric(ctx, id)
	if err != nil {
		return model.Metric{}, fmt.Errorf("cant get metric: %w", err)
	}
	metric := model.Metric{
		ID:   m.ID,
		Type: model.MetricType(m.Type),
	}
	metric.Value = toFloat64Ptr(m.Value)
	metric.Delta = toInt64Ptr(m.Delta)

	return metric, nil
}

func (db *DBStorage) GetAllMetrics(ctx context.Context) (map[string]model.Metric, error) {
	metricsList, err := db.Queries.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("cant get all metrics: %w", err)
	}

	metrics := make(map[string]model.Metric)

	for _, m := range metricsList {
		metric := model.Metric{
			ID:   m.ID,
			Type: model.MetricType(m.Type),
		}
		metric.Value = toFloat64Ptr(m.Value)
		metric.Delta = toInt64Ptr(m.Delta)
		metrics[m.ID] = metric
	}

	return metrics, nil
}

func (db DBStorage) Close() {
	if db.DB != nil {
		fmt.Println("Closing DB cconnection...")
		db.DB.Close()
	}
}

func (db DBStorage) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db DBStorage) UpdateMetricBatch(ctx context.Context, metrics []model.Metric) error {

	groupedMetrics := make(map[string]model.Metric)
	for _, metric := range metrics {
		existing, found := groupedMetrics[metric.ID]
		if found {
			if metric.Type == model.Counter {
				metric.Summ(existing.Delta)
			}
		}
		groupedMetrics[metric.ID] = metric
	}
	var lastErr error
	for attempt := 1; attempt <= 4; attempt++ {
		err := db.insertBatch(ctx, groupedMetrics)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetriableError(err) {
			return fmt.Errorf("fatal DB error: %w", err)
		}

		backoff := time.Duration((attempt-1)*2+1) * time.Second

		fmt.Printf("Retrying DB insert in %v due to error: %v", backoff, err)
		time.Sleep(backoff)
	}
	return fmt.Errorf("failed to insert metrics after retries: %w", lastErr)
}

func (db *DBStorage) insertBatch(ctx context.Context, gMetrics map[string]model.Metric) error {
	tx, err := db.DB.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := db.Queries.WithTx(tx)

	for _, metric := range gMetrics {
		if metric.Type == model.Counter {
			m, err := db.GetMetric(ctx, metric.ID)
			if err == nil {
				metric.Summ(m.Delta)
			}
		}
		err := query.InsertOrUpdateMetric(ctx, sqlc.InsertOrUpdateMetricParams{
			ID:    metric.ID,
			Type:  string(metric.Type),
			Value: sql.NullFloat64{Float64: metric.DerefFloat64(metric.Value), Valid: metric.Value != nil},
			Delta: sql.NullInt64{Int64: metric.DerefInt64(metric.Delta), Valid: metric.Delta != nil},
		})
		if err != nil {
			return fmt.Errorf("can't write metric to DB: %w", err)
		}
	}
	return tx.Commit()
}
