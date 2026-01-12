package agent

import (
	"context"
	"maps"
	"math/rand/v2"
	"reflect"
	"runtime"
	"sync"
	"time"

	models "github.com/AA122AA/metring/internal/server/model"
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
	for {
		select {
		case <-ctx.Done():
			ma.lg.Info("got cancellation, returning")
			return
		default:
			ma.GatherMetrics()
			time.Sleep(time.Duration(ma.pollInterval) * time.Second)
		}
	}
}

func (ma *MetricAgent) GatherMetrics() {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	ma.lg.Debug("Start Gathering metrics")

	memoryStats := &runtime.MemStats{}
	runtime.ReadMemStats(memoryStats)
	v := reflect.ValueOf(*memoryStats)
	t := v.Type()
	for i := range v.NumField() {
		value := v.Field(i)
		field := t.Field(i)
		m := &Metric{
			ID: field.Name,
		}
		var toValue interface{}
		switch field.Type.Kind() {
		case reflect.Uint64, reflect.Uint32:
			if value.CanUint() {
				toValue = float64(value.Uint())
			} else {
				continue
			}
		case reflect.Float64:
			if value.CanFloat() {
				toValue = value.Float()
			} else {
				continue
			}
		default:
			ma.lg.Debug("do not work with this type", zap.String("type", field.Type.Name()))
			continue
		}

		floatValue, ok := toValue.(float64)
		if !ok {
			ma.lg.Error("wrong assertion")
			continue
		}

		m.Value = &floatValue
		m.MType = models.Gauge
		ma.mm[field.Name] = m
	}

	d := int64(1)
	ma.mm["PollCount"] = &Metric{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &d,
	}

	r := rand.Float64()
	ma.mm["RandomValue"] = &Metric{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: &r,
	}

	ma.lg.Debug("Finish Gathering metrics")
}

func (ma *MetricAgent) GetMetrics() map[string]*Metric {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	return maps.Clone(ma.mm)
}
