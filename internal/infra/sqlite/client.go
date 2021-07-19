package sqlite

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DBOps interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Rebind(query string) string
}

type Tx interface {
	DBOps
	Commit() error
	Rollback() error
}

type DB interface {
	DBOps
	Close() error
	BeginTxx(context.Context, *sql.TxOptions) (*sqlx.Tx, error)
}

func Connect() (DB, error) {
	return sqlx.Connect("sqlite3", "sqlite3.db")
}
