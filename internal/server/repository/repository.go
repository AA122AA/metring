package repository

import models "github.com/AA122AA/metring/internal/server/model"

type MetricsRepository interface {
	GetAll() (map[string]*models.Metrics, error)
	Get(name string) (*models.Metrics, error)
	Write(name string, value *models.Metrics) error
}
