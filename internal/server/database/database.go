package database

import (
	"context"
	"database/sql"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Database struct {
	db *sql.DB
	lg *zap.Logger
}

func New(ctx context.Context, driver, dsn string) *Database {
	logger := zctx.From(ctx).Named("Database")
	database, err := sql.Open(driver, dsn)
	if err != nil {
		logger.Fatal("Cannot open db", zap.Error(err))
	}

	return &Database{
		db: database,
		lg: logger,
	}
}

func (d *Database) Ping() error {
	return d.db.Ping()
}

func (d *Database) Close() error {
	return d.db.Close()
}
