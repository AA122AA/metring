package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/AA122AA/metring/internal/server/database"
	"github.com/AA122AA/metring/internal/server/database/query"
	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/go-faster/sdk/zctx"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type PSQLStorage struct {
	db      *database.Database
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

func (ps *PSQLStorage) getAllWithRetry(ctx context.Context) ([]query.GetAllRow, error) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	intervals := []int{1, 3, 5}
	maxRetry := 3

	for try := 0; try <= maxRetry; try++ {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-timer.C:
			metrics, err := ps.queries.GetAll(ctx)
			if err == nil {
				return metrics, nil
			}

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code != pgerrcode.ConnectionException {
					return nil, err
				}

				if try < len(intervals) {
					timer.Reset(time.Duration(intervals[try]) * time.Second)
					continue
				}
			}
			return nil, err
		}
	}

	return nil, fmt.Errorf("unknown error")
}

func (ps *PSQLStorage) GetAll(ctx context.Context) (map[string]*domain.Metrics, error) {
	metrics, err := ps.getAllWithRetry(ctx)
	// metrics, err := ps.queries.GetAll(ctx)
	if err != nil {
		ps.lg.Error("cannot get all metrics", zap.Error(err))
		// return nil, fmt.Errorf("cannot get all metrics: %w", err)
		return nil, NewEmptyRepoError(err)
	}
	mm := make(map[string]*domain.Metrics)

	for _, m := range metrics {
		mm[m.Name] = domain.DBToDomain(&query.GetRow{Name: m.Name, Type: m.Type, Delta: m.Delta, Value: m.Value, Hash: m.Hash})
	}

	return mm, nil
}

func (ps *PSQLStorage) Get(ctx context.Context, name string) (*domain.Metrics, error) {
	metric, err := ps.queries.Get(ctx, name)
	if err != nil {
		return nil, NewEmptyRepoError(err)
	}

	return domain.DBToDomain(&metric), nil
}

func (ps *PSQLStorage) Update(ctx context.Context, value *domain.Metrics) error {
	err := ps.queries.Update(ctx, *parseUpdate(value))
	if err != nil {
		return fmt.Errorf("cannot update metric %v: %w", value.ID, err)
	}

	return nil
}

func (ps *PSQLStorage) UpdateMetrics(ctx context.Context, values []*domain.Metrics) error {
	sort.Slice(values, func(i, j int) bool {
		return values[i].ID < values[j].ID
	})

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

// Подумать как это сделать через generic
func parseUpdate(value *domain.Metrics) *query.UpdateParams {
	arg := &query.UpdateParams{
		Name: value.ID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
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

// Подумать как это сделать через generic
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
