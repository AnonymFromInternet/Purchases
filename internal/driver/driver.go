package driver

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresDB struct {
	con *sql.DB
}

func OpenDB(dsn string) (*sql.DB, error) {
	var err error
	var connectionPool *sql.DB

	connectionPool, err = sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := connectionPool.Ping(); err != nil {
		fmt.Println("cannot ping the database", err)
		return nil, err
	}

	return connectionPool, nil
}
