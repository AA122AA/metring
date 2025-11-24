package service

import (
	"fmt"
	"strconv"

	models "github.com/AA122AA/metring/internal/model"
	"github.com/AA122AA/metring/internal/repository"
)

type Metrics struct {
	repo repository.MetricsRepository
}

func NewMetrics(r repository.MetricsRepository) *Metrics {
	return &Metrics{
		repo: r,
	}
}

func (m *Metrics) Get(mName string) (*models.Metrics, error) {
	return m.repo.Get(mName)
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
