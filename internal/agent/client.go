package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type MetricClient struct {
	reportInterval int
	baseURL        string

	client *http.Client
	agent  *MetricAgent
	lg     *zap.Logger
}

func NewMetricClient(ctx context.Context, mAgent *MetricAgent, cfg *Config) *MetricClient {
	return &MetricClient{
		reportInterval: cfg.ReportInterval,
		baseURL:        cfg.URL,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		agent: mAgent,
		lg:    zctx.From(ctx).Named("metrics client"),
	}
}

func (mc *MetricClient) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(mc.reportInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			mc.lg.Info("got cancellation, returning")
			return
		case <-ticker.C:
			mc.SendUpdate(mc.agent.GetMetrics())
		}
	}
}

func (mc *MetricClient) SendUpdate(mm map[string]*Metric) {
	for k, v := range mm {

		u, err := buildURL(mc.baseURL, "update", v.MType, k, v.Value)
		if err != nil {
			mc.lg.Error("error building url", zap.Error(err))
			continue
		}

		req, err := http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			mc.lg.Error("error making new request", zap.Error(err))
			continue
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := mc.client.Do(req)
		if err != nil {
			mc.lg.Error("error doing request", zap.String("url", req.URL.String()), zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			mc.lg.Error("wrong status", zap.Int("status code", resp.StatusCode), zap.String("status", resp.Status))
			continue
		}
		mc.lg.Debug("sent update")
	}
}

func buildURL(base string, values ...string) (*url.URL, error) {
	if !strings.HasPrefix(base, "http://") {
		base = "http://" + base
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("error while building url: %w", err)
	}
	fullPath := path.Join(values...)
	u.Path = path.Join(u.Path, fullPath)
	return u, nil
}
