package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/stretchr/testify/require"
)

func TestSendJSON(t *testing.T) {
	var tCase struct {
		mName string
		mm    map[string]*Metric
		cfg   *Config
	}

	t.Run("Positive", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			ReportInterval: 10,
		}

		// Test body
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			metric := &Metric{}
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}

			err := json.NewDecoder(r.Body).Decode(metric)
			require.NoError(t, err)
			require.Equal(t, tCase.mm[tCase.mName], metric)
			w.WriteHeader(http.StatusOK)
		})
		mux := http.NewServeMux()
		mux.HandleFunc("POST /update", handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		tCase.cfg.URL = server.URL

		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdateJSON(tCase.mm)
		require.NoError(t, err)
	})

	t.Run("Negative buildURL", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			URL:            ":&*%^((*U!!",
			ReportInterval: 10,
		}

		// Test body
		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdateJSON(tCase.mm)
		require.ErrorContains(t, err, "error building url")
	})

	t.Run("Negative", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			ReportInterval: 10,
		}

		// Test body
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			metric := &Metric{}
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}

			err := json.NewDecoder(r.Body).Decode(metric)
			require.NoError(t, err)
			require.Equal(t, tCase.mm[tCase.mName], metric)
			w.WriteHeader(http.StatusOK)
		})
		mux := http.NewServeMux()
		mux.HandleFunc("POST /update", handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		tCase.cfg.URL = server.URL

		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdateJSON(tCase.mm)
		require.NoError(t, err)
	})
}

func TestSendUpdate(t *testing.T) {
	var tCase struct {
		mName string
		mm    map[string]*Metric
		cfg   *Config
	}

	t.Run("Positive", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			URL:            "http://localhost:8080",
			ReportInterval: 10,
		}

		// Test body
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mType := r.PathValue("mType")
			mName := r.PathValue("mName")
			value := r.PathValue("value")

			require.Equal(t, tCase.mm[tCase.mName].MType, mType)
			require.Equal(t, tCase.mName, mName)
			require.Equal(t, fmt.Sprintf("%v", tCase.mm[tCase.mName].Delta), value)
			require.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
		})

		mux := http.NewServeMux()
		mux.HandleFunc("POST /update/{mType}/{mName}/{value}", handler)
		srv := httptest.NewServer(mux)
		defer srv.Close()

		tCase.cfg.URL = srv.URL

		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdate(tCase.mm)
		require.NoError(t, err)
	})

	t.Run("Negative buildURL", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			URL:            ":&*%^((*U!!",
			ReportInterval: 10,
		}

		// Test body

		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdate(tCase.mm)
		require.ErrorContains(t, err, "error building url")
	})

	t.Run("Negative Do", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			URL:            "http://localhost:8080",
			ReportInterval: 10,
		}

		// Test body
		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdate(tCase.mm)
		require.ErrorContains(t, err, "error doing request")
	})

	t.Run("Negative status", func(t *testing.T) {
		// Prepare
		d := int64(1)
		ctx := context.Background()
		tCase.mName = "counter"
		tCase.mm = map[string]*Metric{
			"counter": {
				ID:    "counter",
				MType: domain.Counter,
				Delta: &d,
			},
		}
		tCase.cfg = &Config{
			PollInterval:   2,
			URL:            "http://localhost:8080",
			ReportInterval: 10,
		}

		// Test body
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		mux := http.NewServeMux()
		mux.HandleFunc("POST /update/{mType}/{mName}/{value}", handler)
		srv := httptest.NewServer(mux)
		defer srv.Close()

		tCase.cfg.URL = srv.URL

		ma := NewMetricAgent(ctx, tCase.cfg)
		mc := NewMetricClient(ctx, ma, tCase.cfg)
		err := mc.SendUpdate(tCase.mm)
		require.ErrorContains(t, err, "wrong status")
	})
}

func TestBuildURL(t *testing.T) {
	cases := []struct {
		name   string
		base   string
		values []string
		want   string
		pass   bool
	}{
		{
			name:   "Positive",
			base:   "localhost",
			values: []string{"value", "gauge"},
			want:   "http://localhost/value/gauge",
			pass:   true,
		},
		{
			name:   "Negative",
			base:   ":&*%^((*U!!",
			values: []string{"value", "gauge"},
			want:   "http://localhost/value/gauge",
			pass:   false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			u, err := buildURL(tCase.base, tCase.values...)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.want, u.String())
				return
			}
			require.ErrorContains(t, err, "error while building url")
		})
	}
}

func TestMakeRequest(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		PollInterval:   1,
		ReportInterval: 2,
	}
	cases := []struct {
		name      string
		url       *url.URL
		body      *bytes.Buffer
		want      string
		pass      bool
		badHost   bool
		badScheme bool
	}{
		{
			name: "Positive post metric",
			url: &url.URL{
				Scheme: "http",
				Path:   "/value",
			},
			body:      bytes.NewBuffer([]byte(`{"id":"PollCount", "type":"counter", "delta":1}`)),
			want:      `{"id":"PollCount", "type":"counter", "delta":1}`,
			pass:      true,
			badHost:   false,
			badScheme: false,
		},
		{
			name: "Negative, no host",
			url: &url.URL{
				Scheme: "http",
				Path:   "/value",
			},
			body:      bytes.NewBuffer([]byte(`{"id":"PollCount", "type":"counter"}`)),
			pass:      false,
			badHost:   true,
			badScheme: false,
		},
		{
			name: "Negative, bad scheme",
			url: &url.URL{
				Scheme: "ht",
				Path:   "/value",
			},
			body:      bytes.NewBuffer([]byte(`{"id":"PollCount", "type":"counter"}`)),
			pass:      false,
			badHost:   false,
			badScheme: true,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			// create mock server
			mux := http.NewServeMux()
			mux.HandleFunc(fmt.Sprintf("POST %v", tCase.url.Path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				defer r.Body.Close()

				if tCase.pass {
					require.Equal(t, tCase.want, string(body))
				}

				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			if !tCase.badHost {
				tCase.url.Host = srv.Listener.Addr().String()
			}

			cfg.URL = srv.URL
			ma := NewMetricAgent(ctx, cfg)
			mc := NewMetricClient(ctx, ma, cfg)

			resp, err := mc.makeRequest(tCase.url, tCase.body, "")

			if tCase.pass {
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
				return
			}
			if tCase.badHost {
				fmt.Printf("%+v\n", err)
				require.ErrorContains(t, err, "no Host in request URL")
				return
			}
			if tCase.badScheme {
				require.ErrorContains(t, err, "unsupported protocol")
				return
			}
		})
	}
}
