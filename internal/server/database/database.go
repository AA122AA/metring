package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AA122AA/metring/db/schema"
	"github.com/go-faster/sdk/zctx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type Database struct {
	pool *pgxpool.Pool

	lg *zap.Logger
}

func New(ctx context.Context, driver, dsn string) *Database {
	logger := zctx.From(ctx).Named("Database")
	database, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Fatal("Cannot open db", zap.Error(err))
	}

	return &Database{
		pool: database,
		lg:   logger,
	}
}

func (d *Database) DB() *pgxpool.Pool {
	return d.pool
}

func (d *Database) Migrate(ctx context.Context) error {
	goose.SetBaseFS(schema.Migrations)

	const cmd = "up"
	db, err := sql.Open("pgx", d.pool.Config().ConnString())
	if err != nil {
		d.lg.Error("Cannot connect to db in migrations", zap.Error(err))
		return fmt.Errorf("cannot run migrations: %w", err)
	}
	defer db.Close()

	err = goose.RunContext(ctx, cmd, db, ".")
	if err != nil {
		d.lg.Error("Cannot run migrations", zap.Error(err))
		return fmt.Errorf("cannot run migrations: %w", err)
	}
	return nil
}

func (d *Database) Ping(ctx context.Context) error {
	d.lg.Debug("Ping db")
	return d.pool.Ping(ctx)
}

func (d *Database) Close() {
	d.lg.Info("Closing db connection")
	d.pool.Close()
}
