package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/randomtoy/gometrics/internal/model"
)

type DBStorage struct {
	DB *sql.DB
}

func NewDBConnector(dsn string) (*DBStorage, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	dbconn := &DBStorage{DB: db}
	return dbconn, nil
}

func (db DBStorage) InitDB() error {
	query := `
		CREATE TABLE IF NOT EXISTS metrics (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL CHECK (type IN ('gauge', 'counter')),
			value DOUBLE PRECISION NULL,
			delta BIGINT NULL
		);
	`
	_, err := db.DB.Exec(query)
	return err
}

func (db DBStorage) UpdateMetric(metric model.Metric) (model.Metric, error) {
	if metric.Type == model.Counter {
		fmt.Println("Updating Counter Type")
		m, err := db.GetMetric(metric.ID)
		fmt.Printf("value: %#v, error: %#v", m, err)
		if err == nil {
			metric.Summ(m.Delta)
		}
	}
	query := `
	INSERT INTO metrics (id, type, value, delta)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE 
	SET value = EXCLUDED.value, delta = EXCLUDED.delta;
	`
	_, err := db.DB.Exec(query, metric.ID, metric.Type, metric.Value, metric.Delta)
	if err != nil {
		return model.Metric{}, fmt.Errorf("cant write metric: %w", err)
	}

	res, err := db.GetMetric(metric.ID)
	if err != nil {
		return model.Metric{}, fmt.Errorf("cant get metric after writing: %w", err)
	}
	return res, nil
}

func (db DBStorage) GetMetric(m string) (model.Metric, error) {

	query := "SELECT id, type, value, delta FROM metrics WHERE id=$1"
	row := db.DB.QueryRow(query, m)

	var metric model.Metric
	err := row.Scan(&metric.ID, &metric.Type, &metric.Value, &metric.Delta)
	if err != nil {
		return model.Metric{}, err
	}

	return metric, nil

}

func (db *DBStorage) GetAllMetrics() (map[string]model.Metric, error) {
	rows, err := db.DB.Query("SELECT id, type, value, delta FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]model.Metric)

	for rows.Next() {
		var metric model.Metric
		err := rows.Scan(&metric.ID, &metric.Type, &metric.Value, &metric.Delta)
		if err != nil {
			return nil, err
		}
		metrics[metric.ID] = metric
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
