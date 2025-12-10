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

func (m *Metrics) Update(mName, mType, value string) error {
	// Создаем модель
	metric := &models.Metrics{
		MType: mType,
	}
	// Парсим значение метрики
	err := parse(metric, value)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Увеличиваем значение, если это Counter
	if metric.MType == models.Counter {
		v, err := m.repo.Get(mName)
		if err != nil {
			if err.Error() == "data not found" {
				return m.repo.Write(mName, metric)
			}
			return fmt.Errorf("%w", err)
		}

		*metric.Delta += *v.Delta
	}

	m.lg.Debug("updated value")

	return m.repo.Write(mName, metric)
}

func parse(metric *models.Metrics, value string) error {
	switch metric.MType {
	case models.Counter:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("bad input value - %w", err)
		}

		metric.Delta = &i
		return nil
	case models.Gauge:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("bad input value - %w", err)
		}

		metric.Value = &f
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}
