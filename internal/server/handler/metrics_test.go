package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service/metrics"
	"github.com/AA122AA/metring/internal/server/service/saver"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := saver.Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
		Restore:         true,
	}

	ctx := context.Background()
	cases := []struct {
		name   string
		url    string
		tPath  string
		status int
		repo   repository.MetricsRepository
		pass   bool
	}{
		{
			name: "Positive",
			url:  "/",
			// tPath:  "/home/artem/Documents/developm/metricsent/Yandex.Practicum/metring/internal/server/templates/*.html",
			tPath:  "../templates/*.html",
			status: 200,
			repo:   repository.NewMockRepo(),
			pass:   true,
		},
		{
			name:  "Negative tPath",
			url:   "/",
			tPath: "xxx/*.html",
			// tPath:  "../templates/*.html",
			status: 500,
			repo:   repository.NewMockRepo(),
			pass:   false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			srv := metrics.NewMetrics(ctx, tCase.repo)
			saver := saver.NewSaver(ctx, cfg, tCase.repo)
			h := NewMetricsHandler(ctx, tCase.tPath, srv, saver)

			r := httptest.NewRequest(http.MethodGet, tCase.url, nil)

			rec := httptest.NewRecorder()
			h.All(rec, r)

			res := rec.Result()

			if tCase.pass {
				require.Equal(t, tCase.status, res.StatusCode)

				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				require.Contains(t, string(body), repository.Alloc)
				require.Contains(t, res.Header.Get("Content-Type"), "text/html")
				return
			}
			require.Equal(t, tCase.status, res.StatusCode)
		})
	}
}

func TestGet(t *testing.T) {
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := saver.Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
		Restore:         true,
	}

	ctx := context.Background()
	cases := []struct {
		name   string
		mName  string
		mType  string
		url    string
		tPath  string
		status int
		pass   bool
		want   string
	}{
		{
			name:   "Positive",
			mName:  repository.Alloc,
			mType:  domain.Gauge,
			url:    "/value/gauge/gauge",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusOK,
			pass:   true,
			want:   "1.25",
		},
		{
			name:   "Negative, wrong type",
			mName:  repository.Alloc,
			mType:  domain.Counter,
			url:    "/value/data/Alloc",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusNotFound,
			pass:   false,
			want:   `No metric with this type`,
		},
		{
			name:   "Negative, wrong name",
			mName:  repository.NoData,
			mType:  domain.Gauge,
			url:    "/value/counter/data",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusNotFound,
			pass:   false,
			want:   `No metric with this name`,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := metrics.NewMetrics(ctx, repo)
			saver := saver.NewSaver(ctx, cfg, repo)
			h := NewMetricsHandler(ctx, tCase.tPath, srv, saver)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("mName", tCase.mName)
			rctx.URLParams.Add("mType", tCase.mType)

			r := httptest.NewRequest(http.MethodGet, tCase.url, nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()
			h.Get(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {
				require.Equal(t, res.StatusCode, http.StatusOK)
				return
			}
			require.Equal(t, tCase.status, res.StatusCode)
		})
	}
}

func TestGetJSON(t *testing.T) {
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := saver.Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
		Restore:         true,
	}
	ctx := context.Background()
	cases := []struct {
		name   string
		metric *domain.MetricsJSON
		url    string
		tPath  string
		status int
		pass   bool
		want   string
	}{
		{
			name: "Positive",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: domain.Gauge,
			},
			url:    "/value",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusOK,
			pass:   true,
			want:   `{"id":"alloc","type":"gauge","value":1.25}`,
		},
		{
			name: "Negative, wrong type",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: domain.Counter,
			},
			url:    "/value",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusNotFound,
			pass:   false,
			want:   `No metric with this type`,
		},
		{
			name: "Negative, wrong name",
			metric: &domain.MetricsJSON{
				ID:    repository.NoData,
				MType: domain.Gauge,
			},
			url:    "/value",
			tPath:  "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			status: http.StatusNotFound,
			pass:   false,
			want:   `No metric with this name`,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := metrics.NewMetrics(ctx, repo)
			saver := saver.NewSaver(ctx, cfg, repo)
			h := NewMetricsHandler(ctx, tCase.tPath, srv, saver)

			body, err := json.Marshal(tCase.metric)
			require.NoError(t, err)
			buf := bytes.NewBuffer(body)

			r := httptest.NewRequest(http.MethodPost, tCase.url, buf)

			rec := httptest.NewRecorder()
			h.GetJSON(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {
				require.Equal(t, res.StatusCode, http.StatusOK)
				return
			}
			require.Equal(t, tCase.status, res.StatusCode)
		})
	}
}

func TestUpdate(t *testing.T) {
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := saver.Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
		Restore:         true,
	}
	ctx := context.Background()
	cases := []struct {
		name       string
		url        string
		tPath      string
		mName      string
		mType      string
		value      string
		want       string
		statusCode int
		pass       bool
	}{
		{
			name:       "Positive",
			url:        "/update/gauge/gauge/1.25",
			tPath:      "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			mName:      repository.Alloc,
			mType:      domain.Gauge,
			value:      "1.25",
			statusCode: http.StatusOK,
			pass:       true,
		},
		{
			name:       "Negative, no mName",
			url:        "/update//gauge/1.25",
			tPath:      "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			mName:      "",
			mType:      "gauge",
			value:      "1.25",
			statusCode: http.StatusNotFound,
			pass:       false,
		},
		{
			name:       "Negative, bad type",
			url:        "/update/gauge/gauge/1.25",
			tPath:      "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			mName:      "gauge",
			mType:      "lol",
			value:      "1.25",
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
		{
			name:       "Negative, bad value",
			url:        "/update/gauge/gauge/1.25",
			tPath:      "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			mName:      "gauge",
			mType:      "gauge",
			value:      "a",
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := metrics.NewMetrics(ctx, repo)
			saver := saver.NewSaver(ctx, cfg, repo)
			h := NewMetricsHandler(ctx, tCase.tPath, srv, saver)

			r := httptest.NewRequest(http.MethodPost, tCase.url, nil)
			r.SetPathValue("mName", tCase.mName)
			r.SetPathValue("mType", tCase.mType)
			r.SetPathValue("value", tCase.value)
			rec := httptest.NewRecorder()
			h.Update(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {
				require.Equal(t, tCase.statusCode, res.StatusCode)
				return
			}

			require.Equal(t, tCase.statusCode, res.StatusCode)
		})
	}
}

func TestUpdateJSON(t *testing.T) {
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := saver.Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
		Restore:         true,
	}
	v := float64(1.25)
	ctx := context.Background()
	cases := []struct {
		name       string
		url        string
		tPath      string
		metric     *domain.MetricsJSON
		want       string
		statusCode int
		pass       bool
	}{
		{
			name:  "Positive",
			url:   "/update",
			tPath: "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: domain.Gauge,
				Value: &v,
			},
			statusCode: http.StatusOK,
			pass:       true,
		},
		{
			name:  "Negative, no mName",
			url:   "/update",
			tPath: "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			metric: &domain.MetricsJSON{
				ID:    "",
				MType: domain.Gauge,
				Value: &v,
			},
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
		{
			name:  "Negative, bad type",
			url:   "/update",
			tPath: "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: "lol",
				Value: &v,
			},
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
		{
			name:  "Negative, bad value",
			url:   "/update/",
			tPath: "/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: domain.Gauge,
			},
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := metrics.NewMetrics(ctx, repo)
			saver := saver.NewSaver(ctx, cfg, repo)
			h := NewMetricsHandler(ctx, tCase.tPath, srv, saver)

			body, err := json.Marshal(tCase.metric)
			require.NoError(t, err)
			buf := bytes.NewBuffer(body)

			r := httptest.NewRequest(http.MethodPost, tCase.url, buf)
			rec := httptest.NewRecorder()
			h.UpdateJSON(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {

				require.Equal(t, tCase.statusCode, res.StatusCode)
				return
			}

			require.Equal(t, tCase.statusCode, res.StatusCode)
		})
	}
}
