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
)

type MetricClient struct {
	reportInterval int
	baseURL        string

	client *http.Client
	agent  *MetricAgent
}

func NewMetricClient(mAgent *MetricAgent, cfg *Config) *MetricClient {
	return &MetricClient{
		reportInterval: cfg.ReportInterval,
		baseURL:        cfg.URL,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		agent: mAgent,
	}
}

func (mc *MetricClient) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(mc.reportInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("got cancellation, returning")
			return
		case <-ticker.C:
			mc.SendUpdate(mc.agent.GetMetrics())
		}
	}
}

func (mc *MetricClient) SendUpdate(mm map[string]*Metric) {
	for k, v := range mm {

		u := buildURL(mc.baseURL, "update", v.MType, k, v.Value)

		req, err := http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			fmt.Printf("error making new request: %v\n", err)
			continue
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := mc.client.Do(req)
		if err != nil {
			fmt.Printf("error doing request to - %v, error: %v\n", req.URL.String(), err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("err - %v\n", resp.Status)
			continue
		}
		fmt.Println("sent update")
	}
}

func buildURL(base string, values ...string) *url.URL {
	if !strings.HasPrefix(base, "http://") {
		base = "http://" + base
	}

	u, err := url.Parse(base)
	if err != nil {
		fmt.Printf("error while building url: %v\n", err)
	}
	fullPath := path.Join(values...)
	u.Path = path.Join(u.Path, fullPath)
	return u
}
