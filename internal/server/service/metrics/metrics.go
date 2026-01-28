package metrics

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/AA122AA/metring/internal/server/constants"
	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metrics struct {
	repo repository.MetricsRepository
	lg   *zap.Logger
}

func NewMetrics(ctx context.Context, r repository.MetricsRepository) *Metrics {
	return &Metrics{
		repo: r,
		lg:   zctx.From(ctx).Named("metrics service"),
	}
}

func (m *Metrics) GetAll(ctx context.Context) (map[string]*domain.Metrics, error) {
	return m.repo.GetAll(ctx)
}

func (m *Metrics) get(ctx context.Context, data *domain.MetricsJSON) (*domain.Metrics, error) {
	if err := validate(data, constants.Get); err != nil {
		return nil, err
	}
	metric, err := m.repo.Get(ctx, data.ID)
	if err != nil {
		return nil, fmt.Errorf("err from repo: %w", err)
	}
	if data.MType != metric.MType {
		return nil, fmt.Errorf("metric %v, and data %v, has different types between data and repo: %v vs %v", metric.ID, data.ID, data.MType, metric.MType)
	}
	metric.ID = data.ID

	return metric, nil
}

func (m *Metrics) Get(ctx context.Context, data *domain.MetricsJSON) (string, error) {
	metric, err := m.get(ctx, data)
	if err != nil {
		return "", err
	}

	var res string
	switch metric.MType {
	case domain.Gauge:
		res = fmt.Sprintf("%g", *metric.Value)
	case domain.Counter:
		res = fmt.Sprintf("%d", *metric.Delta)
	}

	return res, nil
}

func (m *Metrics) GetJSON(ctx context.Context, data *domain.MetricsJSON) (*domain.MetricsJSON, error) {
	metric, err := m.get(ctx, data)
	if err != nil {
		return nil, err
	}

	return domain.TransformToJSON(metric), nil
}

func (m *Metrics) Update(ctx context.Context, data *domain.MetricsJSON) error {
	if err := validate(data, constants.Update); err != nil {
		return err
	}
	metric := domain.TransformFromJSON(data)

	v, err := m.repo.Get(ctx, metric.ID)
	if err != nil {
		var er *repository.EmptyRepoError
		if errors.Is(err, er) {
			return m.repo.Write(ctx, data.ID, metric)
		}
		return fmt.Errorf("%w", err)
	}

	// Увеличиваем значение, если это Counter
	if data.MType == domain.Counter {
		*metric.Delta += *v.Delta
	}

	return m.repo.Update(ctx, metric)
}

func (m *Metrics) Updates(ctx context.Context, data []*domain.MetricsJSON) error {
	mm, err := trim(data)
	if err != nil {
		return err
	}

	toUpdate := make([]*domain.Metrics, 0, len(mm))
	toInsert := make([]*domain.Metrics, 0, len(mm))

	for name, metric := range mm {
		fromRepo, err := m.repo.Get(ctx, name)
		if err != nil {
			var er *repository.EmptyRepoError
			if errors.Is(err, er) {
				toInsert = append(toInsert, metric)
				continue
			}
			return err
		}

		if fromRepo.MType == domain.Counter {
			*metric.Delta += *fromRepo.Delta
		}

		toUpdate = append(toUpdate, metric)
	}

	if len(toInsert) > 0 {
		err := m.repo.WriteMetrics(ctx, toInsert)
		if err != nil {
			return err
		}
	}

	if len(toUpdate) > 0 {
		err := m.repo.UpdateMetrics(ctx, toUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func trim(data []*domain.MetricsJSON) (map[string]*domain.Metrics, error) {
	mm := make(map[string]*domain.Metrics, len(data))
	for _, d := range data {
		err := validate(d, constants.Update)
		if err != nil {
			return nil, err
		}

		metric := domain.TransformFromJSON(d)

		if v, ok := mm[metric.ID]; ok {
			if v.MType == domain.Counter {
				*v.Delta += *metric.Delta
			} else {
				*v.Value = *metric.Value
			}
		} else {
			mm[metric.ID] = metric
		}
	}

	return mm, nil
}

func (m *Metrics) Parse(mType, mName, value, handler string) (*domain.MetricsJSON, error) {
	data := &domain.MetricsJSON{
		ID:    mName,
		MType: mType,
	}

	if handler == constants.Get {
		err := validate(data, handler)
		if err != nil {
			return nil, fmt.Errorf("error while validation: %w", err)
		}

		return data, nil
	}

	switch mType {
	case domain.Counter:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad input value: %w", err)
		}

		data.Delta = &i
		return data, nil
	case domain.Gauge:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("bad input value - %w", err)
		}

		data.Value = &f
		return data, nil
	default:
		return nil, fmt.Errorf("wrong type")
	}
}

func validate(data *domain.MetricsJSON, handler string) error {
	if data.ID == "" {
		return fmt.Errorf("empty name")
	}
	if data.Value == nil && data.Delta == nil && handler == constants.Update {
		return fmt.Errorf("empty Value or Delta")
	}
	switch data.MType {
	case domain.Counter:
		return nil
	case domain.Gauge:
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}
