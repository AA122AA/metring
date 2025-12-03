package repository

import (
	"fmt"

	models "github.com/AA122AA/metring/internal/server/model"
)

type mockRepo struct{}

func NewMockRepo() *mockRepo {
	return &mockRepo{}
}

func (mr *mockRepo) GetAll() (map[string]*models.Metrics, error) {
	v := int64(1)
	return map[string]*models.Metrics{
		"Alloc": {
			ID:    "1",
			MType: models.Counter,
			Delta: &v,
		},
	}, nil
}

func (mr *mockRepo) Get(name string) (*models.Metrics, error) {
	switch name {
	case "counter":
		v := int64(2)
		return &models.Metrics{
			MType: models.Counter,
			Delta: &v,
		}, nil
	case "data":
		return nil, fmt.Errorf("data not found")
	case "error":
		return nil, fmt.Errorf("some error")
	case "gauge":
		v := float64(1.25)
		return &models.Metrics{
			MType: models.Gauge,
			Value: &v,
		}, nil
	default:
		return &models.Metrics{}, nil
	}
}

func (mr *mockRepo) Write(name string, value *models.Metrics) error {
	return nil
}
