package repository

import (
	"context"

	"github.com/AA122AA/metring/internal/server/domain"
)

type MetricsRepository interface {
	GetAll(ctx context.Context) (map[string]*domain.Metrics, error)
	Get(ctx context.Context, name string) (*domain.Metrics, error)
	Write(ctx context.Context, name string, value *domain.Metrics) error
	Update(ctx context.Context, value *domain.Metrics) error
}
