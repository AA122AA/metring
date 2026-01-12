package metrics

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AA122AA/metring/internal/server/constants"
	models "github.com/AA122AA/metring/internal/server/model"
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

func (m *Metrics) GetAll() (map[string]*models.Metrics, error) {
	return m.repo.GetAll()
}

func (m *Metrics) get(data *models.MetricsJSON) (*models.Metrics, error) {
	if err := validate(data, constants.Get); err != nil {
		return nil, err
	}
	metric, err := m.repo.Get(data.ID)
	if err != nil {
		return nil, fmt.Errorf("err from repo: %w", err)
	}
	if data.MType != metric.MType {
		return nil, fmt.Errorf("metric %v, and data %v, has different types between data and repo: %v vs %v", metric.ID, data.ID, data.MType, metric.MType)
	}
	metric.ID = data.ID

	return metric, nil
}

func (m *Metrics) Get(data *models.MetricsJSON) (string, error) {
	metric, err := m.get(data)
	if err != nil {
		return "", err
	}

	var res string
	switch metric.MType {
	case models.Gauge:
		res = fmt.Sprintf("%g", *metric.Value)
	case models.Counter:
		res = fmt.Sprintf("%d", *metric.Delta)
	}

	return res, nil
}

func (m *Metrics) GetJSON(data *models.MetricsJSON) (*models.MetricsJSON, error) {
	metric, err := m.get(data)
	if err != nil {
		return nil, err
	}

	return models.TransformToJSON(metric), nil
}

func (m *Metrics) Update(data *models.MetricsJSON) error {
	if err := validate(data, constants.Update); err != nil {
		return err
	}
	metric := models.TransformFromJSON(data)

	// Увеличиваем значение, если это Counter
	if data.MType == models.Counter {
		v, err := m.repo.Get(metric.ID)
		if err != nil {
			if err.Error() == "data not found" {
				return m.repo.Write(data.ID, metric)
			}
			return fmt.Errorf("%w", err)
		}

		*metric.Delta += *v.Delta
	}

	m.lg.Debug("updated value")

	return m.repo.Write(metric.ID, metric)
}

func validate(data *models.MetricsJSON, handler string) error {
	if data.ID == "" {
		return fmt.Errorf("empty name")
	}
	if data.Value == nil && data.Delta == nil && handler == constants.Update {
		return fmt.Errorf("empty Value or Delta")
	}
	switch data.MType {
	case models.Counter:
		return nil
	case models.Gauge:
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}

func (m *Metrics) Parse(mType, mName, value, handler string) (*models.MetricsJSON, error) {
	data := &models.MetricsJSON{
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
	case models.Counter:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad input value: %w", err)
		}

		data.Delta = &i
		return data, nil
	case models.Gauge:
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
