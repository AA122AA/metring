package service

import (
	"context"
	"fmt"
	"strconv"

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

func (m *Metrics) Get(mType, mName string) (string, error) {
	metric, err := m.repo.Get(mName)
	if err != nil {
		return "", fmt.Errorf("err from repo: %w", err)
	}
	if mType != metric.MType {
		return "", fmt.Errorf("wrong metric type")
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

func (m *Metrics) Update(data *models.MetricsJSON) error {
	if err := validate(data); err != nil {
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

func validate(data *models.MetricsJSON) error {
	if data.ID == "" {
		return fmt.Errorf("empty name")
	}
	if data.Value == nil && data.Delta == nil {
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

func (m *Metrics) Parse(mType, mName, value string) (*models.MetricsJSON, error) {
	data := &models.MetricsJSON{
		ID:    mName,
		MType: mType,
	}
	switch mType {
	case models.Counter:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad input value - %w", err)
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
