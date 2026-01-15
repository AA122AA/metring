package repository

import (
	"context"
	"fmt"

	"github.com/AA122AA/metring/internal/server/domain"
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

func (mr *mockRepo) GetAll(ctx context.Context) (map[string]*domain.Metrics, error) {
	i := int64(1)
	f := float64(1.25)
	return map[string]*domain.Metrics{
		Alloc: {
			ID:    "1",
			MType: domain.Gauge,
			Value: &f,
		},
		PollCount: {
			ID:    "2",
			MType: domain.Counter,
			Delta: &i,
		},
	}, nil
}

func (mr *mockRepo) Get(ctx context.Context, name string) (*domain.Metrics, error) {
	switch name {
	case PollCount:
		fmt.Printf("Got name - %v\n", name)
		v := int64(2)
		return &domain.Metrics{
			MType: domain.Counter,
			Delta: &v,
		}, nil
	case Alloc:
		fmt.Printf("Got name - %v\n", name)
		v := float64(1.25)
		return &domain.Metrics{
			MType: domain.Gauge,
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
		return &domain.Metrics{}, nil
	}
}

func (mr *mockRepo) Write(ctx context.Context, name string, value *domain.Metrics) error {
	return nil
}

func (mr *mockRepo) Update(ctx context.Context, value *domain.Metrics) error {
	return nil
}
