package repository

import (
	"fmt"

	models "github.com/AA122AA/metring/internal/server/model"
)

const (
	Alloc     = "alloc"
	Counter   = "counter"
	PollCount = "pollCount"
	NoData    = "noData"
	Error     = "error"
)

type mockRepo struct{}

func NewMockRepo() *mockRepo {
	return &mockRepo{}
}

func (mr *mockRepo) GetAll() (map[string]*models.Metrics, error) {
	i := int64(1)
	f := float64(1.25)
	return map[string]*models.Metrics{
		Alloc: {
			ID:    "1",
			MType: models.Gauge,
			Value: &f,
		},
		PollCount: {
			ID:    "2",
			MType: models.Counter,
			Delta: &i,
		},
	}, nil
}

func (mr *mockRepo) Get(name string) (*models.Metrics, error) {
	switch name {
	case PollCount:
		fmt.Printf("Got name - %v\n", name)
		v := int64(2)
		return &models.Metrics{
			MType: models.Counter,
			Delta: &v,
		}, nil
	case Alloc:
		fmt.Printf("Got name - %v\n", name)
		v := float64(1.25)
		return &models.Metrics{
			MType: models.Gauge,
			Value: &v,
		}, nil
	case NoData:
		fmt.Printf("Got name - %v\n", name)
		return nil, fmt.Errorf("data not found")
	case Error:
		fmt.Printf("Got name - %v\n", name)
		return nil, fmt.Errorf("some error")

	default:
		fmt.Printf("Got name - %v\n", name)
		return &models.Metrics{}, nil
	}
}

func (mr *mockRepo) Write(name string, value *models.Metrics) error {
	return nil
}
