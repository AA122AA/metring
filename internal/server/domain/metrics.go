package domain

import "github.com/AA122AA/metring/internal/server/database/query"

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MetricsJSON struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func TransformFromJSON(data *MetricsJSON) *Metrics {
	return &Metrics{
		ID:    data.ID,
		MType: data.MType,
		Delta: data.Delta,
		Value: data.Value,
	}
}

func TransformToJSON(data *Metrics) *MetricsJSON {
	return &MetricsJSON{
		ID:    data.ID,
		MType: data.MType,
		Delta: data.Delta,
		Value: data.Value,
	}
}

func DBToDomain(metric *query.Metric) *Metrics {
	return &Metrics{
		ID:    metric.Name,
		MType: metric.Type,
		Delta: &metric.Delta.Int64,
		Value: &metric.Value.Float64,
		Hash:  metric.Hash.String,
	}
}
