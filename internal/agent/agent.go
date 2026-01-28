package agent

import (
	"context"
	"fmt"
	"maps"
	"math/rand/v2"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MetricAgent struct {
	mm           map[string]*Metric
	mu           sync.Mutex
	pollInterval int
	lg           *zap.Logger
}

func NewMetricAgent(ctx context.Context, cfg *Config) *MetricAgent {
	return &MetricAgent{
		mm:           make(map[string]*Metric),
		pollInterval: cfg.PollInterval,
		lg:           zctx.From(ctx).Named("metrics agent"),
	}
}

func (ma *MetricAgent) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(ma.pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			ma.lg.Info("got cancellation, returning")
			return
		case <-ticker.C:
			ma.GatherMetrics()
		}
	}
}

func (ma *MetricAgent) GetMetrics() map[string]*Metric {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	return maps.Clone(ma.mm)
}

func (ma *MetricAgent) GatherMetrics() {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	ma.lg.Debug("Start Gathering metrics")

	// Читаем метрики
	memoryStats := &runtime.MemStats{}
	runtime.ReadMemStats(memoryStats)

	v := reflect.ValueOf(*memoryStats)
	t := v.Type()
	for i := range v.NumField() {
		m, err := createMetric(v, t, i)
		if err != nil {
			continue
		}

		ma.mm[m.ID] = m
	}

	d := int64(1)
	ma.mm["PollCount"] = &Metric{
		ID:    "PollCount",
		MType: domain.Counter,
		Delta: &d,
	}

	r := rand.Float64()
	ma.mm["RandomValue"] = &Metric{
		ID:    "RandomValue",
		MType: domain.Gauge,
		Value: &r,
	}

	ma.lg.Debug("Finish Gathering metrics")
}

func createMetric(v reflect.Value, t reflect.Type, i int) (*Metric, error) {
	value := v.Field(i)
	field := t.Field(i)
	m := &Metric{
		ID: field.Name,
	}

	var toValue any
	switch field.Type.Kind() {
	case reflect.Uint64, reflect.Uint32:
		if value.CanUint() {
			toValue = float64(value.Uint())
		} else {
			return nil, fmt.Errorf("field %v can not be Uint", field.Name)
		}
	case reflect.Float64:
		if value.CanFloat() {
			toValue = value.Float()
		} else {
			return nil, fmt.Errorf("field %v can not be Float", field.Name)
		}
	default:
		return nil, fmt.Errorf("do not work with type %v", field.Type.Name())
	}

	floatValue, ok := toValue.(float64)
	if !ok {
		return nil, fmt.Errorf("toValue is not float64")
	}

	m.Value = &floatValue
	m.MType = domain.Gauge

	return m, nil
}
