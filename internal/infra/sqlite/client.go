package sqlite

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Connect() (*sqlx.DB, error) {
	return sqlx.Connect("sqlite3", "./sqlite3.db")
}