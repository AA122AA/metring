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
		fmt.Printf("Got name - %v\n", name)
		v := int64(2)
		return &models.Metrics{
			MType: models.Counter,
			Delta: &v,
		}, nil
	case "data":
		fmt.Printf("Got name - %v\n", name)
		return nil, fmt.Errorf("data not found")
	case "error":
		fmt.Printf("Got name - %v\n", name)
		return nil, fmt.Errorf("some error")
	case "gauge":
		fmt.Printf("Got name - %v\n", name)
		v := float64(1.25)
		return &models.Metrics{
			MType: models.Gauge,
			Value: &v,
		}, nil
	default:
		fmt.Printf("Got name - %v\n", name)
		return &models.Metrics{}, nil
	}
}

func (mr *mockRepo) Write(name string, value *models.Metrics) error {
	return nil
}
