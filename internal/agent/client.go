package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type MetricClient struct {
	reportInterval int
	baseURL        string
	key            string

	client         *http.Client
	agent          *MetricAgent
	lg             *zap.Logger
	maxRetry       int
	retryIntervals []int
}

func NewMetricClient(ctx context.Context, mAgent *MetricAgent, cfg *Config) *MetricClient {
	return &MetricClient{
		reportInterval: cfg.ReportInterval,
		baseURL:        cfg.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		agent:          mAgent,
		lg:             zctx.From(ctx).Named("metrics client"),
		maxRetry:       3,
		retryIntervals: []int{1, 3, 5},
		key:            cfg.Key,
	}
}

func (mc *MetricClient) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	timer := time.NewTimer(time.Duration(mc.reportInterval) * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			mc.lg.Info("got cancellation, returning")
			return
		case <-timer.C:
			mc.withRetry(ctx, mc.SendUpdateJSONBatch, mc.agent.GetMetrics())
			timer.Reset(time.Duration(mc.reportInterval) * time.Second)
		}
	}
}

func (mc *MetricClient) withRetry(ctx context.Context, f func(map[string]*Metric) error, mm map[string]*Metric) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for try := 0; try <= mc.maxRetry; try++ {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			err := f(mm)
			if err == nil {
				return
			}
			var re *ReqError
			if !errors.Is(err, re) {
				t := reflect.TypeOf(err)
				mc.lg.Debug("type of err", zap.Any("type", t))
				mc.lg.Error("error not in request", zap.Error(err))
				return
			}
			if try < len(mc.retryIntervals) {
				timer.Reset(time.Duration(mc.retryIntervals[try]) * time.Second)
			}
		}
	}

	timer.Reset(5 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			return
		}
	}
}

func (mc *MetricClient) SendUpdateJSONBatch(mm map[string]*Metric) error {
	metrics := make([]*Metric, 0, len(mm))
	for _, v := range mm {
		metrics = append(metrics, v)
	}

	u, err := buildURL(mc.baseURL, "updates/")
	if err != nil {
		mc.lg.Error("error building url", zap.Error(err))
		return err
	}
	mc.lg.Debug("url", zap.Any("url", u))

	body, err := json.Marshal(metrics)
	if err != nil {
		mc.lg.Error("error marshling body", zap.Error(err))
		return err
	}

	var hash string
	if mc.key != "" {
		hash = mc.createHash(body)
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(body)
	if err != nil {
		mc.lg.Error("error compressing body", zap.Error(err))
		return err
	}
	w.Close()

	resp, err := mc.makeRequest(u, &buf, hash)
	if err != nil {
		mc.lg.Error("error doing request", zap.String("url", u.String()), zap.Error(err))
		return NewReqError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		mc.lg.Error("wrong status", zap.Int("status code", resp.StatusCode), zap.String("status", resp.Status))
		return err
	}
	mc.lg.Debug("sent update successfully")

	return nil
}

func (mc *MetricClient) SendUpdateJSON(mm map[string]*Metric) error {
	var errs []error
	for _, v := range mm {
		u, err := buildURL(mc.baseURL, "update")
		if err != nil {
			mc.lg.Error("error building url", zap.Error(err))
			errs = append(errs, fmt.Errorf("error building url: %w", err))
			continue
		}

		body, err := json.Marshal(v)
		if err != nil {
			mc.lg.Error("error marshling body", zap.Error(err))
			errs = append(errs, fmt.Errorf("error marshling body: %w", err))
			continue
		}

		var hash string
		if mc.key != "" {
			hash = mc.createHash(body)
		}

		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, err = w.Write(body)
		if err != nil {
			mc.lg.Error("error compressing body", zap.Error(err))
			errs = append(errs, fmt.Errorf("error compressing body: %w", err))
			continue
		}
		w.Close()

		resp, err := mc.makeRequest(u, &buf, hash)
		if err != nil {
			mc.lg.Error("error doing request", zap.String("url", resp.Request.URL.String()), zap.Error(err))
			errs = append(errs, fmt.Errorf("error doing request: %w", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			mc.lg.Error("wrong status", zap.Int("status code", resp.StatusCode), zap.String("status", resp.Status))
			errs = append(errs, fmt.Errorf("wrong status code: %v, status: %v", resp.StatusCode, resp.Status))
			continue
		}
		mc.lg.Debug("sent update")
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (mc *MetricClient) SendUpdate(mm map[string]*Metric) error {
	var errs []error
	for k, v := range mm {
		var value string
		if v.Delta != nil {
			value = fmt.Sprintf("%v", v.Delta)
		}
		if v.Value != nil {
			value = fmt.Sprintf("%v", v.Value)
		}

		u, err := buildURL(mc.baseURL, "update", v.MType, k, value)
		if err != nil {
			mc.lg.Error("error building url", zap.Error(err))
			errs = append(errs, fmt.Errorf("error building url: %w", err))
			continue
		}

		req, err := http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			mc.lg.Error("error making new request", zap.Error(err))
			errs = append(errs, fmt.Errorf("error making new request: %w", err))
			continue
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := mc.client.Do(req)
		if err != nil {
			mc.lg.Error("error doing request", zap.String("url", req.URL.String()), zap.Error(err))
			errs = append(errs, fmt.Errorf("error doing request: %w", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			mc.lg.Error("wrong status", zap.Int("status code", resp.StatusCode), zap.String("status", resp.Status))
			errs = append(errs, fmt.Errorf("wrong status code: %v, status: %v", resp.StatusCode, resp.Status))
			continue
		}
		mc.lg.Debug("sent update")
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
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

func (mc *MetricClient) makeRequest(u *url.URL, buf *bytes.Buffer, hash string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, u.String(), buf)
	if err != nil {
		mc.lg.Error("error making new request", zap.Error(err))
		return nil, err
	}

	if mc.key != "" {
		req.Header.Set("HashSHA256", hash)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	return mc.client.Do(req)
}

func (mc *MetricClient) createHash(body []byte) string {
	h := hmac.New(sha256.New, []byte(mc.key))
	h.Write(body)
	hash := h.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
