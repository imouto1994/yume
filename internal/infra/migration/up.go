package migration

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func UpLatest() {
	db, err := sql.Open("sqlite3", "sqlite3.db")
	if err != nil {
		zap.L().Fatal("failed to establish connection to database", zap.Error(err))
	}
	defer db.Close()

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		zap.L().Fatal("failed to create SQLite3 instance", zap.Error(err))
	}

	fileSource, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		zap.L().Fatal("failed to open source file", zap.Error(err))
	}

	m, err := migrate.NewWithInstance("file", fileSource, "sqlite3.db", driver)
	if err != nil {
		zap.L().Fatal("failed to create migrate instance", zap.Error(err))
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		zap.L().Fatal("failed to migrate up", zap.Error(err))
	}

	zap.L().Info("migrate up to latest successfully!")
}
