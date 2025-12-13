package agent

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metric struct {
	MType string
	Value string
}

type MetricAgent struct {
	// TODO: добавить локи, чтоб безопасно работать с мапой
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
	// TODO: Возможно стоит переделать на рефлексию
	ma.mm["Alloc"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.Alloc),
	}
	ma.mm["BuckHashSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.BuckHashSys),
	}
	ma.mm["Frees"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.Frees),
	}
	ma.mm["GCCPUFraction"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.GCCPUFraction),
	}
	ma.mm["GCSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.GCSys),
	}
	ma.mm["HeapAlloc"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapAlloc),
	}
	ma.mm["HeapIdle"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapIdle),
	}
	ma.mm["HeapInuse"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapInuse),
	}
	ma.mm["HeapObjects"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapObjects),
	}
	ma.mm["HeapReleased"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapReleased),
	}
	ma.mm["HeapSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.HeapSys),
	}
	ma.mm["LastGC"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.LastGC),
	}
	ma.mm["Lookups"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.Lookups),
	}
	ma.mm["MCacheInuse"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.MCacheInuse),
	}
	ma.mm["MCacheSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.MCacheSys),
	}
	ma.mm["MSpanInuse"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.MSpanInuse),
	}
	ma.mm["MSpanSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.MSpanSys),
	}
	ma.mm["Mallocs"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.Mallocs),
	}
	ma.mm["NextGC"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.NextGC),
	}
	ma.mm["NumForcedGC"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.NumForcedGC),
	}
	ma.mm["NumGC"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.NumGC),
	}
	ma.mm["OtherSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.OtherSys),
	}
	ma.mm["PauseTotalNs"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.PauseTotalNs),
	}
	ma.mm["StackInuse"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.StackInuse),
	}
	ma.mm["StackSys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.StackSys),
	}
	ma.mm["Sys"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.Sys),
	}
	ma.mm["TotalAlloc"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", memoryStats.TotalAlloc),
	}
	ma.mm["PollCount"] = &Metric{
		MType: models.Counter,
		Value: fmt.Sprintf("%v", uint64(1)),
	}
	ma.mm["RandomValue"] = &Metric{
		MType: models.Gauge,
		Value: fmt.Sprintf("%v", rand.Float64()),
	}

	ma.lg.Debug("Finish Gathering metrics")
}

func (ma *MetricAgent) GetMetrics() map[string]*Metric {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	return ma.mm
}
