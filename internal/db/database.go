package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBConnector struct {
	DB *sql.DB
}

func NewDBConnector(dsn string) (DBConnector, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return DBConnector{}, fmt.Errorf("Connection error: %v", err)
	}
	defer db.Close()
	var dbconn DBConnector
	dbconn.DB = db
	return dbconn, nil
}

func (dbconn DBConnector) Ping(ctx context.Context) error {
	return dbconn.DB.PingContext(ctx)
}
