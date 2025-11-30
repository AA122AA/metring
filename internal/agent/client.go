package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type MetricClient struct {
	reportInterval int
	baseUrl        string

	client *http.Client
	agent  *MetricAgent
}

func NewMetricClient(mAgent *MetricAgent, cfg *Config) *MetricClient {
	return &MetricClient{
		reportInterval: cfg.ReportInterval,
		baseUrl:        cfg.Url,
		client: &http.Client{
			Timeout: 3 * time.Second,
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
		u, err := url.JoinPath(mc.baseUrl, "update", v.MType, k, v.Value)
		if err != nil {
			fmt.Printf("error creating url, got - %v\n", u)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, u, nil)
		if err != nil {
			fmt.Printf("error making new request\n")
			continue
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := mc.client.Do(req)
		if err != nil {
			fmt.Printf("error doing request to - %v\n", req.URL.String())
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("err - %v\n", resp.Status)
			continue
		}
		fmt.Println("sent update")
	}
}
