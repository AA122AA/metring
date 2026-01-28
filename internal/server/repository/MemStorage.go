package repository

import (
	"context"

	"github.com/AA122AA/metring/internal/server/domain"
)

type MemStorage struct {
	Values map[string]*domain.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Values: make(map[string]*domain.Metrics),
	}
}

func (ms *MemStorage) GetAll(ctx context.Context) (map[string]*domain.Metrics, error) {
	if len(ms.Values) != 0 {
		return ms.Values, nil
	}
	// return nil, fmt.Errorf("no metrics")
	return nil, NewEmptyRepoError(nil)
}

func (ms *MemStorage) Get(ctx context.Context, name string) (*domain.Metrics, error) {
	if v, ok := ms.Values[name]; ok {
		return v, nil
	}

	return nil, NewEmptyRepoError(nil)
}

func (ms *MemStorage) Write(ctx context.Context, name string, value *domain.Metrics) error {
	ms.Values[name] = value
	return nil
}

func (ms *MemStorage) WriteMetrics(ctx context.Context, values []*domain.Metrics) error {
	for _, v := range values {
		ms.Values[v.ID] = v
	}

	return nil
}

func (ms *MemStorage) Update(ctx context.Context, value *domain.Metrics) error {
	ms.Values[value.ID] = value
	return nil
}

func (ms *MemStorage) UpdateMetrics(ctx context.Context, values []*domain.Metrics) error {
	for _, v := range values {
		ms.Values[v.ID] = v
	}

	return nil
}
