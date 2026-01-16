package repository

import (
	"context"
	"fmt"

	"github.com/AA122AA/metring/internal/server/database"
	"github.com/AA122AA/metring/internal/server/database/query"
	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/go-faster/sdk/zctx"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type PSQLStorage struct {
	db *database.Database
	// TODO: create sqlc pack
	queries *query.Queries
	lg      *zap.Logger
}

func NewPSQLStorage(ctx context.Context, queries *query.Queries, db *database.Database) *PSQLStorage {
	return &PSQLStorage{
		db:      db,
		queries: queries,
		lg:      zctx.From(ctx).Named("Postges storage"),
	}
}

func (ps *PSQLStorage) GetAll(ctx context.Context) (map[string]*domain.Metrics, error) {
	metrics, err := ps.queries.GetAll(ctx)
	if err != nil {
		ps.lg.Error("cannot get all metrics", zap.Error(err))
		return nil, fmt.Errorf("cannot get all metrics: %w", err)
	}
	mm := make(map[string]*domain.Metrics)

	for _, m := range metrics {
		mm[m.Name] = domain.DBToDomain(&m)
	}

	return mm, nil
}

func (ps *PSQLStorage) Get(ctx context.Context, name string) (*domain.Metrics, error) {
	metric, err := ps.queries.Get(ctx, name)
	if err != nil {
		ps.lg.Error("cannot get metric", zap.String("metric name", name), zap.Error(err))
		return nil, fmt.Errorf("data not found")
	}

	return domain.DBToDomain(&metric), nil
}

func (ps *PSQLStorage) Update(ctx context.Context, value *domain.Metrics) error {
	err := ps.queries.Update(ctx, *parseUpdate(value))
	if err != nil {
		ps.lg.Error("cannot update metric", zap.String("metric name", value.ID), zap.Error(err))
		return fmt.Errorf("cannot update metric %v: %w", value.ID, err)
	}

	return nil
}

func (ps *PSQLStorage) UpdateMetrics(ctx context.Context, values []*domain.Metrics) error {
	tx, err := ps.db.BeginTx(ctx)
	if err != nil {
		ps.lg.Error("cannot begin transaction", zap.Error(err))
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := ps.queries.WithTx(tx)
	for _, metric := range values {
		err := q.Update(ctx, *parseUpdate(metric))
		if err != nil {
			ps.lg.Error("cannot update metric", zap.String("metric name", metric.ID), zap.Error(err))
			return fmt.Errorf("cannot update metric %v: %w", metric.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func (ps *PSQLStorage) Write(ctx context.Context, name string, value *domain.Metrics) error {
	err := ps.queries.Write(ctx, *parseWrite(value))
	if err != nil {
		ps.lg.Error("cannot write metric", zap.String("metric name", name), zap.Error(err))
		return fmt.Errorf("cannot write metric %v: %w", name, err)
	}

	return nil
}

func (ps *PSQLStorage) WriteMetrics(ctx context.Context, values []*domain.Metrics) error {
	tx, err := ps.db.BeginTx(ctx)
	if err != nil {
		ps.lg.Error("cannot begin transaction", zap.Error(err))
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := ps.queries.WithTx(tx)
	for _, metric := range values {
		err := q.Write(ctx, *parseWrite(metric))
		if err != nil {
			ps.lg.Error("cannot write metric", zap.String("metric name", metric.ID), zap.Error(err))
			return fmt.Errorf("cannot write metric %v: %w", metric.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func parseUpdate(value *domain.Metrics) *query.UpdateParams {
	arg := &query.UpdateParams{
		Name: value.ID,
	}
	switch value.MType {
	case domain.Counter:
		arg.Delta = pgtype.Int8{
			Int64: *value.Delta,
			Valid: true,
		}
	case domain.Gauge:
		arg.Value = pgtype.Float8{
			Float64: *value.Value,
			Valid:   true,
		}
	}

	return arg
}

func parseWrite(value *domain.Metrics) *query.WriteParams {
	arg := &query.WriteParams{
		Name: value.ID,
		Type: value.MType,
	}
	switch value.MType {
	case domain.Counter:
		arg.Delta = pgtype.Int8{
			Int64: *value.Delta,
			Valid: true,
		}
	case domain.Gauge:
		arg.Value = pgtype.Float8{
			Float64: *value.Value,
			Valid:   true,
		}
	}

	return arg
}
