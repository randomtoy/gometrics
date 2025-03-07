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
	fmt.Println(dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return DBConnector{}, err
	}
	var dbconn DBConnector
	dbconn.DB = db
	return dbconn, nil
}

func (dbconn *DBConnector) Close() {
	if dbconn.DB != nil {
		fmt.Println("Closing DB cconnection...")
		dbconn.DB.Close()
	}
}

func (dbconn *DBConnector) Ping(ctx context.Context) error {
	return dbconn.DB.Ping()
}
