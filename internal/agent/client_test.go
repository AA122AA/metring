package agent

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/stretchr/testify/require"
)

func TestSendUpdate_always_pass(t *testing.T) {
	testMName := "counter"
	// v := float64(1.25)
	d := int64(1)
	ctx := context.Background()
	testMap := map[string]*Metric{
		testMName: {
			ID:    "counter",
			MType: models.Counter,
			// Value: "1",
			Delta: &d,
		},
	}
	cases := []struct {
		name  string
		mName string
		mm    map[string]*Metric
		cfg   *Config
	}{
		{
			name:  "Positive",
			mName: testMName,
			mm:    testMap,
			cfg: &Config{
				PollInterval:   2,
				URL:            "http://localhost:8080",
				ReportInterval: 10,
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("POST /update/{mType}/{mName}/{value}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mType := r.PathValue("mType")
				mName := r.PathValue("mName")
				value := r.PathValue("value")
				require.Equal(t, tCase.mm[tCase.mName].MType, mType)
				require.Equal(t, tCase.mName, mName)
				require.Equal(t, fmt.Sprintf("%v", tCase.mm[tCase.mName].Delta), value)
				require.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			tCase.cfg.URL = srv.URL

			ma := NewMetricAgent(ctx, tCase.cfg)
			mc := NewMetricClient(ctx, ma, tCase.cfg)
			mc.SendUpdate(tCase.mm)
		})
	}
}

func TestSendJSON(t *testing.T) {
	mux := http.NewServeMux()
	ctx := context.Background()

	mux.HandleFunc("POST /update", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		fmt.Printf("got metric - %v\n", metric)
		w.WriteHeader(http.StatusOK)
	}))

	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := &Config{
		PollInterval:   2,
		ReportInterval: 4,
		URL:            server.URL,
	}
	ma := NewMetricAgent(ctx, cfg)
	mc := NewMetricClient(ctx, ma, cfg)
	ma.GatherMetrics()
	mc.SendUpdateJSON(ma.mm)
}
