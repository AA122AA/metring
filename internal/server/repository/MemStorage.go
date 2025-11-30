package repository

import (
	"fmt"

	models "github.com/AA122AA/metring/internal/server/model"
)

type MemStorage struct {
	Values map[string]*models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Values: make(map[string]*models.Metrics),
	}
}

func (ms *MemStorage) Get(name string) (*models.Metrics, error) {
	if v, ok := ms.Values[name]; ok {
		return v, nil
	}

	return nil, fmt.Errorf("data not found")
}

func (ms *MemStorage) Write(name string, value *models.Metrics) error {
	ms.Values[name] = value
	return nil
}
