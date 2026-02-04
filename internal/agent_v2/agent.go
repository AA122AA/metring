package agentv2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/go-faster/sdk/zctx"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MetricAgent struct {
	pollInterval int
	lg           *zap.Logger
}

func NewMetricAgent(ctx context.Context, cfg *Config) *MetricAgent {
	return &MetricAgent{
		pollInterval: cfg.PollInterval,
		lg:           zctx.From(ctx).Named("metrics agent"),
	}
}

func (ma *MetricAgent) Run(ctx context.Context, wg *sync.WaitGroup) <-chan map[string]*Metric {
	// хочу хранить результаты за последние 5 минут
	toStore := 5 * 60 / ma.pollInterval
	results := make(chan map[string]*Metric, toStore)

	psResults := ma.GatherGopsutilMetrics(ctx, wg)
	runResults := ma.GatherRuntimeMetrics(ctx, wg)

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case res := <-psResults:
				results <- res
			case res := <-runResults:
				results <- res
			}
		}
	}(ctx)
	// go func() {
	// 	defer wg.Done()
	// 	for res := range psResults {
	// 		results <- res
	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for res := range runResults {
	// 		results <- res
	// 	}
	// }()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			_, psClosed := <-psResults
			_, runClosed := <-runResults
			if !psClosed && !runClosed {
				ma.lg.Info("got cancellation, returning")
				close(results)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return results
}

func (ma *MetricAgent) GatherGopsutilMetrics(ctx context.Context, wg *sync.WaitGroup) <-chan map[string]*Metric {
	var mu sync.Mutex
	out := make(chan map[string]*Metric)
	metricsMap := make(map[string]*Metric, 30)
	ticker := time.NewTicker(time.Duration(ma.pollInterval) * time.Second)

	wg.Add(1)
	go func() {
		defer ticker.Stop()
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				ma.lg.Info("got cancellation, returning")
				close(out)
				return
			case <-ticker.C:

				v, err := mem.VirtualMemory()
				if err != nil {
					ma.lg.Error("error while getting virtual memoty", zap.Error(err))
				}

				mu.Lock()
				free := float64(v.Free)
				metricsMap["FreeMemory"] = &Metric{
					ID:    "FreeMemory",
					MType: domain.Gauge,
					Value: &free,
				}

				total := float64(v.Total)
				metricsMap["TotalMemory"] = &Metric{
					ID:    "TotalMemory",
					MType: domain.Gauge,
					Value: &total,
				}

				c, err := cpu.Percent(1*time.Second, true)
				if err != nil {
					ma.lg.Error("error while getting cpu utilization", zap.Error(err))
				}
				for i, util := range c {
					name := fmt.Sprintf("CPUutilization%d", i)
					metricsMap[name] = &Metric{
						ID:    name,
						MType: domain.Gauge,
						Value: &util,
					}
				}
				mu.Unlock()

				out <- metricsMap
			}
		}
	}()
	return out
}

func (ma *MetricAgent) GatherRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup) <-chan map[string]*Metric {
	var mu sync.Mutex
	memoryStats := &runtime.MemStats{}
	out := make(chan map[string]*Metric)
	metricsMap := make(map[string]*Metric, 30)
	ticker := time.NewTicker(time.Duration(ma.pollInterval) * time.Second)

	wg.Add(1)
	go func() {
		defer ticker.Stop()
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				ma.lg.Info("got cancellation, returning")
				close(out)
				return
			case <-ticker.C:
				ma.lg.Debug("Start Gathering metrics")
				// Читаем метрики
				runtime.ReadMemStats(memoryStats)

				// Записываем в мапу
				v := reflect.ValueOf(*memoryStats)
				t := v.Type()

				mu.Lock()
				for i := range v.NumField() {
					m, err := createMetric(v, t, i)
					if err != nil {
						continue
					}

					metricsMap[m.ID] = m
				}

				d := int64(1)
				metricsMap["PollCount"] = &Metric{
					ID:    "PollCount",
					MType: domain.Counter,
					Delta: &d,
				}

				r := rand.Float64()
				metricsMap["RandomValue"] = &Metric{
					ID:    "RandomValue",
					MType: domain.Gauge,
					Value: &r,
				}
				mu.Unlock()

				out <- metricsMap
				ma.lg.Debug("Finish Gathering metrics")
			}
		}
	}()

	return out
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
