package repository

import models "github.com/AA122AA/metring/internal/server/model"

type MetricsRepository interface {
	Get(string) (*models.Metrics, error)
	Write(string, *models.Metrics) error
}
