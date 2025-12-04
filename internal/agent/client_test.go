package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/stretchr/testify/require"
)

func TestSendUpdate_always_pass(t *testing.T) {
	testMName := "counter"
	testMap := map[string]*Metric{
		testMName: {
			MType: models.Counter,
			Value: "1",
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
				require.Equal(t, tCase.mm[tCase.mName].Value, value)
				require.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			tCase.cfg.URL = srv.URL

			ma := NewMetricAgent(tCase.cfg)
			mc := NewMetricClient(ma, tCase.cfg)
			mc.SendUpdate(tCase.mm)
		})
	}
}
